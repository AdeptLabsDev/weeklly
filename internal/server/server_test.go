package server

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"strings"
	"testing"
	"time"

	_ "time/tzdata"

	"github.com/AdeptLabsDev/weeklly/internal/auth"
	"github.com/AdeptLabsDev/weeklly/internal/config"
	"github.com/AdeptLabsDev/weeklly/internal/store"
	"github.com/AdeptLabsDev/weeklly/web"
)

const testClientID = "cliente-de-teste"

// app sobe o servidor completo atrás de TLS (os cookies são Secure) com um
// cliente que guarda cookies e não segue redirecionamentos, para inspecioná-los.
type app struct {
	t      *testing.T
	server *Server
	http   *httptest.Server
	client *http.Client
	google *auth.FakeProvider
}

func newApp(t *testing.T, withGoogle bool) *app {
	t.Helper()
	ctx := context.Background()
	st, err := store.Open(ctx, filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = st.Close() })
	if err := st.Migrate(ctx); err != nil {
		t.Fatal(err)
	}
	cfg, err := config.Load(func(string) string { return "" })
	if err != nil {
		t.Fatal(err)
	}

	a := &app{t: t}
	opts := Options{
		Config:  cfg,
		Logger:  slog.New(slog.NewTextHandler(io.Discard, nil)),
		Store:   st,
		Web:     web.FS,
		Version: "test",
	}
	if withGoogle {
		a.google = auth.NewFakeProvider(testClientID)
		t.Cleanup(a.google.Close)
		opts.Google = auth.NewGoogle(testClientID, "segredo", "https://weeklly.test/entrar/google/callback", a.google.Endpoints(), a.google.Server.Client())
	}
	a.server, err = New(opts)
	if err != nil {
		t.Fatal(err)
	}
	// "Agora" é quarta-feira, 9 de setembro de 2026, 15h UTC (12h em São Paulo).
	a.server.now = func() time.Time { return time.Date(2026, time.September, 9, 15, 0, 0, 0, time.UTC) }

	a.http = httptest.NewTLSServer(a.server)
	t.Cleanup(a.http.Close)
	a.client = newFreshClient(t, a.http)
	return a
}

type reply struct {
	status   int
	header   http.Header
	body     string
	location string
}

func (a *app) get(path string) reply {
	a.t.Helper()
	resp, err := a.client.Get(a.http.URL + path)
	if err != nil {
		a.t.Fatal(err)
	}
	return read(a.t, resp)
}

func (a *app) post(path string, form url.Values) reply {
	a.t.Helper()
	req, err := http.NewRequest(http.MethodPost, a.http.URL+path, strings.NewReader(form.Encode()))
	if err != nil {
		a.t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Sec-Fetch-Site", "same-origin")
	resp, err := a.client.Do(req)
	if err != nil {
		a.t.Fatal(err)
	}
	return read(a.t, resp)
}

func read(t *testing.T, resp *http.Response) reply {
	t.Helper()
	defer func() { _ = resp.Body.Close() }()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	return reply{status: resp.StatusCode, header: resp.Header, body: string(body), location: resp.Header.Get("Location")}
}

func (a *app) createWeek(name string) string {
	a.t.Helper()
	r := a.post("/semanas", url.Values{"name": {name}})
	if r.status != http.StatusSeeOther || !strings.HasPrefix(r.location, "/semana/") {
		a.t.Fatalf("criar semana: status %d, Location %q, corpo:\n%s", r.status, r.location, r.body)
	}
	return r.location
}

func (a *app) hasSessionCookie() bool {
	u, _ := url.Parse(a.http.URL)
	for _, c := range a.client.Jar.Cookies(u) {
		if c.Name == sessionCookie && c.Value != "" {
			return true
		}
	}
	return false
}

func TestFreshVisitorSeesEmptyHubAndLogin(t *testing.T) {
	a := newApp(t, false)
	r := a.get("/")
	if r.status != http.StatusOK {
		t.Fatalf("status = %d", r.status)
	}
	for _, want := range []string{"Suas semanas", "Você ainda não tem uma semana", "Criar a primeira semana", "Entrar com Google", `href="/semanas/nova"`, "logo-mark"} {
		if !strings.Contains(r.body, want) {
			t.Errorf("hub sem %q", want)
		}
	}
	if strings.Contains(r.body, "nav-week") || strings.Contains(r.body, `class="avatar"`) {
		t.Error("visitante sem semana não deveria ver seletor de semana nem avatar")
	}
	if a.hasSessionCookie() {
		t.Error("visitar o hub não pode criar sessão")
	}
	if ct := r.header.Get("Cache-Control"); ct != "private, no-cache" {
		t.Errorf("Cache-Control = %q", ct)
	}
}

func TestCreateWeekOpensTheBoardAndStartsASession(t *testing.T) {
	a := newApp(t, false)

	r := a.post("/semanas", url.Values{"name": {"   "}})
	if r.status != http.StatusUnprocessableEntity || !strings.Contains(r.body, "Dê um nome para a semana.") {
		t.Errorf("nome vazio: status %d", r.status)
	}
	if a.hasSessionCookie() {
		t.Error("formulário inválido não pode criar sessão")
	}

	location := a.createWeek("  Semana   padrão ")
	if !a.hasSessionCookie() {
		t.Fatal("criar a primeira semana deveria abrir uma sessão")
	}

	board := a.get(location)
	if board.status != http.StatusOK {
		t.Fatalf("quadro: status %d", board.status)
	}
	for _, want := range []string{
		`class="nav-week"`, "Semana padrão", "Hoje é quarta-feira", `id="hoje"`, `href="#hoje"`,
		"Nada aqui ainda", `class="pager-dot`, `id="weeks-menu"`, "editada ", `id="new-week"`,
		`id="rename-week"`, `id="delete-week"`, `class="task-add"`, `data-theme="dark"`, "/duplicar",
	} {
		if !strings.Contains(board.body, want) {
			t.Errorf("quadro sem %q", want)
		}
	}
	if n := strings.Count(board.body, `<li class="card`); n != 7 {
		t.Errorf("%d cartões, quero 7", n)
	}
	if n := strings.Count(board.body, `class="card is-today"`); n != 1 {
		t.Errorf("%d cartões de hoje, quero 1", n)
	}
	if strings.Contains(board.body, "<script>") || strings.Contains(board.body, `style="`) {
		t.Error("script ou estilo inline: a CSP bloqueia")
	}

	// Abrir o app volta para a semana.
	home := a.get("/")
	if home.status != http.StatusSeeOther || home.location != location {
		t.Errorf("home: status %d, Location %q, quero %q", home.status, home.location, location)
	}

	// O hub lista a semana; a segunda semana entra no seletor e vira a atual.
	second := a.createWeek("Semana de provas")
	hub := a.get("/semanas")
	if hub.status != http.StatusOK || !strings.Contains(hub.body, "Semana de provas") || !strings.Contains(hub.body, "Semana padrão") {
		t.Errorf("hub: status %d", hub.status)
	}
	if home := a.get("/"); home.location != second {
		t.Errorf("home deveria abrir a última semana usada: %q", home.location)
	}
	board2 := a.get(second)
	if !strings.Contains(board2.body, `aria-current="page"`) || strings.Count(board2.body, `class="menu-item`) < 2 {
		t.Error("seletor de semanas incompleto")
	}
}

func TestWeeksAreInvisibleToOthers(t *testing.T) {
	owner := newApp(t, false)
	location := owner.createWeek("Minha semana")

	// Outro navegador (sem cookie) no mesmo servidor.
	stranger := newFreshClient(t, owner.http)
	resp, err := stranger.Get(owner.http.URL + location)
	if err != nil {
		t.Fatal(err)
	}
	r := read(t, resp)
	if r.status != http.StatusNotFound || !strings.Contains(r.body, "Isso não está aqui") {
		t.Errorf("semana alheia: status %d", r.status)
	}

	for _, bad := range []string{"/semana/curto", "/semana/" + strings.Repeat("a", 16), "/nao-existe", "/semanas/x"} {
		if r := owner.get(bad); r.status != http.StatusNotFound {
			t.Errorf("%s: status %d, quero 404", bad, r.status)
		}
	}
}

func TestLoginUnavailableWithoutCredentials(t *testing.T) {
	a := newApp(t, false)
	r := a.get("/entrar/google")
	if r.status != http.StatusServiceUnavailable || !strings.Contains(r.body, "Login ainda não configurado") {
		t.Errorf("status %d", r.status)
	}
	if r := a.get(callbackPath + "?code=x&state=y"); r.status != http.StatusNotFound {
		t.Errorf("callback sem login: status %d", r.status)
	}
}

func TestGoogleLoginClaimsAnonymousWeeksAndShowsAvatar(t *testing.T) {
	a := newApp(t, true)
	location := a.createWeek("Semana do visitante")

	start := a.get("/entrar/google")
	if start.status != http.StatusFound || !strings.HasPrefix(start.location, a.google.Server.URL+"/auth?") {
		t.Fatalf("início do login: status %d, Location %q", start.status, start.location)
	}
	state := url.Values{}
	if u, err := url.Parse(start.location); err == nil {
		state = u.Query()
	}

	// State errado: nada acontece.
	if r := a.get(callbackPath + "?code=abc&state=errado"); r.status != http.StatusBadRequest {
		t.Errorf("state errado: status %d", r.status)
	}

	// O cookie do desafio foi apagado pelo callback errado; começa de novo.
	start = a.get("/entrar/google")
	if u, err := url.Parse(start.location); err == nil {
		state = u.Query()
	}
	code := a.google.Authorize(auth.FakeUser{Sub: "g-miguel", Email: "miguel@example.com", EmailVerified: true, Name: "Miguel", Picture: "https://lh3.googleusercontent.com/foto"}, state.Get("nonce"))
	done := a.get(callbackPath + "?code=" + url.QueryEscape(code) + "&state=" + url.QueryEscape(state.Get("state")))
	if done.status != http.StatusSeeOther || done.location != "/" {
		t.Fatalf("callback: status %d, Location %q, corpo:\n%s", done.status, done.location, done.body)
	}

	// A semana anônima continua acessível e a barra mostra a foto.
	board := a.get(location)
	if board.status != http.StatusOK {
		t.Fatalf("quadro depois do login: %d", board.status)
	}
	if !strings.Contains(board.body, `class="avatar"`) || !strings.Contains(board.body, "lh3.googleusercontent.com/foto") || strings.Contains(board.body, "Entrar com Google") {
		t.Error("barra deveria mostrar só a foto de perfil")
	}
	if !strings.Contains(board.body, "miguel@example.com") || !strings.Contains(board.body, `action="/sair"`) {
		t.Error("menu da conta incompleto")
	}

	// Sair encerra a sessão: o quadro some e o hub volta vazio.
	out := a.post("/sair", nil)
	if out.status != http.StatusSeeOther || a.hasSessionCookie() {
		t.Errorf("sair: status %d, cookie ainda presente: %v", out.status, a.hasSessionCookie())
	}
	if r := a.get(location); r.status != http.StatusNotFound {
		t.Errorf("depois de sair a semana ainda abre: %d", r.status)
	}

	// Entrar de novo, em um navegador limpo, recupera a semana pela conta.
	fresh := newFreshClient(t, a.http)
	a.client = fresh
	start = a.get("/entrar/google")
	if u, err := url.Parse(start.location); err == nil {
		state = u.Query()
	}
	code = a.google.Authorize(auth.FakeUser{Sub: "g-miguel", Email: "miguel@example.com", EmailVerified: true, Name: "Miguel"}, state.Get("nonce"))
	if r := a.get(callbackPath + "?code=" + url.QueryEscape(code) + "&state=" + url.QueryEscape(state.Get("state"))); r.status != http.StatusSeeOther {
		t.Fatalf("segundo login: %d\n%s", r.status, r.body)
	}
	if r := a.get(location); r.status != http.StatusOK || !strings.Contains(r.body, "Semana do visitante") {
		t.Errorf("semana não voltou com a conta: %d", r.status)
	}
	if r := a.get("/"); r.location != location {
		t.Errorf("home depois do login: %q", r.location)
	}
}

func TestGoogleLoginRejectsUnverifiedEmailAndCancel(t *testing.T) {
	a := newApp(t, true)
	start := a.get("/entrar/google")
	q := url.Values{}
	if u, err := url.Parse(start.location); err == nil {
		q = u.Query()
	}
	code := a.google.Authorize(auth.FakeUser{Sub: "g-x", Email: "x@example.com", EmailVerified: false}, q.Get("nonce"))
	if r := a.get(callbackPath + "?code=" + url.QueryEscape(code) + "&state=" + url.QueryEscape(q.Get("state"))); r.status != http.StatusForbidden {
		t.Errorf("e-mail sem confirmação: status %d", r.status)
	}
	if a.hasSessionCookie() {
		t.Error("login recusado não pode abrir sessão")
	}

	a.get("/entrar/google")
	if r := a.get(callbackPath + "?error=access_denied"); r.status != http.StatusOK || !strings.Contains(r.body, "Login cancelado") {
		t.Errorf("cancelado: status %d", r.status)
	}
}

// newFreshClient é um navegador novo: mesmo transporte (confia no certificado
// de teste), pote de cookies próprio. srv.Client() devolve sempre o mesmo
// objeto, por isso não serve para simular duas pessoas.
func newFreshClient(t *testing.T, srv *httptest.Server) *http.Client {
	t.Helper()
	jar, _ := cookiejar.New(nil)
	return &http.Client{
		Transport:     srv.Client().Transport,
		Jar:           jar,
		CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
	}
}

func TestHealthz(t *testing.T) {
	a := newApp(t, false)
	r := a.get("/healthz")
	if r.status != http.StatusOK {
		t.Fatalf("status = %d", r.status)
	}
	var body map[string]string
	if err := json.Unmarshal([]byte(r.body), &body); err != nil {
		t.Fatal(err)
	}
	if body["status"] != "ok" || body["version"] != "test" {
		t.Errorf("body = %v", body)
	}
	if r.header.Get("Cache-Control") != "no-store" {
		t.Error("healthz precisa de Cache-Control: no-store")
	}
}

func TestSecurityHeadersOnEveryResponse(t *testing.T) {
	a := newApp(t, false)
	for _, target := range []string{"/", "/healthz", "/nao-existe", "/static/favicon.svg", "/semanas/nova"} {
		h := a.get(target).header
		if !strings.HasPrefix(h.Get("Content-Security-Policy"), "default-src 'none'") {
			t.Errorf("%s: CSP = %q", target, h.Get("Content-Security-Policy"))
		}
		if !strings.Contains(h.Get("Content-Security-Policy"), "img-src 'self' data: https://lh3.googleusercontent.com") {
			t.Errorf("%s: CSP sem a foto do Google", target)
		}
		if h.Get("X-Content-Type-Options") != "nosniff" || h.Get("X-Frame-Options") != "DENY" {
			t.Errorf("%s: headers de segurança ausentes", target)
		}
		if h.Get("Strict-Transport-Security") == "" {
			t.Errorf("%s: HSTS ausente em produção", target)
		}
	}
}

func TestStaticAssets(t *testing.T) {
	a := newApp(t, false)

	r := a.get("/static/favicon.svg")
	if r.status != http.StatusOK || !strings.HasPrefix(r.header.Get("Content-Type"), "image/svg+xml") {
		t.Errorf("favicon: status %d, type %q", r.status, r.header.Get("Content-Type"))
	}
	if !strings.Contains(r.header.Get("Cache-Control"), "max-age=3600") {
		t.Errorf("sem hash: Cache-Control = %q", r.header.Get("Cache-Control"))
	}
	r = a.get(a.server.assets.URL("favicon.svg"))
	if !strings.Contains(r.header.Get("Cache-Control"), "immutable") {
		t.Errorf("com hash: Cache-Control = %q", r.header.Get("Cache-Control"))
	}
	r = a.get("/static/app.js")
	if r.status != http.StatusOK || !strings.HasPrefix(r.header.Get("Content-Type"), "text/javascript") {
		t.Errorf("app.js: status %d, type %q", r.status, r.header.Get("Content-Type"))
	}
	r = a.get("/static/fonts/nunito.woff2")
	if r.status != http.StatusOK || r.header.Get("Content-Type") != "font/woff2" {
		t.Errorf("fonte: status %d, type %q", r.status, r.header.Get("Content-Type"))
	}
	for _, bad := range []string{"/static/", "/static/fonts", "/static/nope.css", "/static/..%2fgo.mod"} {
		r := a.get(bad)
		if r.status == http.StatusOK {
			t.Errorf("%s: status 200", bad)
		}
		if strings.Contains(r.body, "module github.com") {
			t.Errorf("%s: vazou arquivo fora de static", bad)
		}
	}
}

func TestRecovererTurnsPanicInto500(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	h := recoverer(logger)(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		panic("boom")
	}))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Code != http.StatusInternalServerError {
		t.Errorf("status = %d", rec.Code)
	}
}

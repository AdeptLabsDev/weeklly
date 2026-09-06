package server

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"encoding/xml"
	"html"
	"image/png"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"testing"

	"github.com/AdeptLabsDev/weeklly/internal/config"
	"github.com/AdeptLabsDev/weeklly/internal/i18n"
)

const seoOrigin = "https://weeklly.example"

// Cada página pública, com a tradução irmã (mesmo grupo de hreflang).
var seoPages = []struct {
	path, lang, opposite, image, sibling string
}{
	{"/planejador-semanal", "pt-BR", "en", "/static/og-planner-pt.png", "/en/weekly-planner"},
	{"/en/weekly-planner", "en", "pt-BR", "/static/og-planner-en.png", "/planejador-semanal"},
	{"/perguntas-frequentes", "pt-BR", "en", "/static/og-planner-pt.png", "/en/faq"},
	{"/en/faq", "en", "pt-BR", "/static/og-planner-en.png", "/perguntas-frequentes"},
}

func TestPublicPlannerMetadataAndStableLanguage(t *testing.T) {
	for _, page := range seoPages {
		t.Run(page.lang, func(t *testing.T) {
			a := newApp(t, false)
			a.server.opts.Config.SearchIndexing = true
			a.server.opts.Config.BaseURL = seoOrigin
			req, err := http.NewRequest(http.MethodGet, a.http.URL+page.path+"?utm_source=regression", nil)
			if err != nil {
				t.Fatal(err)
			}
			// A origem da requisição e as preferências não decidem a URL pública
			// nem o idioma de uma página cuja tradução tem endereço próprio.
			req.Host = "untrusted.example"
			req.Header.Set("Accept-Language", page.opposite)
			req.AddCookie(&http.Cookie{Name: langCookie, Value: page.opposite})
			r := seoRequest(t, a, req)
			if r.status != http.StatusOK {
				t.Fatalf("landing: status %d", r.status)
			}
			if got := seoTags(r.body, "html"); len(got) != 1 || got[0]["lang"] != page.lang {
				t.Errorf("idioma da URL %s não foi preservado: %v", page.path, got)
			}
			seoCheckDirectives(t, r.header.Get("X-Robots-Tag"), "index", "follow", "max-image-preview:large")
			seoCheckDirectives(t, seoMeta(t, r.body, "name", "robots"), "index", "follow", "max-image-preview:large")
			if r.header.Get("Set-Cookie") != "" || a.hasSessionCookie() {
				t.Error("visitar uma página pública não deve criar cookies ou sessão")
			}
			if strings.Contains(r.body, "untrusted.example") || strings.Contains(r.body, "utm_source=regression") {
				t.Error("host ou query da requisição vazaram para o documento público")
			}

			canonical := seoOrigin + page.path
			links := seoTags(r.body, "link")
			canonicals := 0
			alternates := map[string]string{}
			for _, link := range links {
				switch link["rel"] {
				case "canonical":
					canonicals++
					if link["href"] != canonical {
						t.Errorf("canonical = %q, quero %q", link["href"], canonical)
					}
				case "alternate":
					if lang := link["hreflang"]; lang != "" {
						if _, exists := alternates[lang]; exists {
							t.Errorf("hreflang duplicado: %s", lang)
						}
						alternates[lang] = link["href"]
					}
				}
			}
			if canonicals != 1 {
				t.Errorf("%d canonicals, quero um", canonicals)
			}
			ptPath, enPath := page.path, page.sibling
			if page.lang == "en" {
				ptPath, enPath = page.sibling, page.path
			}
			for lang, target := range map[string]string{
				"pt-BR":     seoOrigin + ptPath,
				"en":        seoOrigin + enPath,
				"x-default": seoOrigin + ptPath,
			} {
				if alternates[lang] != target {
					t.Errorf("hreflang %s = %q, quero %q", lang, alternates[lang], target)
				}
			}
			other := page.sibling
			linkedTranslation := false
			for _, anchor := range seoTags(r.body, "a") {
				if anchor["href"] == other || anchor["href"] == seoOrigin+other {
					linkedTranslation = true
				}
			}
			if !linkedTranslation {
				t.Error("a outra tradução precisa de um link HTML rastreável")
			}

			description := seoMeta(t, r.body, "name", "description")
			if strings.TrimSpace(description) == "" {
				t.Error("description vazia")
			}
			if got := seoMeta(t, r.body, "property", "og:url"); got != canonical {
				t.Errorf("og:url = %q", got)
			}
			for _, key := range []string{"og:title", "og:type"} {
				if strings.TrimSpace(seoMeta(t, r.body, "property", key)) == "" {
					t.Errorf("%s vazio", key)
				}
			}
			if got := seoMeta(t, r.body, "property", "og:description"); got != description {
				t.Errorf("og:description diverge da descrição pública: %q", got)
			}
			if got := seoMeta(t, r.body, "name", "twitter:card"); got != "summary_large_image" {
				t.Errorf("twitter:card = %q", got)
			}
			if strings.TrimSpace(seoMeta(t, r.body, "name", "twitter:title")) == "" {
				t.Error("twitter:title vazio")
			}
			if got := seoMeta(t, r.body, "name", "twitter:description"); got != description {
				t.Errorf("twitter:description diverge da descrição pública: %q", got)
			}
			imageURL := seoMeta(t, r.body, "property", "og:image")
			if got := seoMeta(t, r.body, "name", "twitter:image"); got != imageURL {
				t.Errorf("twitter:image = %q, og:image = %q", got, imageURL)
			}
			u, err := url.Parse(imageURL)
			if err != nil || u.Scheme != "https" || u.Host != "weeklly.example" || u.Path != page.image || u.Query().Get("v") == "" {
				t.Fatalf("imagem social precisa da origem pública e hash do asset: %q (%v)", imageURL, err)
			}
			asset := a.get(u.RequestURI())
			if asset.status != http.StatusOK || !strings.HasPrefix(asset.header.Get("Content-Type"), "image/png") {
				t.Fatalf("imagem social: status %d, type %q", asset.status, asset.header.Get("Content-Type"))
			}
			if _, err := png.DecodeConfig(bytes.NewReader([]byte(asset.body))); err != nil {
				t.Errorf("imagem social não é PNG válido: %v", err)
			}
			seoCheckSchema(t, r, canonical)
		})
	}
}

// O tema é a única preferência que a página pública segue: muda o
// data-theme e o botão sol/lua, nunca o conteúdo nem o idioma.
func TestPublicPlannerFollowsTheme(t *testing.T) {
	a := newApp(t, false)
	for _, page := range seoPages {
		req, err := http.NewRequest(http.MethodGet, a.http.URL+page.path, nil)
		if err != nil {
			t.Fatal(err)
		}
		req.AddCookie(&http.Cookie{Name: themeCookie, Value: themeLight})
		r := seoRequest(t, a, req)
		if r.status != http.StatusOK || !strings.Contains(r.body, `data-theme="light"`) {
			t.Errorf("%s: tema claro não aplicado (status %d)", page.path, r.status)
		}
		if got := seoTags(r.body, "html"); len(got) != 1 || got[0]["lang"] != page.lang {
			t.Errorf("%s: o tema não pode mudar o idioma: %v", page.path, got)
		}
		if !strings.Contains(r.body, `class="theme"`) || !strings.Contains(r.body, `popovertarget="language-menu"`) {
			t.Errorf("%s: barra sem tema e idioma", page.path)
		}
		if r.header.Get("Set-Cookie") != "" {
			t.Errorf("%s: landing modificou cookies", page.path)
		}
	}
}

func TestPublicPlannerIgnoresPrivateSessionAndPreferences(t *testing.T) {
	a := newApp(t, false)
	a.server.opts.Config.SearchIndexing = true
	a.server.opts.Config.BaseURL = seoOrigin
	fresh := make(map[string]string, len(seoPages))
	for _, page := range seoPages {
		fresh[page.path] = a.get(page.path).body
	}
	privateName := "Plano privado exclusivo do teste SEO"
	location := a.createWeek(privateName)
	for _, page := range seoPages {
		req, err := http.NewRequest(http.MethodGet, a.http.URL+page.path, nil)
		if err != nil {
			t.Fatal(err)
		}
		req.AddCookie(&http.Cookie{Name: cursorCookie, Value: cursorCustom})
		req.AddCookie(&http.Cookie{Name: accentCookie, Value: "pink"})
		req.AddCookie(&http.Cookie{Name: langCookie, Value: page.opposite})
		r := seoRequest(t, a, req)
		if r.status != http.StatusOK || r.body != fresh[page.path] {
			t.Errorf("%s: a sessão e as preferências mudaram a landing pública", page.path)
		}
		if strings.Contains(r.body, privateName) || strings.Contains(r.body, location) {
			t.Errorf("%s: dados da semana privada vazaram", page.path)
		}
		if r.header.Get("Set-Cookie") != "" {
			t.Errorf("%s: landing modificou cookies", page.path)
		}
		// O único script executável é o do próprio produto (tema com
		// varredura); o conteúdo não depende dele.
		for _, script := range seoTags(r.body, "script") {
			if script["type"] == "application/ld+json" {
				continue
			}
			if !strings.HasPrefix(script["src"], "/static/app.js") || script["type"] != "module" {
				t.Errorf("%s: landing carregou script inesperado: %v", page.path, script)
			}
		}
	}
	if home := a.get("/"); home.status != http.StatusSeeOther || home.location != location {
		t.Errorf("a raiz deixou de abrir a última semana: status %d, Location %q", home.status, home.location)
	}

	// Cookie inválido também é irrelevante: as rotas públicas não passam pela
	// resolução de sessão, que apagaria o cookie desconhecido na resposta.
	client := newFreshClient(t, a.http)
	for _, target := range append(publicPaths(), "/robots.txt", "/sitemap.xml") {
		req, err := http.NewRequest(http.MethodGet, a.http.URL+target, nil)
		if err != nil {
			t.Fatal(err)
		}
		req.AddCookie(&http.Cookie{Name: sessionCookie, Value: "invalid-session"})
		resp, err := client.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		r := read(t, resp)
		if r.status != http.StatusOK || r.header.Get("Set-Cookie") != "" {
			t.Errorf("%s dependeu do cookie inválido: status %d, Set-Cookie %q", target, r.status, r.header.Get("Set-Cookie"))
		}
	}
}

// A landing apresenta; as perguntas têm página própria, ligada pelo rodapé.
func TestPublicPagesStructure(t *testing.T) {
	a := newApp(t, false)
	a.server.opts.Config.SearchIndexing = true
	a.server.opts.Config.BaseURL = seoOrigin
	for _, lang := range i18n.All {
		l := i18n.L(lang)
		landing := a.get(landingPaths[lang])
		if landing.status != http.StatusOK {
			t.Fatalf("%s: landing status %d", lang, landing.status)
		}
		if strings.Contains(landing.body, `id="faq"`) || strings.Contains(landing.body, "<details") {
			t.Errorf("%s: a landing ainda traz as perguntas", lang)
		}
		if strings.Contains(landing.body, `href="#how-it-works"`) {
			t.Errorf("%s: a barra ainda tem o atalho de como funciona", lang)
		}
		if n := strings.Count(landing.body, `class="how-radio"`); n != 3 {
			t.Errorf("%s: %d passos, quero 3", lang, n)
		}
		if n := strings.Count(landing.body, `class="how-day"`); n != 7 {
			t.Errorf("%s: %d dias na cena, quero 7", lang, n)
		}
		if !strings.Contains(landing.body, ">"+l.WeekdayShort(0)+"</text>") {
			t.Errorf("%s: a cena não traz os dias no idioma da página", lang)
		}
		if !strings.Contains(landing.body, `href="`+faqPaths[lang]+`"`) {
			t.Errorf("%s: o rodapé não liga às perguntas em %s", lang, faqPaths[lang])
		}
		// A faixa dos aplicativos vem depois de "como funciona" e antes dos
		// pontos fortes; sem URL da loja, o selo não é link.
		how, apps, features := strings.Index(landing.body, `id="how-it-works"`), strings.Index(landing.body, `id="apps"`), strings.Index(landing.body, `id="features"`)
		if how >= apps || apps >= features || !strings.Contains(landing.body, l.T("landing.apps.heading")) {
			t.Errorf("%s: faixa dos aplicativos fora do lugar (%d, %d, %d)", lang, how, apps, features)
		}
		if strings.Contains(landing.body, `<a class="store-badge"`) || !strings.Contains(landing.body, `<span class="store-badge is-soon">`) {
			t.Errorf("%s: selo da loja virou link sem URL da loja", lang)
		}

		faq := a.get(faqPaths[lang])
		if faq.status != http.StatusOK {
			t.Fatalf("%s: faq status %d", lang, faq.status)
		}
		if got := seoTags(faq.body, "html"); len(got) != 1 || got[0]["lang"] != string(lang) {
			t.Errorf("%s: idioma da página de perguntas: %v", lang, got)
		}
		if n := strings.Count(faq.body, "<details"); n != len(faqTopics) {
			t.Errorf("%s: %d perguntas na página, quero %d", lang, n, len(faqTopics))
		}
		if strings.Count(faq.body, "<h1") != 1 || !strings.Contains(faq.body, l.T("faq.heading")) {
			t.Errorf("%s: página de perguntas sem H1 próprio", lang)
		}
		for _, topic := range faqTopics {
			if !strings.Contains(faq.body, html.EscapeString(l.T("faq."+topic+".question"))) {
				t.Errorf("%s: pergunta %s ausente", lang, topic)
			}
		}
		if !strings.Contains(faq.body, `href="/semanas/nova?lang=`+string(lang)+`"`) {
			t.Errorf("%s: página de perguntas sem chamada para criar semana", lang)
		}
		if !strings.Contains(faq.body, `"FAQPage"`) || strings.Count(faq.body, `"@type":"Question"`) != len(faqTopics) {
			t.Errorf("%s: JSON-LD sem as %d perguntas", lang, len(faqTopics))
		}
	}
}

func TestSEOPrivateRoutesRemainNoindex(t *testing.T) {
	a := newApp(t, false)
	a.server.opts.Config.SearchIndexing = true
	a.server.opts.Config.BaseURL = seoOrigin
	location := a.createWeek("Semana privada para SEO")
	for _, target := range []string{"/", "/semanas", "/semanas/nova", location, location + "/renomear", location + "/excluir", "/nao-existe", "/healthz", "/entrar/google", callbackPath} {
		r := a.get(target)
		seoCheckDirectives(t, r.header.Get("X-Robots-Tag"), "noindex", "nofollow")
		for _, link := range seoTags(r.body, "link") {
			if link["rel"] == "canonical" || link["hreflang"] != "" {
				t.Errorf("%s: página privada ou de erro recebeu metadados indexáveis: %v", target, link)
			}
		}
	}
}

func TestSEORobotsAndSitemapOnlyPublishPublicTranslations(t *testing.T) {
	a := newApp(t, false)
	a.server.opts.Config.SearchIndexing = true
	a.server.opts.Config.BaseURL = seoOrigin
	privateLocation := a.createWeek("Semana que não aparece no sitemap")
	robots := a.get("/robots.txt")
	if robots.status != http.StatusOK || !strings.HasPrefix(robots.header.Get("Content-Type"), "text/plain") {
		t.Fatalf("robots: status %d, type %q", robots.status, robots.header.Get("Content-Type"))
	}
	directives := seoRobotsLines(robots.body)
	if directives["user-agent"] != "*" || directives["disallow"] != "" || directives["sitemap"] != seoOrigin+"/sitemap.xml" {
		t.Errorf("robots público incorreto: %q", robots.body)
	}
	if _, exists := directives["disallow"]; !exists {
		t.Error("robots público não declarou Disallow vazio")
	}
	sitemap := a.get("/sitemap.xml?ignored=yes")
	urls := seoSitemapURLs(t, sitemap)
	if len(urls) != len(seoPages) {
		t.Errorf("sitemap contém %d URLs, quero %d", len(urls), len(seoPages))
	}
	seen := make(map[string]bool, len(urls))
	for _, target := range urls {
		if seen[target] {
			t.Errorf("URL duplicada no sitemap: %q", target)
		}
		seen[target] = true
	}
	for _, page := range seoPages {
		if !seen[seoOrigin+page.path] {
			t.Errorf("tradução ausente no sitemap: %s", page.path)
		}
	}
	if strings.Contains(sitemap.body, privateLocation) || strings.Contains(sitemap.body, "/semanas") || strings.Contains(sitemap.body, "ignored=yes") {
		t.Error("sitemap incluiu rota privada ou query da requisição")
	}
}

func TestSEODisabledWithoutProductionOptInAndHTTPS(t *testing.T) {
	for _, tc := range []struct {
		name    string
		env     config.Env
		base    string
		enabled bool
	}{
		{"default-off", config.Production, seoOrigin, false},
		{"development", config.Development, seoOrigin, true},
		{"no-origin", config.Production, "", true},
		{"http-origin", config.Production, "http://weeklly.example", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			a := newApp(t, false)
			a.server.opts.Config.Env = tc.env
			a.server.opts.Config.BaseURL = tc.base
			a.server.opts.Config.SearchIndexing = tc.enabled
			for _, page := range seoPages {
				r := a.get(page.path)
				if r.status != http.StatusOK {
					t.Errorf("%s: status %d", page.path, r.status)
				}
				seoCheckDirectives(t, r.header.Get("X-Robots-Tag"), "noindex", "nofollow")
				seoCheckDirectives(t, seoMeta(t, r.body, "name", "robots"), "noindex", "nofollow")
				if tc.base == "" {
					for _, link := range seoTags(r.body, "link") {
						if link["rel"] == "canonical" || link["hreflang"] != "" {
							t.Errorf("sem BaseURL, não se deve inventar URL pública: %v", link)
						}
					}
					if strings.Contains(r.body, a.http.URL) {
						t.Error("a origem da requisição substituiu BaseURL ausente")
					}
				}
			}
			robots := a.get("/robots.txt")
			directives := seoRobotsLines(robots.body)
			if robots.status != http.StatusOK || directives["user-agent"] != "*" || directives["disallow"] != "/" {
				t.Errorf("robots deveria bloquear rastreamento: status %d, corpo %q", robots.status, robots.body)
			}
			if urls := seoSitemapURLs(t, a.get("/sitemap.xml")); len(urls) != 0 {
				t.Errorf("sitemap desabilitado publicou URLs: %v", urls)
			}
		})
	}
}

func seoRequest(t *testing.T, a *app, req *http.Request) reply {
	t.Helper()
	resp, err := a.client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	return read(t, resp)
}

// Os testes leem atributos pelo nome para tolerar ordenação, aspas e
// espaços sem depender da formatação do template.
func seoTags(body, tag string) []map[string]string {
	pattern := regexp.MustCompile(`(?is)<` + tag + `\b([^>]*)>`)
	var tags []map[string]string
	for _, match := range pattern.FindAllStringSubmatch(body, -1) {
		tags = append(tags, seoAttributes(match[1]))
	}
	return tags
}

func seoAttributes(raw string) map[string]string {
	pattern := regexp.MustCompile(`([A-Za-z_:][A-Za-z0-9_:.-]*)\s*=\s*(?:"([^"]*)"|'([^']*)')`)
	attrs := map[string]string{}
	for _, match := range pattern.FindAllStringSubmatch(raw, -1) {
		value := match[2]
		if match[3] != "" {
			value = match[3]
		}
		attrs[strings.ToLower(match[1])] = html.UnescapeString(value)
	}
	return attrs
}

func seoMeta(t *testing.T, body, attribute, key string) string {
	t.Helper()
	var values []string
	for _, meta := range seoTags(body, "meta") {
		if meta[attribute] == key {
			values = append(values, meta["content"])
		}
	}
	if len(values) != 1 {
		t.Errorf("meta %s=%q: encontrei %d, quero um", attribute, key, len(values))
		return ""
	}
	return values[0]
}

func seoCheckDirectives(t *testing.T, raw string, want ...string) {
	t.Helper()
	got := map[string]bool{}
	for _, part := range strings.Split(strings.ToLower(raw), ",") {
		got[strings.TrimSpace(part)] = true
	}
	for _, directive := range want {
		if !got[directive] {
			t.Errorf("diretiva %q ausente em %q", directive, raw)
		}
	}
	if got["index"] && got["noindex"] || got["follow"] && got["nofollow"] {
		t.Errorf("diretivas contraditórias: %q", raw)
	}
}

func seoCheckSchema(t *testing.T, r reply, canonical string) {
	t.Helper()
	pattern := regexp.MustCompile(`(?is)<script\b([^>]*)>(.*?)</script\s*>`)
	var scripts []string
	for _, match := range pattern.FindAllStringSubmatch(r.body, -1) {
		if seoAttributes(match[1])["type"] == "application/ld+json" {
			scripts = append(scripts, match[2])
		}
	}
	if len(scripts) != 1 {
		t.Fatalf("JSON-LD: encontrei %d scripts, quero um grafo", len(scripts))
	}
	raw := scripts[0]
	var schema struct {
		Context string           `json:"@context"`
		Graph   []map[string]any `json:"@graph"`
	}
	if err := json.Unmarshal([]byte(raw), &schema); err != nil {
		t.Fatalf("JSON-LD não é objeto JSON válido: %v", err)
	}
	if schema.Context != "https://schema.org" {
		t.Errorf("@context = %q", schema.Context)
	}
	found := map[string]bool{}
	for _, node := range schema.Graph {
		// @type é uma string ou uma lista ("WebPage" e "FAQPage" juntos).
		var kinds []string
		switch v := node["@type"].(type) {
		case string:
			kinds = []string{v}
		case []any:
			for _, k := range v {
				kinds = append(kinds, k.(string))
			}
		}
		for _, kind := range kinds {
			found[kind] = true
			if kind == "WebPage" && node["url"] != canonical {
				t.Errorf("WebPage.url = %v, quero %q", node["url"], canonical)
			}
		}
	}
	if !found["WebPage"] || !found["WebApplication"] {
		t.Errorf("grafo incompleto: tipos %v", found)
	}
	if strings.Contains(canonical, "/faq") || strings.Contains(canonical, "/perguntas-frequentes") {
		if !found["FAQPage"] {
			t.Errorf("página de perguntas sem FAQPage: tipos %v", found)
		}
	} else if found["FAQPage"] {
		t.Error("FAQPage fora da página de perguntas")
	}
	// Nenhuma avaliação ou oferta pode ser inventada para habilitar um
	// resultado enriquecido que o produto ainda não sustenta.
	var complete any
	if err := json.Unmarshal([]byte(raw), &complete); err != nil {
		t.Fatal(err)
	}
	seoCheckFactualSchema(t, complete)
	sum := sha256.Sum256([]byte(raw))
	hash := "'sha256-" + base64.StdEncoding.EncodeToString(sum[:]) + "'"
	csp := r.header.Get("Content-Security-Policy")
	scriptSource := ""
	for _, directive := range strings.Split(csp, ";") {
		fields := strings.Fields(directive)
		if len(fields) > 0 && fields[0] == "script-src" {
			scriptSource = directive
		}
	}
	if !strings.Contains(scriptSource, hash) {
		t.Errorf("CSP não autoriza o conteúdo exato do JSON-LD: quero %s em %q", hash, scriptSource)
	}
	if strings.Contains(csp, "'unsafe-inline'") || strings.Contains(csp, "'unsafe-eval'") {
		t.Error("dados estruturados enfraqueceram a CSP")
	}
}

func seoCheckFactualSchema(t *testing.T, value any) {
	t.Helper()
	switch value := value.(type) {
	case map[string]any:
		for key, child := range value {
			switch key {
			case "aggregateRating", "review", "reviews", "ratingValue", "offers", "price":
				t.Errorf("schema contém afirmação comercial não comprovada: %s", key)
			}
			seoCheckFactualSchema(t, child)
		}
	case []any:
		for _, child := range value {
			seoCheckFactualSchema(t, child)
		}
	}
}

func seoRobotsLines(body string) map[string]string {
	directives := map[string]string{}
	for _, line := range strings.Split(body, "\n") {
		key, value, ok := strings.Cut(line, ":")
		if ok {
			directives[strings.ToLower(strings.TrimSpace(key))] = strings.TrimSpace(value)
		}
	}
	return directives
}

func seoSitemapURLs(t *testing.T, r reply) []string {
	t.Helper()
	if r.status != http.StatusOK || !strings.Contains(r.header.Get("Content-Type"), "xml") {
		t.Fatalf("sitemap: status %d, type %q", r.status, r.header.Get("Content-Type"))
	}
	var sitemap struct {
		XMLName xml.Name `xml:"urlset"`
		URLs    []struct {
			Location string `xml:"loc"`
		} `xml:"url"`
	}
	if err := xml.Unmarshal([]byte(r.body), &sitemap); err != nil {
		t.Fatalf("sitemap XML inválido: %v", err)
	}
	if sitemap.XMLName.Space != "http://www.sitemaps.org/schemas/sitemap/0.9" {
		t.Errorf("namespace do sitemap = %q", sitemap.XMLName.Space)
	}
	urls := make([]string, 0, len(sitemap.URLs))
	for _, entry := range sitemap.URLs {
		urls = append(urls, entry.Location)
	}
	return urls
}

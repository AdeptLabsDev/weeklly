package server

import (
	"net/http"
	"strings"
	"time"

	"github.com/AdeptLabsDev/weeklly/internal/i18n"
)

// Tema e idioma são cookies legíveis pelo JavaScript: o servidor entrega a
// página já certa (sem piscar) e o script troca sem recarregar quando dá.
const (
	themeCookie  = "weeklly_theme"
	themeDark    = "dark"
	themeLight   = "light"
	langCookie   = "weeklly_lang"
	cursorCookie = "weeklly_cursor"
	cursorSystem = "system"
	cursorCustom = "custom"
	accentCookie = "weeklly_accent"
	accentMono   = "mono"
	prefTTL      = 365 * 24 * time.Hour
)

// Cores de destaque, na ordem do menu. "mono" é o padrão: sem cor (D22).
var accents = []string{accentMono, "orange", "amber", "green", "blue", "violet", "pink", "red"}

func themeFrom(r *http.Request) string {
	if c, err := r.Cookie(themeCookie); err == nil && c.Value == themeLight {
		return themeLight
	}
	return themeDark
}

// cursorFrom lê o tipo de cursor: o do sistema, salvo escolha explícita.
func cursorFrom(r *http.Request) string {
	if c, err := r.Cookie(cursorCookie); err == nil && c.Value == cursorCustom {
		return cursorCustom
	}
	return cursorSystem
}

// accentFrom lê a cor de destaque; valor desconhecido vira "mono".
func accentFrom(r *http.Request) string {
	if c, err := r.Cookie(accentCookie); err == nil && validAccent(c.Value) {
		return c.Value
	}
	return accentMono
}

func validAccent(v string) bool {
	for _, a := range accents {
		if a == v {
			return true
		}
	}
	return false
}

// locale escolhe o idioma da requisição: cookie, senão Accept-Language,
// senão português.
func (s *Server) locale(r *http.Request) i18n.Locale {
	if lang, ok := entryLanguage(r); ok {
		return i18n.L(lang)
	}
	cookie := ""
	if c, err := r.Cookie(langCookie); err == nil {
		cookie = c.Value
	}
	return i18n.L(i18n.Negotiate(cookie, r.Header.Get("Accept-Language")))
}

// entryLanguage mantém o idioma escolhido na página pública ao entrar no
// aplicativo. A query só vale nos dois pontos de entrada, nunca nos quadros.
func entryLanguage(r *http.Request) (i18n.Lang, bool) {
	if (r.Method == http.MethodGet || r.Method == http.MethodHead) && (r.URL.Path == "/" || r.URL.Path == "/semanas/nova") {
		return i18n.Parse(r.URL.Query().Get("lang"))
	}
	return "", false
}

func (s *Server) rememberEntryLanguage(w http.ResponseWriter, r *http.Request) {
	if lang, ok := entryLanguage(r); ok {
		http.SetCookie(w, s.preferenceCookie(langCookie, string(lang)))
	}
}

// handleTheme é o caminho sem JavaScript do botão sol/lua.
func (s *Server) handleTheme(w http.ResponseWriter, r *http.Request) {
	theme := themeDark
	if r.FormValue("theme") == themeLight {
		theme = themeLight
	}
	http.SetCookie(w, s.preferenceCookie(themeCookie, theme))
	s.redirectBack(w, r)
}

// handleLanguage guarda o idioma escolhido nas configurações.
func (s *Server) handleLanguage(w http.ResponseWriter, r *http.Request) {
	lang, ok := i18n.Parse(r.FormValue("lang"))
	if !ok {
		lang = i18n.PT
	}
	s.savePreference(w, r, langCookie, string(lang))
}

// handleCursor guarda o tipo de cursor: o do sistema ou o do produto.
func (s *Server) handleCursor(w http.ResponseWriter, r *http.Request) {
	v := cursorSystem
	if r.FormValue("cursor") == cursorCustom {
		v = cursorCustom
	}
	s.savePreference(w, r, cursorCookie, v)
}

// handleAccent guarda a cor de destaque.
func (s *Server) handleAccent(w http.ResponseWriter, r *http.Request) {
	v := r.FormValue("accent")
	if !validAccent(v) {
		v = accentMono
	}
	s.savePreference(w, r, accentCookie, v)
}

// savePreference grava o cookie e responde: 204 para o script, senão volta
// para a página de origem.
func (s *Server) savePreference(w http.ResponseWriter, r *http.Request, cookie, value string) {
	http.SetCookie(w, s.preferenceCookie(cookie, value))
	if wantsJSON(r) {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	s.redirectBack(w, r)
}

// preferenceCookie é um cookie de preferência: um ano, legível pelo script
// (não é segredo), com os demais atributos do padrão.
func (s *Server) preferenceCookie(name, value string) *http.Cookie {
	c := s.cookie(name, value, "/", int(prefTTL.Seconds())) //nolint:gosec // G124: HttpOnly desligado abaixo de propósito
	c.HttpOnly = false
	return c
}

// redirectBack volta para a página de origem quando ela é nossa; senão, para o hub.
func (s *Server) redirectBack(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, backTo(r), http.StatusSeeOther) //nolint:gosec // G710: backTo só devolve caminhos do próprio host
}

// backTo é o caminho da página de origem, se for deste host; senão, o hub.
// Só o caminho e a query são reaproveitados, nunca esquema ou host.
func backTo(r *http.Request) string {
	if ref := r.Header.Get("Referer"); ref != "" {
		if u, err := r.URL.Parse(ref); err == nil && u.Host == r.Host && strings.HasPrefix(u.Path, "/") {
			return u.RequestURI()
		}
	}
	return "/semanas"
}

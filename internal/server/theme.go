package server

import (
	"net/http"
	"strings"
	"time"
)

// O tema é um cookie legível pelo JavaScript: o servidor entrega a página já
// no tema certo (sem piscar) e o botão troca sem recarregar.
const (
	themeCookie = "weeklly_theme"
	themeDark   = "dark"
	themeLight  = "light"
	themeTTL    = 365 * 24 * time.Hour
)

func themeFrom(r *http.Request) string {
	if c, err := r.Cookie(themeCookie); err == nil && c.Value == themeLight {
		return themeLight
	}
	return themeDark
}

// handleTheme é o caminho sem JavaScript do botão sol/lua.
func (s *Server) handleTheme(w http.ResponseWriter, r *http.Request) {
	theme := themeDark
	if r.FormValue("theme") == themeLight {
		theme = themeLight
	}
	c := s.cookie(themeCookie, theme, "/", int(themeTTL.Seconds())) //nolint:gosec // G124: HttpOnly desligado abaixo de propósito
	c.HttpOnly = false                                              // o script lê e escreve o tema; não é segredo
	http.SetCookie(w, c)
	s.redirectBack(w, r)
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

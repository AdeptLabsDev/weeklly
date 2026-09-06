package server

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/AdeptLabsDev/weeklly/internal/i18n"
	"github.com/AdeptLabsDev/weeklly/internal/week"
)

// handleHome abre o app onde o usuário parou: a última semana aberta, ou a
// mais recente. Sem semanas, mostra o hub (princípio 1 do roadmap).
func (s *Server) handleHome(w http.ResponseWriter, r *http.Request) {
	s.rememberEntryLanguage(w, r)
	v := currentVisitor(r)
	if v.ok {
		target := v.session.LastWeekID
		if target == "" {
			list, err := s.opts.Store.ListWeeks(r.Context(), v.user.ID, v.user.WeekOrder)
			if err != nil {
				s.serverError(w, r, err)
				return
			}
			if len(list) > 0 {
				target = list[0].ID
			}
		}
		if target != "" {
			http.Redirect(w, r, weekURL(target), http.StatusSeeOther)
			return
		}
	}
	s.renderHub(w, r)
}

// handleHub lista as semanas do usuário.
func (s *Server) handleHub(w http.ResponseWriter, r *http.Request) {
	s.renderHub(w, r)
}

func (s *Server) renderHub(w http.ResponseWriter, r *http.Request) {
	nav, err := s.nav(r, "")
	if err != nil {
		s.serverError(w, r, err)
		return
	}
	l := s.locale(r)
	s.render(w, r, http.StatusOK, "hub", page{
		Title: s.title(l, l.T("hub.title")),
		Nav:   nav,
		Data:  hubView{Weeks: nav.Weeks, Order: nav.Order},
	})
}

// handleNewWeekPage é a versão em página do diálogo "Nova semana": funciona
// sem JavaScript e é o destino dos links que o script transforma em diálogo.
func (s *Server) handleNewWeekPage(w http.ResponseWriter, r *http.Request) {
	s.rememberEntryLanguage(w, r)
	s.renderForm(w, r, http.StatusOK, newWeekForm(s.locale(r)))
}

// handleCreateWeek cria a semana e abre o quadro. Um visitante sem sessão
// ganha uma aqui: é o primeiro momento em que há algo dele para guardar.
func (s *Server) handleCreateWeek(w http.ResponseWriter, r *http.Request) {
	l := s.locale(r)
	r.Body = http.MaxBytesReader(w, r.Body, 16<<10)
	if err := r.ParseForm(); err != nil {
		s.renderForm(w, r, http.StatusBadRequest, withError(newWeekForm(l), "", l.T("form.readError")))
		return
	}
	raw := r.PostFormValue("name")
	if _, err := week.CleanName(raw); err != nil {
		s.renderForm(w, r, http.StatusUnprocessableEntity, withError(newWeekForm(l), raw, l.Error(err)))
		return
	}

	v, err := s.ensureSession(w, r)
	if err != nil {
		s.serverError(w, r, err)
		return
	}
	wk, err := s.opts.Store.CreateWeek(r.Context(), v.user.ID, raw)
	if err != nil {
		s.serverError(w, r, err)
		return
	}
	if err := s.opts.Store.SetLastWeek(r.Context(), v.session.ID, wk.ID); err != nil {
		s.serverError(w, r, err)
		return
	}
	http.Redirect(w, r, weekURL(wk.ID), http.StatusSeeOther)
}

// handleBoard mostra uma semana. Semana de outra pessoa, inexistente ou sem
// sessão dá 404: a URL não revela nada.
func (s *Server) handleBoard(w http.ResponseWriter, r *http.Request) {
	wk, ok := s.ownedWeek(w, r)
	if !ok {
		return
	}
	v := currentVisitor(r)
	if v.session.LastWeekID != wk.ID {
		if err := s.opts.Store.SetLastWeek(r.Context(), v.session.ID, wk.ID); err != nil {
			s.serverError(w, r, err)
			return
		}
	}

	nav, err := s.nav(r, wk.ID)
	if err != nil {
		s.serverError(w, r, err)
		return
	}
	l := s.locale(r)
	rename := renameForm(l, wk)
	s.render(w, r, http.StatusOK, "board", page{
		Title:     s.title(l, wk.Name),
		BodyClass: "is-board",
		Nav:       nav,
		Data:      newBoardView(l, wk, s.today()),
		Rename:    &rename,
		Delete:    &deleteView{L: l, Action: weekURL(wk.ID) + "/excluir", Name: wk.Name, Cancel: weekURL(wk.ID)},
	})
}

// handleHealth confirma que o processo e o banco respondem. Sem cache: cada
// consulta é uma verificação real.
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	w.Header().Set("Cache-Control", "no-store")
	if err := s.opts.Store.Health(ctx); err != nil {
		s.opts.Logger.Error("healthz", "err", err, "request_id", requestID(r.Context()))
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"status": "degraded", "database": "unavailable"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "version": s.opts.Version})
}

// handleNotFound é a página 404, com a voz do produto e os mesmos headers.
func (s *Server) handleNotFound(w http.ResponseWriter, r *http.Request) {
	l := s.locale(r)
	s.renderMessage(w, r, http.StatusNotFound, messageView{
		Heading:  l.T("notFound.heading"),
		Body:     l.T("notFound.body"),
		LinkText: l.T("notFound.link"),
		LinkURL:  "/semanas",
	})
}

func (s *Server) renderMessage(w http.ResponseWriter, r *http.Request, status int, m messageView) {
	nav, err := s.nav(r, "")
	if err != nil {
		s.opts.Logger.Error("montando navegação", "err", err, "request_id", requestID(r.Context()))
		nav = navView{}
	}
	s.render(w, r, status, "message", page{
		Title: s.title(s.locale(r), m.Heading),
		Nav:   nav,
		Data:  m,
	})
}

// today devolve o dia da semana atual no fuso configurado. O fuso do
// navegador entra na Fase 1.
func (s *Server) today() week.Weekday {
	return week.Today(s.now(), s.opts.Config.Timezone)
}

// title monta o título da aba: "Nome · weeklly".
func (s *Server) title(l i18n.Locale, name string) string {
	return name + " · " + l.T("app.name")
}

// render completa a página com o que toda página tem (tema, idioma, diálogo
// de nova semana) e executa o template; erro vira 500.
func (s *Server) render(w http.ResponseWriter, r *http.Request, status int, name string, p page) {
	if p.Theme == "" {
		p.Theme = themeFrom(r)
	}
	if p.Public {
		p.Cursor = cursorSystem
		p.Accent = accentMono
	} else {
		p.Cursor = cursorFrom(r)
		p.Accent = accentFrom(r)
	}
	if p.L.IsZero() {
		p.L = s.locale(r)
	}
	if p.NewWeek.Action == "" {
		p.NewWeek = newWeekForm(p.L)
	}
	if p.SEO.Robots == "" {
		p.SEO.Robots = privateRobots
	}
	w.Header().Set("Content-Language", string(p.L.Lang()))
	w.Header().Set("X-Robots-Tag", p.SEO.Robots)
	if err := s.views.render(w, status, name, p); err != nil {
		s.serverError(w, r, err)
	}
}

func (s *Server) serverError(w http.ResponseWriter, r *http.Request, err error) {
	s.opts.Logger.Error("erro no handler", "err", err, "path", r.URL.Path, "request_id", requestID(r.Context()))
	w.Header().Set("X-Robots-Tag", privateRobots)
	w.Header().Set("Cache-Control", "private, no-store")
	http.Error(w, "erro interno", http.StatusInternalServerError)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v) // falha aqui é conexão fechada pelo cliente; não há o que fazer
}

func weekURL(id string) string { return "/semana/" + id }

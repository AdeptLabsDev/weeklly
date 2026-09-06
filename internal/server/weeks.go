package server

import (
	"errors"
	"net/http"

	"github.com/AdeptLabsDev/weeklly/internal/ids"
	"github.com/AdeptLabsDev/weeklly/internal/store"
	"github.com/AdeptLabsDev/weeklly/internal/week"
)

// ownedWeek carrega a semana da URL se ela é do visitante; senão responde 404.
func (s *Server) ownedWeek(w http.ResponseWriter, r *http.Request) (week.Week, bool) {
	v := currentVisitor(r)
	id := r.PathValue("id")
	if !v.ok || !ids.Valid(id) {
		s.handleNotFound(w, r)
		return week.Week{}, false
	}
	wk, err := s.opts.Store.Week(r.Context(), v.user.ID, id)
	if errors.Is(err, store.ErrNotFound) {
		s.handleNotFound(w, r)
		return week.Week{}, false
	}
	if err != nil {
		s.serverError(w, r, err)
		return week.Week{}, false
	}
	return wk, true
}

// handleRenamePage é a versão em página do diálogo "Renomear".
func (s *Server) handleRenamePage(w http.ResponseWriter, r *http.Request) {
	wk, ok := s.ownedWeek(w, r)
	if !ok {
		return
	}
	s.renderForm(w, r, http.StatusOK, renameForm(s.locale(r), wk))
}

// handleRename troca o nome da semana.
func (s *Server) handleRename(w http.ResponseWriter, r *http.Request) {
	wk, ok := s.ownedWeek(w, r)
	if !ok {
		return
	}
	l := s.locale(r)
	r.Body = http.MaxBytesReader(w, r.Body, 16<<10)
	if err := r.ParseForm(); err != nil {
		s.renderForm(w, r, http.StatusBadRequest, withError(renameForm(l, wk), "", l.T("form.readError")))
		return
	}
	raw := r.PostFormValue("name")
	err := s.opts.Store.RenameWeek(r.Context(), currentVisitor(r).user.ID, wk.ID, raw)
	if err != nil {
		if isNameError(err) {
			s.renderForm(w, r, http.StatusUnprocessableEntity, withError(renameForm(l, wk), raw, l.Error(err)))
			return
		}
		s.serverError(w, r, err)
		return
	}
	http.Redirect(w, r, weekURL(wk.ID), http.StatusSeeOther)
}

// handleDuplicate cria uma cópia da semana com as tarefas e abre a cópia.
func (s *Server) handleDuplicate(w http.ResponseWriter, r *http.Request) {
	wk, ok := s.ownedWeek(w, r)
	if !ok {
		return
	}
	l := s.locale(r)
	name := l.T("duplicate.suffix", wk.Name)
	if _, err := week.CleanName(name); err != nil {
		name = l.T("duplicate.fallback")
	}
	v := currentVisitor(r)
	copyWeek, err := s.opts.Store.DuplicateWeek(r.Context(), v.user.ID, wk.ID, name)
	if err != nil {
		s.serverError(w, r, err)
		return
	}
	if err := s.opts.Store.SetLastWeek(r.Context(), v.session.ID, copyWeek.ID); err != nil {
		s.serverError(w, r, err)
		return
	}
	http.Redirect(w, r, weekURL(copyWeek.ID), http.StatusSeeOther)
}

// handleDeletePage é a versão em página do diálogo "Excluir".
func (s *Server) handleDeletePage(w http.ResponseWriter, r *http.Request) {
	wk, ok := s.ownedWeek(w, r)
	if !ok {
		return
	}
	nav, err := s.nav(r, wk.ID)
	if err != nil {
		s.serverError(w, r, err)
		return
	}
	l := s.locale(r)
	s.render(w, r, http.StatusOK, "confirm-delete", page{
		Title: s.title(l, l.T("delete.heading", wk.Name)),
		Nav:   nav,
		Data:  deleteView{L: l, Action: weekURL(wk.ID) + "/excluir", Name: wk.Name, Cancel: weekURL(wk.ID)},
	})
}

// handleDelete apaga a semana e volta ao hub.
func (s *Server) handleDelete(w http.ResponseWriter, r *http.Request) {
	wk, ok := s.ownedWeek(w, r)
	if !ok {
		return
	}
	if err := s.opts.Store.DeleteWeek(r.Context(), currentVisitor(r).user.ID, wk.ID); err != nil {
		s.serverError(w, r, err)
		return
	}
	http.Redirect(w, r, "/semanas", http.StatusSeeOther)
}

// handleOrder guarda a ordem preferida da lista de semanas.
func (s *Server) handleOrder(w http.ResponseWriter, r *http.Request) {
	v := currentVisitor(r)
	if v.ok {
		order := store.ParseWeekOrder(r.FormValue("order"))
		if err := s.opts.Store.SetWeekOrder(r.Context(), v.user.ID, order); err != nil && !errors.Is(err, store.ErrNotFound) {
			s.serverError(w, r, err)
			return
		}
	}
	if wantsJSON(r) {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	s.redirectBack(w, r)
}

func (s *Server) renderForm(w http.ResponseWriter, r *http.Request, status int, form formView) {
	nav, err := s.nav(r, "")
	if err != nil {
		s.serverError(w, r, err)
		return
	}
	s.render(w, r, status, "week-form", page{
		Title: s.title(form.L, form.Heading),
		Nav:   nav,
		Data:  form,
	})
}

func withError(f formView, name, msg string) formView {
	if name != "" {
		f.Name = name
	}
	f.Error = msg
	return f
}

func isNameError(err error) bool {
	return errors.Is(err, week.ErrEmptyName) || errors.Is(err, week.ErrNameTooLong) || errors.Is(err, week.ErrInvalidName)
}

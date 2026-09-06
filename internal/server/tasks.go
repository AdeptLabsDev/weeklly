package server

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/AdeptLabsDev/weeklly/internal/ids"
	"github.com/AdeptLabsDev/weeklly/internal/store"
	"github.com/AdeptLabsDev/weeklly/internal/week"
)

// Tarefas têm dois caminhos com o mesmo formulário: sem JavaScript, o
// navegador envia o form e volta para o quadro no dia certo; com
// JavaScript, o mesmo POST pede JSON (Accept) e recebe a tarefa já
// renderizada, que o script encaixa no lugar.

func wantsJSON(r *http.Request) bool {
	return strings.Contains(r.Header.Get("Accept"), "application/json")
}

// handleAddTask cria uma tarefa no fim do dia.
func (s *Server) handleAddTask(w http.ResponseWriter, r *http.Request) {
	v := currentVisitor(r)
	weekID := r.PathValue("id")
	if !v.ok || !ids.Valid(weekID) {
		s.taskError(w, r, http.StatusNotFound, s.locale(r).T("task.error.week"))
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, 16<<10)
	if err := r.ParseForm(); err != nil {
		s.taskError(w, r, http.StatusBadRequest, s.locale(r).T("task.error.read"))
		return
	}
	day, err := strconv.Atoi(r.PostFormValue("weekday"))
	if err != nil || !week.Weekday(day).Valid() {
		s.taskError(w, r, http.StatusBadRequest, s.locale(r).T("task.error.day"))
		return
	}

	t, err := s.opts.Store.AddTask(r.Context(), v.user.ID, weekID, week.Weekday(day), r.PostFormValue("title"), r.PostFormValue("time"))
	if err != nil {
		s.taskFailure(w, r, err)
		return
	}
	s.taskOK(w, r, http.StatusCreated, weekID, t)
}

// handleEditTask muda título e/ou horário.
func (s *Server) handleEditTask(w http.ResponseWriter, r *http.Request) {
	v := currentVisitor(r)
	if !v.ok {
		s.taskError(w, r, http.StatusNotFound, s.locale(r).T("task.error.notHere"))
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, 16<<10)
	if err := r.ParseForm(); err != nil {
		s.taskError(w, r, http.StatusBadRequest, s.locale(r).T("task.error.read"))
		return
	}
	var patch store.TaskPatch
	if title, ok := r.PostForm["title"]; ok && len(title) > 0 {
		patch.Title = &title[0]
	}
	if at, ok := r.PostForm["time"]; ok && len(at) > 0 {
		patch.Time = &at[0]
	}
	if patch.Title == nil && patch.Time == nil {
		s.taskError(w, r, http.StatusBadRequest, s.locale(r).T("task.error.nothing"))
		return
	}
	s.applyPatch(w, r, patch)
}

// handleDoneTask marca ou desmarca a tarefa como feita.
func (s *Server) handleDoneTask(w http.ResponseWriter, r *http.Request) {
	v := currentVisitor(r)
	if !v.ok {
		s.taskError(w, r, http.StatusNotFound, s.locale(r).T("task.error.notHere"))
		return
	}
	done := r.FormValue("done") != "0"
	s.applyPatch(w, r, store.TaskPatch{Done: &done})
}

func (s *Server) applyPatch(w http.ResponseWriter, r *http.Request, patch store.TaskPatch) {
	v := currentVisitor(r)
	t, err := s.opts.Store.UpdateTask(r.Context(), v.user.ID, r.PathValue("id"), patch)
	if err != nil {
		s.taskFailure(w, r, err)
		return
	}
	s.taskOK(w, r, http.StatusOK, r.FormValue("week"), t)
}

// handleMoveTask põe a tarefa em outra posição, no mesmo dia ou em outro.
func (s *Server) handleMoveTask(w http.ResponseWriter, r *http.Request) {
	v := currentVisitor(r)
	if !v.ok {
		s.taskError(w, r, http.StatusNotFound, s.locale(r).T("task.error.notHere"))
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, 16<<10)
	if err := r.ParseForm(); err != nil {
		s.taskError(w, r, http.StatusBadRequest, s.locale(r).T("task.error.read"))
		return
	}
	day, err := strconv.Atoi(r.PostFormValue("weekday"))
	if err != nil || !week.Weekday(day).Valid() {
		s.taskError(w, r, http.StatusBadRequest, s.locale(r).T("task.error.day"))
		return
	}
	position, err := strconv.Atoi(r.PostFormValue("position"))
	if err != nil || position < 0 {
		s.taskError(w, r, http.StatusBadRequest, s.locale(r).T("task.error.position"))
		return
	}
	t, err := s.opts.Store.MoveTask(r.Context(), v.user.ID, r.PathValue("id"), week.Weekday(day), position)
	if err != nil {
		s.taskFailure(w, r, err)
		return
	}
	s.taskOK(w, r, http.StatusOK, r.FormValue("week"), t)
}

// handleDeleteTask apaga a tarefa.
func (s *Server) handleDeleteTask(w http.ResponseWriter, r *http.Request) {
	v := currentVisitor(r)
	if !v.ok {
		s.taskError(w, r, http.StatusNotFound, s.locale(r).T("task.error.notHere"))
		return
	}
	if err := s.opts.Store.DeleteTask(r.Context(), v.user.ID, r.PathValue("id")); err != nil {
		s.taskFailure(w, r, err)
		return
	}
	if wantsJSON(r) {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	s.redirectBack(w, r)
}

// taskOK responde com a tarefa renderizada (JSON) ou volta para o quadro.
func (s *Server) taskOK(w http.ResponseWriter, r *http.Request, status int, weekID string, t week.Task) {
	if wantsJSON(r) {
		html, err := s.views.partial("task-item", newTaskView(s.locale(r), t))
		if err != nil {
			s.serverError(w, r, err)
			return
		}
		writeJSON(w, status, map[string]any{
			"id":       t.ID,
			"weekday":  int(t.Weekday),
			"position": t.Position,
			"done":     t.Done,
			"html":     html,
		})
		return
	}
	if !ids.Valid(weekID) {
		s.redirectBack(w, r)
		return
	}
	// weekID passou por ids.Valid: só [a-z2-7], nunca um host ou esquema.
	http.Redirect(w, r, fmt.Sprintf("%s#dia-%d", weekURL(weekID), int(t.Weekday)), http.StatusSeeOther) //nolint:gosec // G710: caminho montado de um id validado
}

// taskFailure traduz o erro do domínio em status e frase.
func (s *Server) taskFailure(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, store.ErrNotFound):
		s.taskError(w, r, http.StatusNotFound, s.locale(r).T("task.error.notHere"))
	case errors.Is(err, week.ErrDayFull), errors.Is(err, week.ErrEmptyTitle), errors.Is(err, week.ErrTitleTooLong),
		errors.Is(err, week.ErrInvalidTime), errors.Is(err, week.ErrInvalidName):
		s.taskError(w, r, http.StatusUnprocessableEntity, s.locale(r).Error(err))
	default:
		s.serverError(w, r, err)
	}
}

// taskError responde JSON para o script ou uma página curta para o navegador.
func (s *Server) taskError(w http.ResponseWriter, r *http.Request, status int, msg string) {
	if wantsJSON(r) {
		writeJSON(w, status, map[string]string{"error": msg})
		return
	}
	l := s.locale(r)
	s.renderMessage(w, r, status, messageView{
		Heading:  l.T("task.error.title"),
		Body:     msg,
		LinkText: l.T("task.error.back"),
		LinkURL:  backTo(r),
	})
}

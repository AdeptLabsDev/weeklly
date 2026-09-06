package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/AdeptLabsDev/weeklly/internal/store"
)

const sessionCookie = "weeklly_session"

type sessionKey struct{}

// visitor é o que a sessão diz sobre quem está pedindo. ok é falso quando
// não há sessão válida.
type visitor struct {
	session store.Session
	user    store.User
	ok      bool
}

// withSession carrega a sessão do cookie uma vez por requisição. Cookie
// desconhecido ou vencido é apagado no caminho.
func (s *Server) withSession(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var v visitor
		if c, err := r.Cookie(sessionCookie); err == nil && c.Value != "" {
			sess, user, err := s.opts.Store.SessionByToken(r.Context(), c.Value)
			switch {
			case err == nil:
				v = visitor{session: sess, user: user, ok: true}
			case errors.Is(err, store.ErrNotFound):
				s.clearCookie(w, sessionCookie)
			default:
				s.serverError(w, r, fmt.Errorf("carregando sessão: %w", err))
				return
			}
		}
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), sessionKey{}, v)))
	})
}

func currentVisitor(r *http.Request) visitor {
	v, _ := r.Context().Value(sessionKey{}).(visitor)
	return v
}

// ensureSession devolve a sessão atual ou cria um usuário anônimo com uma
// sessão nova. É o que acontece quando um visitante cria a primeira semana.
func (s *Server) ensureSession(w http.ResponseWriter, r *http.Request) (visitor, error) {
	if v := currentVisitor(r); v.ok {
		return v, nil
	}
	ctx := r.Context()
	user, err := s.opts.Store.CreateAnonymousUser(ctx)
	if err != nil {
		return visitor{}, err
	}
	token, sess, err := s.opts.Store.CreateSession(ctx, user.ID)
	if err != nil {
		return visitor{}, err
	}
	s.setSessionCookie(w, token)
	return visitor{session: sess, user: user, ok: true}, nil
}

// startSession abre uma sessão nova para o usuário e grava o cookie.
func (s *Server) startSession(w http.ResponseWriter, r *http.Request, userID string) error {
	token, _, err := s.opts.Store.CreateSession(r.Context(), userID)
	if err != nil {
		return err
	}
	s.setSessionCookie(w, token)
	return nil
}

func (s *Server) setSessionCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, s.cookie(sessionCookie, token, "/", int(store.SessionTTL.Seconds())))
}

func (s *Server) clearCookie(w http.ResponseWriter, name string) {
	http.SetCookie(w, s.cookie(name, "", "/", -1))
}

// cookie monta todo cookie do servidor com os mesmos atributos: HttpOnly,
// SameSite=Lax (o retorno do Google é uma navegação de topo e precisa levar
// o cookie) e Secure fora de development, onde o site roda em http://localhost.
func (s *Server) cookie(name, value, path string, maxAge int) *http.Cookie {
	return &http.Cookie{ //nolint:gosec // G124: Secure só é falso em development
		Name:     name,
		Value:    value,
		Path:     path,
		MaxAge:   maxAge,
		HttpOnly: true,
		Secure:   !s.opts.Config.IsDev(),
		SameSite: http.SameSiteLaxMode,
	}
}

package server

import (
	"errors"
	"net/http"
	"time"

	"github.com/AdeptLabsDev/weeklly/internal/auth"
	"github.com/AdeptLabsDev/weeklly/internal/store"
)

const (
	callbackPath = "/entrar/google/callback"
	loginCookie  = "weeklly_login"
	loginTTL     = 10 * time.Minute
)

// handleLoginStart manda o navegador para o Google, guardando num cookie
// curto o que o retorno precisa conferir (D14).
func (s *Server) handleLoginStart(w http.ResponseWriter, r *http.Request) {
	if s.opts.Google == nil {
		l := s.locale(r)
		s.renderMessage(w, r, http.StatusServiceUnavailable, messageView{
			Heading:  l.T("login.off.heading"),
			Body:     l.T("login.off.body"),
			LinkText: l.T("back"),
			LinkURL:  "/",
		})
		return
	}
	c := auth.NewChallenge()
	http.SetCookie(w, s.cookie(loginCookie, c.Encode(), callbackPath, int(loginTTL.Seconds())))
	http.Redirect(w, r, s.opts.Google.AuthURL(c), http.StatusFound)
}

// handleLoginCallback recebe o código do Google, verifica tudo e liga a
// identidade à conta certa:
//   - conta Google já conhecida: as semanas do visitante anônimo migram para ela;
//   - conta nova e visitante anônimo: o próprio visitante ganha a identidade;
//   - conta nova sem visitante: cria usuário e sessão.
func (s *Server) handleLoginCallback(w http.ResponseWriter, r *http.Request) {
	if s.opts.Google == nil {
		s.handleNotFound(w, r)
		return
	}
	cookie, err := r.Cookie(loginCookie)
	s.clearLoginCookie(w)
	challenge, ok := auth.Challenge{}, false
	if err == nil {
		challenge, ok = auth.DecodeChallenge(cookie.Value)
	}
	q := r.URL.Query()
	l := s.locale(r)

	if q.Get("error") != "" {
		s.renderMessage(w, r, http.StatusOK, messageView{
			Heading:  l.T("login.cancelled.heading"),
			Body:     l.T("login.cancelled.body"),
			LinkText: l.T("back"),
			LinkURL:  "/",
		})
		return
	}
	if !ok || q.Get("state") == "" || q.Get("state") != challenge.State || q.Get("code") == "" {
		s.renderMessage(w, r, http.StatusBadRequest, messageView{
			Heading:  l.T("login.mismatch.heading"),
			Body:     l.T("login.mismatch.body"),
			LinkText: l.T("login.mismatch.link"),
			LinkURL:  "/entrar/google",
		})
		return
	}

	identity, err := s.opts.Google.Exchange(r.Context(), q.Get("code"), challenge)
	if errors.Is(err, auth.ErrEmailUnverified) {
		s.renderMessage(w, r, http.StatusForbidden, messageView{
			Heading:  l.T("login.unverified.heading"),
			Body:     l.T("login.unverified.body"),
			LinkText: l.T("login.unverified.link"),
			LinkURL:  "/entrar/google",
		})
		return
	}
	if err != nil {
		s.opts.Logger.Error("login com google", "err", err, "request_id", requestID(r.Context()))
		s.renderMessage(w, r, http.StatusBadGateway, messageView{
			Heading:  l.T("login.failed.heading"),
			Body:     l.T("login.failed.body"),
			LinkText: l.T("login.failed.link"),
			LinkURL:  "/entrar/google",
		})
		return
	}

	if err := s.signIn(w, r, identity); err != nil {
		if errors.Is(err, store.ErrEmailTaken) {
			s.renderMessage(w, r, http.StatusConflict, messageView{
				Heading:  l.T("login.taken.heading"),
				Body:     l.T("login.taken.body"),
				LinkText: l.T("back"),
				LinkURL:  "/",
			})
			return
		}
		s.serverError(w, r, err)
		return
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (s *Server) signIn(w http.ResponseWriter, r *http.Request, id auth.Identity) error {
	ctx := r.Context()
	st := s.opts.Store
	g := store.GoogleIdentity{Sub: id.Sub, Email: id.Email, Name: id.Name, Picture: id.Picture}
	v := currentVisitor(r)

	existing, err := st.UserByGoogleSub(ctx, id.Sub)
	switch {
	case err == nil:
		// Conta conhecida. Um visitante anônimo traz as semanas dele junto.
		if v.ok && v.user.ID != existing.ID && v.user.Anonymous() {
			if err := st.MergeUsers(ctx, v.user.ID, existing.ID); err != nil {
				return err
			}
		} else if !v.ok || v.user.ID != existing.ID {
			if err := s.startSession(w, r, existing.ID); err != nil {
				return err
			}
		}
		return st.LinkGoogle(ctx, existing.ID, g) // renova nome e foto
	case errors.Is(err, store.ErrNotFound):
		if v.ok && v.user.Anonymous() {
			return st.LinkGoogle(ctx, v.user.ID, g)
		}
		user, err := st.CreateAnonymousUser(ctx)
		if err != nil {
			return err
		}
		if err := st.LinkGoogle(ctx, user.ID, g); err != nil {
			return err
		}
		return s.startSession(w, r, user.ID)
	default:
		return err
	}
}

// handleLogout encerra a sessão deste navegador.
func (s *Server) handleLogout(w http.ResponseWriter, r *http.Request) {
	if c, err := r.Cookie(sessionCookie); err == nil && c.Value != "" {
		if err := s.opts.Store.DeleteSession(r.Context(), c.Value); err != nil {
			s.serverError(w, r, err)
			return
		}
	}
	s.clearCookie(w, sessionCookie)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (s *Server) clearLoginCookie(w http.ResponseWriter) {
	http.SetCookie(w, s.cookie(loginCookie, "", callbackPath, -1))
}

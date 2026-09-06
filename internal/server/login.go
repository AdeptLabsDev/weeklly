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
		s.renderMessage(w, r, http.StatusServiceUnavailable, messageView{
			Heading:  "Login ainda não configurado",
			Body:     "Este ambiente não tem as credenciais do Google. Dá para usar o weeklly sem entrar: as semanas ficam neste navegador até você entrar em outro momento.",
			LinkText: "Voltar",
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

	if q.Get("error") != "" {
		s.renderMessage(w, r, http.StatusOK, messageView{
			Heading:  "Login cancelado",
			Body:     "Você não entrou. Nada mudou por aqui.",
			LinkText: "Voltar",
			LinkURL:  "/",
		})
		return
	}
	if !ok || q.Get("state") == "" || q.Get("state") != challenge.State || q.Get("code") == "" {
		s.renderMessage(w, r, http.StatusBadRequest, messageView{
			Heading:  "Não deu para entrar",
			Body:     "O retorno do Google não bateu com o pedido, o que acontece quando o login demora mais de 10 minutos ou abre em outra aba. Tente de novo.",
			LinkText: "Entrar com Google",
			LinkURL:  "/entrar/google",
		})
		return
	}

	identity, err := s.opts.Google.Exchange(r.Context(), q.Get("code"), challenge)
	if errors.Is(err, auth.ErrEmailUnverified) {
		s.renderMessage(w, r, http.StatusForbidden, messageView{
			Heading:  "E-mail sem confirmação",
			Body:     "O Google não confirma o e-mail desta conta, e o weeklly só aceita e-mails confirmados. Use outra conta.",
			LinkText: "Tentar com outra conta",
			LinkURL:  "/entrar/google",
		})
		return
	}
	if err != nil {
		s.opts.Logger.Error("login com google", "err", err, "request_id", requestID(r.Context()))
		s.renderMessage(w, r, http.StatusBadGateway, messageView{
			Heading:  "O Google não respondeu como esperado",
			Body:     "Não foi possível confirmar sua conta agora. Tente de novo em instantes.",
			LinkText: "Tentar de novo",
			LinkURL:  "/entrar/google",
		})
		return
	}

	if err := s.signIn(w, r, identity); err != nil {
		if errors.Is(err, store.ErrEmailTaken) {
			s.renderMessage(w, r, http.StatusConflict, messageView{
				Heading:  "Este e-mail já tem conta",
				Body:     "Já existe uma conta com este e-mail ligada a outra forma de entrar. Entre por ela, ou use outra conta Google.",
				LinkText: "Voltar",
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

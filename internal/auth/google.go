// Package auth implementa o login com o Google por OpenID Connect,
// fluxo authorization code com PKCE, usando só a biblioteca padrão (D14).
//
// O que sai daqui é uma Identity verificada: assinatura do id_token conferida
// contra as chaves públicas do Google, emissor, audiência, validade e nonce.
package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Endpoints são as URLs do provedor. Parametrizadas para que os testes usem
// um provedor falso.
type Endpoints struct {
	Auth   string
	Token  string
	JWKS   string
	Issuer string
}

// GoogleEndpoints são as URLs públicas do Google.
var GoogleEndpoints = Endpoints{ //nolint:gosec // G101: são URLs públicas, não credenciais
	Auth:   "https://accounts.google.com/o/oauth2/v2/auth",
	Token:  "https://oauth2.googleapis.com/token",
	JWKS:   "https://www.googleapis.com/oauth2/v3/certs",
	Issuer: "https://accounts.google.com",
}

// Google é o cliente do login.
type Google struct {
	clientID     string
	clientSecret string
	redirectURL  string
	endpoints    Endpoints
	client       *http.Client
	keys         *keySet
	now          func() time.Time
}

// NewGoogle monta o cliente. redirectURL é a URL absoluta do callback.
func NewGoogle(clientID, clientSecret, redirectURL string, endpoints Endpoints, client *http.Client) *Google {
	if client == nil {
		client = &http.Client{Timeout: 15 * time.Second}
	}
	return &Google{
		clientID:     clientID,
		clientSecret: clientSecret,
		redirectURL:  redirectURL,
		endpoints:    endpoints,
		client:       client,
		keys:         newKeySet(endpoints.JWKS, client),
		now:          time.Now,
	}
}

// Challenge é o que o servidor guarda entre o início do login e o retorno:
// state contra CSRF, nonce contra replay do id_token, verifier do PKCE.
type Challenge struct {
	State    string
	Nonce    string
	Verifier string
}

// NewChallenge gera os três segredos aleatórios.
func NewChallenge() Challenge {
	return Challenge{State: random(24), Nonce: random(24), Verifier: random(48)}
}

// Encode serializa para um cookie. Os três valores são base64url, então o
// ponto é um separador seguro.
func (c Challenge) Encode() string {
	return c.State + "." + c.Nonce + "." + c.Verifier
}

// DecodeChallenge lê o que Encode gravou.
func DecodeChallenge(s string) (Challenge, bool) {
	parts := strings.Split(s, ".")
	if len(parts) != 3 || parts[0] == "" || parts[1] == "" || parts[2] == "" {
		return Challenge{}, false
	}
	return Challenge{State: parts[0], Nonce: parts[1], Verifier: parts[2]}, true
}

// AuthURL é para onde o navegador vai para escolher a conta Google.
func (g *Google) AuthURL(c Challenge) string {
	sum := sha256.Sum256([]byte(c.Verifier))
	q := url.Values{
		"client_id":             {g.clientID},
		"redirect_uri":          {g.redirectURL},
		"response_type":         {"code"},
		"scope":                 {"openid email profile"},
		"state":                 {c.State},
		"nonce":                 {c.Nonce},
		"code_challenge":        {base64.RawURLEncoding.EncodeToString(sum[:])},
		"code_challenge_method": {"S256"},
		"prompt":                {"select_account"},
	}
	return g.endpoints.Auth + "?" + q.Encode()
}

// Identity é quem o Google diz que a pessoa é, já verificado.
type Identity struct {
	Sub     string
	Email   string
	Name    string
	Picture string
}

// ErrEmailUnverified é devolvido quando o Google não confirma o e-mail.
var ErrEmailUnverified = errors.New("o Google não confirmou este e-mail")

// Exchange troca o código do callback por uma identidade verificada.
func (g *Google) Exchange(ctx context.Context, code string, c Challenge) (Identity, error) {
	form := url.Values{
		"grant_type":    {"authorization_code"},
		"code":          {code},
		"redirect_uri":  {g.redirectURL},
		"client_id":     {g.clientID},
		"client_secret": {g.clientSecret},
		"code_verifier": {c.Verifier},
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, g.endpoints.Token, strings.NewReader(form.Encode()))
	if err != nil {
		return Identity{}, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := g.client.Do(req)
	if err != nil {
		return Identity{}, fmt.Errorf("trocando código: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return Identity{}, err
	}

	var tok struct {
		IDToken          string `json:"id_token"`
		Error            string `json:"error"`
		ErrorDescription string `json:"error_description"`
	}
	if err := json.Unmarshal(body, &tok); err != nil {
		return Identity{}, fmt.Errorf("resposta do token ilegível (status %d)", resp.StatusCode)
	}
	if resp.StatusCode != http.StatusOK || tok.Error != "" {
		return Identity{}, fmt.Errorf("token recusado (status %d): %s %s", resp.StatusCode, tok.Error, tok.ErrorDescription)
	}
	if tok.IDToken == "" {
		return Identity{}, errors.New("resposta do token sem id_token")
	}

	claims, err := g.verifyIDToken(ctx, tok.IDToken, c.Nonce)
	if err != nil {
		return Identity{}, err
	}
	if !claims.EmailVerified || claims.Email == "" {
		return Identity{}, ErrEmailUnverified
	}
	return Identity{Sub: claims.Sub, Email: claims.Email, Name: claims.Name, Picture: claims.Picture}, nil
}

func random(n int) string {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		panic(err) // crypto/rand só falha se o sistema estiver sem entropia; não há o que fazer
	}
	return base64.RawURLEncoding.EncodeToString(b)
}

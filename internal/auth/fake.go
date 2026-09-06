package auth

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"math/big"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"time"
)

// FakeProvider é um provedor OpenID Connect de mentira para testes: emite
// id_tokens assinados com uma chave própria e publica a chave pública.
//
// Fica no pacote (e não em _test.go) para que o servidor HTTP também possa
// testar o fluxo de ponta a ponta.
type FakeProvider struct {
	Server   *httptest.Server
	key      *rsa.PrivateKey
	kid      string
	clientID string

	mu    sync.Mutex
	codes map[string]FakeUser // código de autorização → quem está entrando
	// RejectTokens faz o endpoint de token responder erro.
	RejectTokens bool
}

// FakeUser é a pessoa que o provedor falso vai devolver.
type FakeUser struct {
	Sub           string
	Email         string
	EmailVerified bool
	Name          string
	Picture       string
	// Nonce sobrescreve o nonce do pedido; vazio usa o correto.
	Nonce string
	// ExpiresIn sobrescreve a validade; zero usa uma hora.
	ExpiresIn time.Duration
}

// NewFakeProvider sobe o provedor. Feche com Close.
func NewFakeProvider(clientID string) *FakeProvider {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		panic(err)
	}
	p := &FakeProvider{key: key, kid: "fake-key-1", clientID: clientID, codes: map[string]FakeUser{}}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /certs", p.handleCerts)
	mux.HandleFunc("POST /token", p.handleToken)
	p.Server = httptest.NewServer(mux)
	return p
}

// Close derruba o servidor.
func (p *FakeProvider) Close() { p.Server.Close() }

// Endpoints são as URLs do provedor falso.
func (p *FakeProvider) Endpoints() Endpoints {
	return Endpoints{
		Auth:   p.Server.URL + "/auth",
		Token:  p.Server.URL + "/token",
		JWKS:   p.Server.URL + "/certs",
		Issuer: p.Server.URL,
	}
}

// Authorize simula a pessoa escolhendo a conta: devolve o código que o
// callback vai receber. nonce é o que veio na URL de autorização.
func (p *FakeProvider) Authorize(user FakeUser, nonce string) string {
	if user.Nonce == "" {
		user.Nonce = nonce
	}
	code := random(16)
	p.mu.Lock()
	p.codes[code] = user
	p.mu.Unlock()
	return code
}

// NonceFrom extrai o nonce de uma URL de autorização.
func NonceFrom(authURL string) string {
	u, err := url.Parse(authURL)
	if err != nil {
		return ""
	}
	return u.Query().Get("nonce")
}

func (p *FakeProvider) handleCerts(w http.ResponseWriter, _ *http.Request) {
	pub := &p.key.PublicKey
	doc := map[string]any{"keys": []map[string]string{{
		"kty": "RSA",
		"kid": p.kid,
		"alg": "RS256",
		"use": "sig",
		"n":   base64.RawURLEncoding.EncodeToString(pub.N.Bytes()),
		"e":   base64.RawURLEncoding.EncodeToString(big.NewInt(int64(pub.E)).Bytes()),
	}}}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(doc)
}

func (p *FakeProvider) handleToken(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if p.RejectTokens {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "invalid_grant", "error_description": "rejeitado de propósito"})
		return
	}
	if err := r.ParseForm(); err != nil || r.PostForm.Get("client_id") != p.clientID || r.PostForm.Get("code_verifier") == "" {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "invalid_client"})
		return
	}
	p.mu.Lock()
	user, ok := p.codes[r.PostForm.Get("code")]
	delete(p.codes, r.PostForm.Get("code"))
	p.mu.Unlock()
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "invalid_grant"})
		return
	}
	_ = json.NewEncoder(w).Encode(map[string]string{
		"access_token": "fake",
		"token_type":   "Bearer",
		"id_token":     p.sign(user),
	})
}

func (p *FakeProvider) sign(user FakeUser) string {
	expires := time.Hour
	if user.ExpiresIn != 0 {
		expires = user.ExpiresIn
	}
	now := time.Now()
	header, _ := json.Marshal(map[string]string{"alg": "RS256", "kid": p.kid, "typ": "JWT"})
	payload, _ := json.Marshal(map[string]any{
		"iss":            p.Server.URL,
		"aud":            p.clientID,
		"sub":            user.Sub,
		"email":          user.Email,
		"email_verified": user.EmailVerified,
		"name":           user.Name,
		"picture":        user.Picture,
		"nonce":          user.Nonce,
		"iat":            now.Unix(),
		"exp":            now.Add(expires).Unix(),
	})
	signing := base64.RawURLEncoding.EncodeToString(header) + "." + base64.RawURLEncoding.EncodeToString(payload)
	digest := sha256.Sum256([]byte(signing))
	sig, err := rsa.SignPKCS1v15(rand.Reader, p.key, crypto.SHA256, digest[:])
	if err != nil {
		panic(err)
	}
	return signing + "." + base64.RawURLEncoding.EncodeToString(sig)
}

// SignForeign assina um token com outra chave, para testar assinatura inválida.
func (p *FakeProvider) SignForeign(user FakeUser) string {
	other := &FakeProvider{key: mustKey(), kid: p.kid, clientID: p.clientID, Server: p.Server}
	return other.sign(user)
}

func mustKey() *rsa.PrivateKey {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		panic(err)
	}
	return key
}

// tokenParts ajuda os testes a inspecionar um token.
func tokenParts(raw string) []string { return strings.Split(raw, ".") }

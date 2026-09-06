package auth

import (
	"context"
	"errors"
	"net/url"
	"strings"
	"testing"
	"time"
)

const clientID = "cliente-de-teste.apps.googleusercontent.com"

func newTestGoogle(t *testing.T) (*Google, *FakeProvider) {
	t.Helper()
	p := NewFakeProvider(clientID)
	t.Cleanup(p.Close)
	g := NewGoogle(clientID, "segredo", "http://127.0.0.1:8080/entrar/google/callback", p.Endpoints(), p.Server.Client())
	return g, p
}

func TestChallengeRoundTrip(t *testing.T) {
	c := NewChallenge()
	if c.State == "" || c.Nonce == "" || c.Verifier == "" || c.State == c.Nonce {
		t.Fatalf("Challenge = %+v", c)
	}
	back, ok := DecodeChallenge(c.Encode())
	if !ok || back != c {
		t.Errorf("DecodeChallenge(Encode()) = %+v, %v", back, ok)
	}
	for _, bad := range []string{"", "a.b", "a..c", "a.b.c.d"} {
		if _, ok := DecodeChallenge(bad); ok {
			t.Errorf("DecodeChallenge(%q) aceitou", bad)
		}
	}
}

func TestAuthURLCarriesPKCEAndNonce(t *testing.T) {
	g, _ := newTestGoogle(t)
	c := NewChallenge()
	u, err := url.Parse(g.AuthURL(c))
	if err != nil {
		t.Fatal(err)
	}
	q := u.Query()
	if q.Get("client_id") != clientID || q.Get("state") != c.State || q.Get("nonce") != c.Nonce {
		t.Errorf("query = %v", q)
	}
	if q.Get("code_challenge_method") != "S256" || q.Get("code_challenge") == "" || q.Get("code_challenge") == c.Verifier {
		t.Error("PKCE ausente ou verifier vazando na URL")
	}
	if !strings.Contains(q.Get("scope"), "openid") || q.Get("response_type") != "code" {
		t.Errorf("scope/response_type = %q/%q", q.Get("scope"), q.Get("response_type"))
	}
}

func TestExchangeHappyPath(t *testing.T) {
	g, p := newTestGoogle(t)
	c := NewChallenge()
	code := p.Authorize(FakeUser{Sub: "g-1", Email: "miguel@example.com", EmailVerified: true, Name: "Miguel", Picture: "https://lh3.googleusercontent.com/a"}, NonceFrom(g.AuthURL(c)))

	id, err := g.Exchange(context.Background(), code, c)
	if err != nil {
		t.Fatalf("Exchange: %v", err)
	}
	if id.Sub != "g-1" || id.Email != "miguel@example.com" || id.Name != "Miguel" || id.Picture == "" {
		t.Errorf("Identity = %+v", id)
	}

	// O código só vale uma vez.
	if _, err := g.Exchange(context.Background(), code, c); err == nil {
		t.Error("código reutilizado foi aceito")
	}
}

func TestExchangeRejectsBadTokens(t *testing.T) {
	g, p := newTestGoogle(t)
	ctx := context.Background()

	cases := map[string]FakeUser{
		"nonce errado":         {Sub: "g-1", Email: "a@b.c", EmailVerified: true, Nonce: "outro"},
		"e-mail sem confirmar": {Sub: "g-1", Email: "a@b.c", EmailVerified: false},
		"vencido":              {Sub: "g-1", Email: "a@b.c", EmailVerified: true, ExpiresIn: -time.Hour},
		"sem sub":              {Email: "a@b.c", EmailVerified: true},
	}
	for name, user := range cases {
		t.Run(name, func(t *testing.T) {
			c := NewChallenge()
			code := p.Authorize(user, c.Nonce)
			if _, err := g.Exchange(ctx, code, c); err == nil {
				t.Error("token inválido foi aceito")
			} else if name == "e-mail sem confirmar" && !errors.Is(err, ErrEmailUnverified) {
				t.Errorf("erro = %v, quero ErrEmailUnverified", err)
			}
		})
	}

	t.Run("token recusado pelo provedor", func(t *testing.T) {
		p.RejectTokens = true
		defer func() { p.RejectTokens = false }()
		c := NewChallenge()
		if _, err := g.Exchange(ctx, "qualquer", c); err == nil || !strings.Contains(err.Error(), "rejeitado") {
			t.Errorf("erro = %v", err)
		}
	})
}

func TestVerifyIDTokenChecksSignatureAndAudience(t *testing.T) {
	g, p := newTestGoogle(t)
	ctx := context.Background()
	c := NewChallenge()
	user := FakeUser{Sub: "g-1", Email: "a@b.c", EmailVerified: true, Nonce: c.Nonce}

	if _, err := g.verifyIDToken(ctx, p.sign(user), c.Nonce); err != nil {
		t.Fatalf("token válido rejeitado: %v", err)
	}
	if _, err := g.verifyIDToken(ctx, p.SignForeign(user), c.Nonce); err == nil {
		t.Error("assinatura de outra chave foi aceita")
	}

	parts := tokenParts(p.sign(user))
	tampered := parts[0] + "." + parts[1] + "x." + parts[2]
	if _, err := g.verifyIDToken(ctx, tampered, c.Nonce); err == nil {
		t.Error("corpo adulterado foi aceito")
	}
	if _, err := g.verifyIDToken(ctx, "nao.e.jwt", c.Nonce); err == nil {
		t.Error("lixo foi aceito")
	}

	otherAudience := NewGoogle("outro-cliente", "s", "http://x", p.Endpoints(), p.Server.Client())
	if _, err := otherAudience.verifyIDToken(ctx, p.sign(user), c.Nonce); err == nil {
		t.Error("audiência de outro cliente foi aceita")
	}

	// Relógio do servidor muito à frente: token vencido.
	g.now = func() time.Time { return time.Now().Add(3 * time.Hour) }
	if _, err := g.verifyIDToken(ctx, p.sign(user), c.Nonce); err == nil {
		t.Error("token vencido foi aceito")
	}
}

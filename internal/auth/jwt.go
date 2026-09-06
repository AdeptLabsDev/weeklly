package auth

import (
	"context"
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"strings"
	"sync"
	"time"
)

// clockSkew é a tolerância para relógios desalinhados entre nós e o Google.
const clockSkew = 2 * time.Minute

type claims struct {
	Issuer        string   `json:"iss"`
	Subject       string   `json:"sub"`
	Audience      audience `json:"aud"`
	Expires       int64    `json:"exp"`
	IssuedAt      int64    `json:"iat"`
	Nonce         string   `json:"nonce"`
	Email         string   `json:"email"`
	EmailVerified bool     `json:"email_verified"`
	Name          string   `json:"name"`
	Picture       string   `json:"picture"`
	Sub           string   `json:"-"`
}

// audience aceita string ou lista, como a especificação permite.
type audience []string

// UnmarshalJSON lê "aud" tanto como string quanto como lista.
func (a *audience) UnmarshalJSON(b []byte) error {
	var one string
	if err := json.Unmarshal(b, &one); err == nil {
		*a = audience{one}
		return nil
	}
	var many []string
	if err := json.Unmarshal(b, &many); err != nil {
		return err
	}
	*a = many
	return nil
}

func (a audience) contains(s string) bool {
	for _, v := range a {
		if v == s {
			return true
		}
	}
	return false
}

// verifyIDToken confere assinatura (RS256 contra as chaves publicadas),
// emissor, audiência, validade e nonce. Qualquer falha é erro: nunca se
// aproveita parte de um token inválido.
func (g *Google) verifyIDToken(ctx context.Context, raw, nonce string) (claims, error) {
	parts := strings.Split(raw, ".")
	if len(parts) != 3 {
		return claims{}, errors.New("id_token malformado")
	}
	headerJSON, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return claims{}, errors.New("id_token: cabeçalho ilegível")
	}
	var header struct {
		Alg string `json:"alg"`
		Kid string `json:"kid"`
	}
	if err := json.Unmarshal(headerJSON, &header); err != nil || header.Alg != "RS256" || header.Kid == "" {
		return claims{}, errors.New("id_token: algoritmo ou chave não suportados")
	}

	key, err := g.keys.get(ctx, header.Kid)
	if err != nil {
		return claims{}, err
	}
	signature, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return claims{}, errors.New("id_token: assinatura ilegível")
	}
	digest := sha256.Sum256([]byte(parts[0] + "." + parts[1]))
	if err := rsa.VerifyPKCS1v15(key, crypto.SHA256, digest[:], signature); err != nil {
		return claims{}, errors.New("id_token: assinatura inválida")
	}

	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return claims{}, errors.New("id_token: corpo ilegível")
	}
	var c claims
	if err := json.Unmarshal(payload, &c); err != nil {
		return claims{}, errors.New("id_token: corpo inválido")
	}
	c.Sub = c.Subject

	now := g.now()
	switch {
	case c.Issuer != g.endpoints.Issuer && c.Issuer != strings.TrimPrefix(g.endpoints.Issuer, "https://"):
		return claims{}, fmt.Errorf("id_token: emissor inesperado %q", c.Issuer)
	case !c.Audience.contains(g.clientID):
		return claims{}, errors.New("id_token: audiência inesperada")
	case c.Expires == 0 || now.After(time.Unix(c.Expires, 0).Add(clockSkew)):
		return claims{}, errors.New("id_token: vencido")
	case c.IssuedAt != 0 && time.Unix(c.IssuedAt, 0).After(now.Add(clockSkew)):
		return claims{}, errors.New("id_token: emitido no futuro")
	case c.Nonce == "" || c.Nonce != nonce:
		return claims{}, errors.New("id_token: nonce não confere")
	case c.Sub == "":
		return claims{}, errors.New("id_token: sem sub")
	}
	return c, nil
}

// keySet guarda as chaves públicas do provedor e busca de novo quando aparece
// um kid desconhecido (o Google gira chaves), no máximo uma vez por minuto.
type keySet struct {
	url    string
	client *http.Client

	mu        sync.Mutex
	keys      map[string]*rsa.PublicKey
	fetchedAt time.Time
}

func newKeySet(url string, client *http.Client) *keySet {
	return &keySet{url: url, client: client, keys: map[string]*rsa.PublicKey{}}
}

func (k *keySet) get(ctx context.Context, kid string) (*rsa.PublicKey, error) {
	k.mu.Lock()
	defer k.mu.Unlock()
	if key, ok := k.keys[kid]; ok {
		return key, nil
	}
	if time.Since(k.fetchedAt) < time.Minute && k.fetchedAt.After(time.Time{}) {
		return nil, errors.New("id_token: chave desconhecida")
	}
	if err := k.refresh(ctx); err != nil {
		return nil, err
	}
	if key, ok := k.keys[kid]; ok {
		return key, nil
	}
	return nil, errors.New("id_token: chave desconhecida")
}

func (k *keySet) refresh(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, k.url, nil)
	if err != nil {
		return err
	}
	resp, err := k.client.Do(req)
	if err != nil {
		return fmt.Errorf("buscando chaves do provedor: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("buscando chaves do provedor: %s", resp.Status)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return err
	}
	var doc struct {
		Keys []struct {
			Kty string `json:"kty"`
			Kid string `json:"kid"`
			N   string `json:"n"`
			E   string `json:"e"`
		} `json:"keys"`
	}
	if err := json.Unmarshal(body, &doc); err != nil {
		return fmt.Errorf("chaves do provedor ilegíveis: %w", err)
	}
	keys := map[string]*rsa.PublicKey{}
	for _, jwk := range doc.Keys {
		if jwk.Kty != "RSA" || jwk.Kid == "" {
			continue
		}
		n, err := base64.RawURLEncoding.DecodeString(jwk.N)
		if err != nil {
			continue
		}
		e, err := base64.RawURLEncoding.DecodeString(jwk.E)
		if err != nil {
			continue
		}
		keys[jwk.Kid] = &rsa.PublicKey{N: new(big.Int).SetBytes(n), E: int(new(big.Int).SetBytes(e).Int64())}
	}
	if len(keys) == 0 {
		return errors.New("provedor sem chaves RSA")
	}
	k.keys = keys
	k.fetchedAt = time.Now()
	return nil
}

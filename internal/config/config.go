// Package config lê a configuração do processo a partir do ambiente.
//
// Os padrões são os seguros para produção: modo production e escuta apenas em
// localhost. Desenvolvimento é sempre uma escolha explícita (WEEKLLY_ENV).
package config

import (
	"fmt"
	"net"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// Env identifica o modo de execução.
type Env string

// Modos de execução aceitos em WEEKLLY_ENV.
const (
	Development Env = "development"
	Production  Env = "production"
)

// Config é a configuração completa do processo.
type Config struct {
	// Env muda apenas ergonomia: logs legíveis e templates lidos do disco.
	// Nunca relaxa segurança.
	Env Env
	// Addr é o endereço de escuta, no formato host:porta.
	Addr string
	// DBPath é o caminho do arquivo SQLite. O diretório é criado se não existir.
	DBPath string
	// Timezone decide qual dia da semana é "hoje" enquanto o fuso do usuário
	// não é conhecido (o fuso do navegador entra na Fase 1).
	Timezone *time.Location
	// BaseURL é a origem pública do site (esquema e host, sem caminho), usada
	// para login, canonical e sitemap. Em development cai em http://Addr.
	BaseURL string
	// SearchIndexing libera a descoberta pública apenas na produção oficial.
	SearchIndexing bool
	// Google são as credenciais do login com o Google. Vazias desligam o login.
	Google Google
}

// Google são as credenciais OAuth criadas no Google Cloud.
type Google struct {
	ClientID     string
	ClientSecret string
}

// Configured informa se o login com o Google pode funcionar.
func (g Google) Configured() bool { return g.ClientID != "" && g.ClientSecret != "" }

const (
	defaultAddr     = "127.0.0.1:8080"
	defaultTimezone = "America/Sao_Paulo"
)

// Load monta a configuração a partir de getenv (normalmente os.Getenv).
// Receber a função em vez de ler o ambiente direto mantém o pacote testável.
func Load(getenv func(string) string) (Config, error) {
	cfg := Config{
		Env:    Production,
		Addr:   defaultAddr,
		DBPath: filepath.Join("data", "weeklly.db"),
	}

	switch v := Env(getenv("WEEKLLY_ENV")); v {
	case "":
	case Development, Production:
		cfg.Env = v
	default:
		return Config{}, fmt.Errorf("WEEKLLY_ENV=%q: use %q ou %q", v, Development, Production)
	}

	if v := getenv("WEEKLLY_ADDR"); v != "" {
		cfg.Addr = v
	}
	host, port, err := net.SplitHostPort(cfg.Addr)
	if err != nil {
		return Config{}, fmt.Errorf("WEEKLLY_ADDR=%q: %w", cfg.Addr, err)
	}

	if v := getenv("WEEKLLY_DB_PATH"); v != "" {
		cfg.DBPath = v
	}

	tz := getenv("WEEKLLY_TIMEZONE")
	if tz == "" {
		tz = defaultTimezone
	}
	loc, err := time.LoadLocation(tz)
	if err != nil {
		return Config{}, fmt.Errorf("WEEKLLY_TIMEZONE=%q: %w", tz, err)
	}
	cfg.Timezone = loc

	cfg.Google = Google{
		ClientID:     getenv("WEEKLLY_GOOGLE_CLIENT_ID"),
		ClientSecret: getenv("WEEKLLY_GOOGLE_CLIENT_SECRET"),
	}
	if cfg.Google.ClientID != "" || cfg.Google.ClientSecret != "" {
		if !cfg.Google.Configured() {
			return Config{}, fmt.Errorf("WEEKLLY_GOOGLE_CLIENT_ID e WEEKLLY_GOOGLE_CLIENT_SECRET precisam vir juntos")
		}
	}

	cfg.BaseURL, err = baseURL(getenv("WEEKLLY_BASE_URL"), cfg.Env, host, port)
	if err != nil {
		return Config{}, err
	}
	if cfg.Google.Configured() && cfg.Env == Production && !strings.HasPrefix(cfg.BaseURL, "https://") {
		return Config{}, fmt.Errorf("WEEKLLY_BASE_URL=%q: o login com o Google em produção exige https", cfg.BaseURL)
	}
	if raw := getenv("WEEKLLY_SEARCH_INDEXING"); raw != "" {
		cfg.SearchIndexing, err = strconv.ParseBool(raw)
		if err != nil {
			return Config{}, fmt.Errorf("WEEKLLY_SEARCH_INDEXING: use true ou false")
		}
	}
	if cfg.SearchIndexing && (cfg.Env != Production || !strings.HasPrefix(cfg.BaseURL, "https://")) {
		return Config{}, fmt.Errorf("WEEKLLY_SEARCH_INDEXING=true exige WEEKLLY_ENV=production e WEEKLLY_BASE_URL com https")
	}

	return cfg, nil
}

// baseURL valida a origem pública ou deduz uma em development.
func baseURL(raw string, env Env, host, port string) (string, error) {
	if raw == "" {
		if env == Production {
			return "", nil // sem origem, login e indexação continuam desligados
		}
		if host == "" || host == "0.0.0.0" || host == "::" {
			host = "127.0.0.1"
		}
		return "http://" + net.JoinHostPort(host, port), nil
	}
	u, err := url.Parse(raw)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Hostname() == "" || u.User != nil || (u.Path != "" && u.Path != "/") || u.RawQuery != "" || u.ForceQuery || u.Fragment != "" || strings.ContainsAny(raw, "?#") {
		return "", fmt.Errorf("WEEKLLY_BASE_URL=%q: use só esquema e host, como https://weeklly.app", raw)
	}
	return strings.TrimSuffix(raw, "/"), nil
}

// IsDev informa se o processo roda em modo development.
func (c Config) IsDev() bool { return c.Env == Development }

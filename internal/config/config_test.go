package config

import (
	"strings"
	"testing"

	_ "time/tzdata"
)

func env(vars map[string]string) func(string) string {
	return func(k string) string { return vars[k] }
}

func TestLoadDefaultsAreProductionSafe(t *testing.T) {
	cfg, err := Load(env(nil))
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Env != Production {
		t.Errorf("Env = %q, quero %q", cfg.Env, Production)
	}
	if cfg.Addr != "127.0.0.1:8080" {
		t.Errorf("Addr = %q, quero escuta só em localhost", cfg.Addr)
	}
	if cfg.Timezone.String() != "America/Sao_Paulo" {
		t.Errorf("Timezone = %q", cfg.Timezone)
	}
	if cfg.IsDev() {
		t.Error("IsDev() = true por padrão")
	}
}

func TestLoadReadsEveryVariable(t *testing.T) {
	cfg, err := Load(env(map[string]string{
		"WEEKLLY_ENV":      "development",
		"WEEKLLY_ADDR":     ":9000",
		"WEEKLLY_DB_PATH":  "/data/x.db",
		"WEEKLLY_TIMEZONE": "Europe/Lisbon",
	}))
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if !cfg.IsDev() || cfg.Addr != ":9000" || cfg.DBPath != "/data/x.db" || cfg.Timezone.String() != "Europe/Lisbon" {
		t.Errorf("Config = %+v", cfg)
	}
}

func TestBaseURLAndGoogle(t *testing.T) {
	dev, err := Load(env(map[string]string{"WEEKLLY_ENV": "development", "WEEKLLY_ADDR": ":9000"}))
	if err != nil {
		t.Fatal(err)
	}
	if dev.BaseURL != "http://127.0.0.1:9000" {
		t.Errorf("BaseURL deduzida = %q", dev.BaseURL)
	}
	if dev.Google.Configured() {
		t.Error("Google configurado sem credenciais")
	}

	prod, err := Load(env(map[string]string{
		"WEEKLLY_BASE_URL":             "https://weeklly.app/",
		"WEEKLLY_GOOGLE_CLIENT_ID":     "id",
		"WEEKLLY_GOOGLE_CLIENT_SECRET": "secret",
	}))
	if err != nil {
		t.Fatal(err)
	}
	if prod.BaseURL != "https://weeklly.app" || !prod.Google.Configured() {
		t.Errorf("Config = %+v", prod)
	}
}

func TestLoadRejectsInvalidValues(t *testing.T) {
	cases := map[string]map[string]string{
		"env desconhecido":             {"WEEKLLY_ENV": "staging"},
		"addr sem porta":               {"WEEKLLY_ADDR": "localhost"},
		"fuso inexistente":             {"WEEKLLY_TIMEZONE": "Marte/Olympus"},
		"base url com caminho":         {"WEEKLLY_BASE_URL": "https://weeklly.app/app"},
		"base url sem esquema":         {"WEEKLLY_BASE_URL": "weeklly.app"},
		"credencial do google sozinha": {"WEEKLLY_GOOGLE_CLIENT_ID": "id"},
		"google em produção sem https": {"WEEKLLY_GOOGLE_CLIENT_ID": "id", "WEEKLLY_GOOGLE_CLIENT_SECRET": "s", "WEEKLLY_BASE_URL": "http://weeklly.app"},
	}
	for name, vars := range cases {
		t.Run(name, func(t *testing.T) {
			_, err := Load(env(vars))
			if err == nil {
				t.Fatal("Load aceitou valor inválido")
			}
			if !strings.Contains(err.Error(), "WEEKLLY_") {
				t.Errorf("erro %q não cita a variável", err)
			}
		})
	}
}

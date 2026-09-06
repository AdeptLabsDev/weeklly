package i18n

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	_ "time/tzdata"

	"github.com/AdeptLabsDev/weeklly/internal/week"
)

func TestCatalogsHaveTheSameKeys(t *testing.T) {
	for k := range pt {
		if _, ok := en[k]; !ok {
			t.Errorf("falta em inglês: %s", k)
		}
	}
	for k := range en {
		if _, ok := pt[k]; !ok {
			t.Errorf("falta em português: %s", k)
		}
	}
	for k, v := range pt {
		if strings.Count(v, "%") != strings.Count(en[k], "%") {
			t.Errorf("%s: número de argumentos diferente entre idiomas", k)
		}
		if strings.TrimSpace(v) == "" || strings.TrimSpace(en[k]) == "" {
			t.Errorf("%s: texto vazio", k)
		}
	}
}

func TestParseAndNegotiate(t *testing.T) {
	cases := map[[2]string]Lang{
		{"", ""}:                           PT,
		{"en", ""}:                         EN,
		{"pt-BR", "en-US,en;q=0.9"}:        PT,
		{"", "en-US,en;q=0.9,pt;q=0.8"}:    EN,
		{"", "pt-BR,pt;q=0.9,en;q=0.8"}:    PT,
		{"", "fr-FR,fr;q=0.9,en-GB;q=0.7"}: EN,
		{"", "fr-FR,fr;q=0.9"}:             PT,
		{"klingon", "en"}:                  EN,
	}
	for in, want := range cases {
		if got := Negotiate(in[0], in[1]); got != want {
			t.Errorf("Negotiate(%q, %q) = %q, quero %q", in[0], in[1], got, want)
		}
	}
}

func TestTranslations(t *testing.T) {
	p, e := L(PT), L(EN)
	if p.T("nav.newWeek") != "Nova semana" || e.T("nav.newWeek") != "New week" {
		t.Error("nav.newWeek")
	}
	if p.T("board.todayIs", p.WeekdayLong(week.Saturday)) != "Hoje é sábado." || e.T("board.todayIs", e.WeekdayLong(week.Saturday)) != "Today is Saturday." {
		t.Error("board.todayIs")
	}
	if p.Weekday(week.Monday) != "Segunda" || e.WeekdayShort(week.Sunday) != "sun" || p.Weekday(week.Weekday(9)) != "" {
		t.Error("nomes dos dias")
	}
	if p.T("chave.inexistente") != "chave.inexistente" {
		t.Error("chave desconhecida deveria voltar como está")
	}
	if p.Error(week.ErrEmptyName) != "Dê um nome para a semana." || e.Error(week.ErrInvalidTime) != "Use a time like 09:30." {
		t.Error("erros de domínio")
	}
	if !strings.Contains(e.Error(week.ErrDayFull), "100") {
		t.Errorf("dayFull sem o limite: %q", e.Error(week.ErrDayFull))
	}
}

func TestUpdated(t *testing.T) {
	loc, _ := time.LoadLocation("America/Sao_Paulo")
	now := time.Date(2026, time.September, 9, 15, 0, 0, 0, time.UTC)
	cases := map[time.Time][2]string{
		now.Add(-time.Hour):                                       {"editada hoje", "edited today"},
		now.Add(-13 * time.Hour):                                  {"editada ontem", "edited yesterday"},
		now.Add(-3 * 24 * time.Hour):                              {"editada há 3 dias", "edited 3 days ago"},
		time.Date(2026, time.August, 12, 12, 0, 0, 0, time.UTC):   {"editada em 12 de agosto", "edited on 12 August"},
		time.Date(2025, time.December, 31, 12, 0, 0, 0, time.UTC): {"editada em 31 de dezembro de 2025", "edited on 31 December 2025"},
	}
	for at, want := range cases {
		if got := L(PT).Updated(at, now, loc); got != want[0] {
			t.Errorf("pt %s = %q, quero %q", at, got, want[0])
		}
		if got := L(EN).Updated(at, now, loc); got != want[1] {
			t.Errorf("en %s = %q, quero %q", at, got, want[1])
		}
	}
}

func TestClientJSON(t *testing.T) {
	var m map[string]string
	if err := json.Unmarshal([]byte(L(EN).ClientJSON()), &m); err != nil {
		t.Fatal(err)
	}
	if m["saved"] != "Saved" || m["noTime"] != "No time" {
		t.Errorf("ClientJSON = %v", m)
	}
	for k := range m {
		if strings.HasPrefix(k, "js.") {
			t.Errorf("prefixo não removido: %s", k)
		}
	}
}

func TestLanguages(t *testing.T) {
	for _, lang := range All {
		l := L(lang)
		opts := l.Languages()
		if len(opts) != len(All) {
			t.Fatalf("%s: %d idiomas, quero %d", lang, len(opts), len(All))
		}
		current := 0
		for _, o := range opts {
			if o.Name == "" || o.Name == "lang."+string(o.Tag) {
				t.Errorf("%s: idioma %s sem nome no catálogo", lang, o.Tag)
			}
			if o.Current {
				current++
				if o.Tag != lang {
					t.Errorf("%s: marcado como atual %s", lang, o.Tag)
				}
			}
		}
		if current != 1 {
			t.Errorf("%s: %d idiomas marcados como atual", lang, current)
		}
	}
	// O nome de cada idioma é o mesmo em qualquer catálogo: é o nome no
	// próprio idioma.
	if L(PT).T("lang.en") != L(EN).T("lang.en") || L(PT).T("lang.pt-BR") != L(EN).T("lang.pt-BR") {
		t.Error("nomes dos idiomas diferem entre catálogos")
	}
}

// Package i18n guarda os textos da interface em português do Brasil e em
// inglês, e escolhe o idioma de cada requisição (cookie, senão Accept-Language,
// senão português).
//
// Todo texto visível vem daqui: templates, erros de domínio, rótulos de data e
// as poucas frases do script. Os catálogos são mapas Go; o teste garante que
// os dois têm exatamente as mesmas chaves.
package i18n

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/AdeptLabsDev/weeklly/internal/week"
)

// Lang é um idioma suportado, no formato de etiqueta BCP 47.
type Lang string

// Idiomas suportados. PT é o padrão.
const (
	PT Lang = "pt-BR"
	EN Lang = "en"
)

// All lista os idiomas, na ordem em que aparecem nas configurações.
var All = []Lang{PT, EN}

// Parse reconhece um idioma vindo de cookie ou formulário.
func Parse(s string) (Lang, bool) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "pt-br", "pt":
		return PT, true
	case "en", "en-us", "en-gb":
		return EN, true
	}
	return "", false
}

// Negotiate escolhe o idioma: cookie explícito, senão a primeira preferência
// reconhecida do Accept-Language, senão português.
func Negotiate(cookie, acceptLanguage string) Lang {
	if lang, ok := Parse(cookie); ok {
		return lang
	}
	for _, part := range strings.Split(acceptLanguage, ",") {
		tag, _, _ := strings.Cut(strings.TrimSpace(part), ";")
		if lang, ok := Parse(tag); ok {
			return lang
		}
		if base, _, found := strings.Cut(tag, "-"); found {
			if lang, ok := Parse(base); ok {
				return lang
			}
		}
	}
	return PT
}

// Locale traduz para um idioma.
type Locale struct {
	lang Lang
	m    map[string]string
}

// L devolve o Locale de um idioma.
func L(lang Lang) Locale {
	if lang == EN {
		return Locale{lang: EN, m: en}
	}
	return Locale{lang: PT, m: pt}
}

// Lang devolve o idioma do Locale.
func (l Locale) Lang() Lang { return l.lang }

// IsZero informa se o Locale não foi inicializado.
func (l Locale) IsZero() bool { return l.m == nil }

// T traduz a chave; com args, formata com fmt.Sprintf. Chave desconhecida
// devolve a própria chave, para aparecer na tela em vez de sumir.
func (l Locale) T(key string, args ...any) string {
	s, ok := l.m[key]
	if !ok {
		return key
	}
	if len(args) == 0 {
		return s
	}
	return fmt.Sprintf(s, args...)
}

// Weekday devolve o nome curto com inicial maiúscula: "Segunda", "Monday".
func (l Locale) Weekday(d week.Weekday) string { return l.pick("weekday", d) }

// WeekdayLong devolve o nome completo em minúsculas: "segunda-feira", "monday".
func (l Locale) WeekdayLong(d week.Weekday) string { return l.pick("weekday.long", d) }

// WeekdayShort devolve a abreviação de três letras: "seg", "mon".
func (l Locale) WeekdayShort(d week.Weekday) string { return l.pick("weekday.short", d) }

func (l Locale) pick(prefix string, d week.Weekday) string {
	if !d.Valid() {
		return ""
	}
	return l.T(fmt.Sprintf("%s.%d", prefix, int(d)))
}

// Month devolve o nome do mês em minúsculas.
func (l Locale) Month(m time.Month) string {
	if m < time.January || m > time.December {
		return ""
	}
	return l.T(fmt.Sprintf("month.%d", int(m)))
}

// Updated descreve quando algo mudou pela última vez, em dias: "editada
// hoje", "edited yesterday", "editada há 3 dias", "edited on 12 August".
func (l Locale) Updated(at, now time.Time, loc *time.Location) string {
	a, n := at.In(loc), now.In(loc)
	aDay := time.Date(a.Year(), a.Month(), a.Day(), 0, 0, 0, 0, loc)
	nDay := time.Date(n.Year(), n.Month(), n.Day(), 0, 0, 0, 0, loc)
	days := int(nDay.Sub(aDay).Hours() / 24)
	switch {
	case days <= 0:
		return l.T("updated.today")
	case days == 1:
		return l.T("updated.yesterday")
	case days < 7:
		return l.T("updated.days", days)
	case a.Year() == n.Year():
		return l.T("updated.date", a.Day(), l.Month(a.Month()))
	default:
		return l.T("updated.dateYear", a.Day(), l.Month(a.Month()), a.Year())
	}
}

// Error traduz um erro de domínio em frase para a interface. Erro
// desconhecido vira a frase genérica de falha.
func (l Locale) Error(err error) string {
	switch {
	case errors.Is(err, week.ErrEmptyName):
		return l.T("err.emptyName")
	case errors.Is(err, week.ErrNameTooLong):
		return l.T("err.nameTooLong", week.MaxNameLength)
	case errors.Is(err, week.ErrInvalidName):
		return l.T("err.invalidName")
	case errors.Is(err, week.ErrEmptyTitle):
		return l.T("err.emptyTitle")
	case errors.Is(err, week.ErrTitleTooLong):
		return l.T("err.titleTooLong", week.MaxTitleLength)
	case errors.Is(err, week.ErrInvalidTime):
		return l.T("err.invalidTime")
	case errors.Is(err, week.ErrDayFull):
		return l.T("err.dayFull", week.MaxTasksPerDay)
	default:
		return l.T("err.generic")
	}
}

// ClientJSON devolve, em JSON, as frases que o script usa (chaves "js.*",
// sem o prefixo). Vai num atributo data-* do body.
func (l Locale) ClientJSON() string {
	out := map[string]string{}
	for k, v := range l.m {
		if rest, ok := strings.CutPrefix(k, "js."); ok {
			out[rest] = v
		}
	}
	b, _ := json.Marshal(out)
	return string(b)
}

// Option é um idioma na lista das configurações: etiqueta, nome no próprio
// idioma e se é o escolhido.
type Option struct {
	Tag     Lang
	Name    string
	Current bool
}

// Languages lista os idiomas para o seletor, cada um com o nome no próprio
// idioma ("Português", "English"), o atual marcado.
func (l Locale) Languages() []Option {
	out := make([]Option, 0, len(All))
	for _, lang := range All {
		out = append(out, Option{Tag: lang, Name: l.T("lang." + string(lang)), Current: lang == l.lang})
	}
	return out
}

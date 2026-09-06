package week

import (
	"errors"
	"strings"
	"testing"
	"time"

	_ "time/tzdata"
)

func TestWeekdayOrder(t *testing.T) {
	all := All()
	if all[0] != Monday || all[Days-1] != Sunday || len(all) != 7 {
		t.Fatalf("All() = %v", all)
	}
	for i, d := range all {
		if int(d) != i || !d.Valid() {
			t.Errorf("Weekday(%d) fora de ordem ou inválido", i)
		}
	}
	for _, bad := range []Weekday{-1, 7, 100} {
		if bad.Valid() {
			t.Errorf("Weekday(%d) deveria ser inválido", bad)
		}
	}
}

func TestFromTimeMapsSundayToTheEnd(t *testing.T) {
	cases := map[time.Weekday]Weekday{
		time.Monday: Monday, time.Wednesday: Wednesday, time.Saturday: Saturday, time.Sunday: Sunday,
	}
	for in, want := range cases {
		if got := FromTime(in); got != want {
			t.Errorf("FromTime(%s) = %d, quero %d", in, got, want)
		}
	}
}

func TestTodayDependsOnTimezone(t *testing.T) {
	// 01:30 UTC de terça, 8 de setembro de 2026, ainda é segunda em São Paulo.
	now := time.Date(2026, time.September, 8, 1, 30, 0, 0, time.UTC)
	cases := map[string]Weekday{
		"America/Sao_Paulo": Monday,
		"UTC":               Tuesday,
		"Asia/Tokyo":        Tuesday,
	}
	for tz, want := range cases {
		loc, err := time.LoadLocation(tz)
		if err != nil {
			t.Fatal(err)
		}
		if got := Today(now, loc); got != want {
			t.Errorf("Today em %s = %d, quero %d", tz, got, want)
		}
	}
}

func TestCleanName(t *testing.T) {
	ok := map[string]string{
		"Semana padrão":          "Semana padrão",
		"  Semana   de provas  ": "Semana de provas",
		"Férias\tem\njulho":      "Férias em julho",
		"x":                      "x",
	}
	for in, want := range ok {
		got, err := CleanName(in)
		if err != nil || got != want {
			t.Errorf("CleanName(%q) = %q, %v; quero %q", in, got, err, want)
		}
	}

	if _, err := CleanName("   "); !errors.Is(err, ErrEmptyName) {
		t.Errorf("vazio: %v", err)
	}
	if _, err := CleanName(strings.Repeat("á", MaxNameLength+1)); !errors.Is(err, ErrNameTooLong) {
		t.Errorf("longo: %v", err)
	}
	if got, err := CleanName(strings.Repeat("á", MaxNameLength)); err != nil || got == "" {
		t.Errorf("no limite: %v", err)
	}
	if _, err := CleanName("a\x00b"); !errors.Is(err, ErrInvalidName) {
		t.Errorf("controle: %v", err)
	}
}

func TestCleanTitle(t *testing.T) {
	if got, err := CleanTitle("  Treino   7h "); err != nil || got != "Treino 7h" {
		t.Errorf("CleanTitle = %q, %v", got, err)
	}
	if _, err := CleanTitle(" "); !errors.Is(err, ErrEmptyTitle) {
		t.Errorf("vazio: %v", err)
	}
	if _, err := CleanTitle(strings.Repeat("x", MaxTitleLength+1)); !errors.Is(err, ErrTitleTooLong) {
		t.Errorf("longo: %v", err)
	}
}

func TestCleanTime(t *testing.T) {
	ok := map[string]string{"": "", "  ": "", "09:30": "09:30", "9:05": "09:05", "23:59": "23:59", "00:00": "00:00"}
	for in, want := range ok {
		if got, err := CleanTime(in); err != nil || got != want {
			t.Errorf("CleanTime(%q) = %q, %v; quero %q", in, got, err, want)
		}
	}
	for _, bad := range []string{"24:00", "9:5", "09:60", "930", "9h30", "-1:00", "ab:cd", "09:30:00"} {
		if _, err := CleanTime(bad); !errors.Is(err, ErrInvalidTime) {
			t.Errorf("CleanTime(%q) aceitou: %v", bad, err)
		}
	}
}

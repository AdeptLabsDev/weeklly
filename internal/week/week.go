// Package week contém o modelo do produto: a semana é um espaço de
// planejamento com nome, sete dias de segunda a domingo, sem datas (D12).
//
// O calendário entra em um único ponto, Today, para saber qual dia da semana
// é hoje no fuso do usuário (invariante 7). Nada mais aqui depende de data.
package week

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
)

// Days é o tamanho fixo de uma semana. Não é configurável: é o produto.
const Days = 7

// Weekday é a posição de um dia na semana: 0 é segunda, 6 é domingo.
type Weekday int

// Os sete dias, na ordem do produto.
const (
	Monday Weekday = iota
	Tuesday
	Wednesday
	Thursday
	Friday
	Saturday
	Sunday
)

// All devolve os sete dias em ordem.
func All() [Days]Weekday {
	var out [Days]Weekday
	for i := range out {
		out[i] = Weekday(i)
	}
	return out
}

// Valid informa se d está entre segunda e domingo.
func (d Weekday) Valid() bool { return d >= Monday && d <= Sunday }

// FromTime converte a convenção do pacote time (domingo = 0) para a do
// produto (segunda = 0).
func FromTime(w time.Weekday) Weekday { return Weekday((int(w) + 6) % Days) }

// Today devolve o dia da semana de now no fuso loc. É a única porta de
// entrada do calendário no pacote.
func Today(now time.Time, loc *time.Location) Weekday {
	return FromTime(now.In(loc).Weekday())
}

// Week é um espaço de planejamento: um nome e, por dia, a lista de tarefas
// na ordem em que a pessoa as pôs. O ID é gerado por ids.New.
type Week struct {
	ID   string
	Name string
	Days [Days][]Task
}

// Task é um item do plano de um dia: um título, um horário opcional e se já
// foi feita. A ordem dentro do dia é manual (Position), como num quadro.
type Task struct {
	ID       string
	Weekday  Weekday
	Position int
	Title    string
	Time     string // "HH:MM" ou vazio
	Done     bool
}

// MaxTitleLength é o limite de caracteres (runas) do título de uma tarefa.
const MaxTitleLength = 200

// MaxTasksPerDay é o limite de tarefas num dia. Um plano com mais que isso
// não é um plano.
const MaxTasksPerDay = 100

// Erros de validação de tarefa. O texto para a interface vem do pacote i18n;
// estes são identificadores.
var (
	ErrEmptyTitle   = errors.New("título vazio")
	ErrTitleTooLong = fmt.Errorf("título com mais de %d caracteres", MaxTitleLength)
	ErrInvalidTime  = errors.New("horário inválido")
	ErrDayFull      = fmt.Errorf("dia com mais de %d tarefas", MaxTasksPerDay)
)

// CleanTitle normaliza e valida o título de uma tarefa, com as mesmas regras
// do nome da semana e um limite maior.
func CleanTitle(raw string) (string, error) {
	title := strings.Join(strings.Fields(raw), " ")
	if title == "" {
		return "", ErrEmptyTitle
	}
	if utf8.RuneCountInString(title) > MaxTitleLength {
		return "", ErrTitleTooLong
	}
	for _, r := range title {
		if unicode.IsControl(r) || r == utf8.RuneError {
			return "", ErrInvalidName
		}
	}
	return title, nil
}

// CleanTime aceita vazio (sem horário) ou um horário "H:MM"/"HH:MM" de
// 00:00 a 23:59 e devolve sempre "HH:MM".
func CleanTime(raw string) (string, error) {
	s := strings.TrimSpace(raw)
	if s == "" {
		return "", nil
	}
	h, m, ok := strings.Cut(s, ":")
	if !ok || len(h) == 0 || len(h) > 2 || len(m) != 2 {
		return "", ErrInvalidTime
	}
	hour, err1 := strconv.Atoi(h)
	minute, err2 := strconv.Atoi(m)
	if err1 != nil || err2 != nil || hour < 0 || hour > 23 || minute < 0 || minute > 59 {
		return "", ErrInvalidTime
	}
	return fmt.Sprintf("%02d:%02d", hour, minute), nil
}

// MaxNameLength é o limite de caracteres (runas) do nome de uma semana.
const MaxNameLength = 60

// Erros de validação do nome. O texto para a interface vem do pacote i18n.
var (
	ErrEmptyName   = errors.New("nome vazio")
	ErrNameTooLong = fmt.Errorf("nome com mais de %d caracteres", MaxNameLength)
	ErrInvalidName = errors.New("nome com caracteres inválidos")
)

// CleanName normaliza e valida o nome de uma semana: remove espaços nas
// pontas, colapsa espaços internos, exige de 1 a MaxNameLength caracteres e
// rejeita caracteres de controle.
func CleanName(raw string) (string, error) {
	name := strings.Join(strings.Fields(raw), " ")
	if name == "" {
		return "", ErrEmptyName
	}
	if utf8.RuneCountInString(name) > MaxNameLength {
		return "", ErrNameTooLong
	}
	for _, r := range name {
		if unicode.IsControl(r) || r == utf8.RuneError {
			return "", ErrInvalidName
		}
	}
	return name, nil
}

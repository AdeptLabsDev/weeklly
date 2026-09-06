package server

import (
	"fmt"
	"net/http"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/AdeptLabsDev/weeklly/internal/i18n"
	"github.com/AdeptLabsDev/weeklly/internal/store"
	"github.com/AdeptLabsDev/weeklly/internal/week"
)

// page é o que todo template recebe: o que a casca precisa (título, classe
// do body, tema, idioma, navegação) e os dados da página em Data.
type page struct {
	Title     string
	BodyClass string
	// Public usa a apresentação do produto, sem dados ou scripts do app.
	Public bool
	SEO    seoView
	// Theme é "dark" ou "light"; render preenche a partir do cookie.
	Theme string
	// Cursor é "system" ou "custom"; Accent é "mono" ou uma cor de accents.
	// Os dois vêm de cookies e viram atributos data-* no <html>.
	Cursor string
	Accent string
	// L traduz; render preenche a partir do cookie ou do Accept-Language.
	L    i18n.Locale
	Nav  navView
	Data any
	// NewWeek alimenta o diálogo "Nova semana", presente em toda página.
	NewWeek formView
	// Rename e Delete alimentam os diálogos da semana aberta (só no quadro).
	Rename *formView
	Delete *deleteView
}

// navView é a barra de navegação: quem está, quais semanas tem, qual está
// aberta, em que ordem prefere vê-las.
type navView struct {
	// User é nil para visitantes e usuários anônimos: só quem entrou com uma
	// conta tem foto na barra.
	User    *userView
	Weeks   []weekLink
	Current *weekLink
	Order   string
}

// prefOption é uma opção de preferência no menu de configurações.
type prefOption struct {
	Value   string
	Label   string
	Current bool
}

// Cursors lista os tipos de cursor para o menu: sistema e o do produto.
func (p page) Cursors() []prefOption {
	return []prefOption{
		{Value: cursorSystem, Label: p.L.T("cursor.system"), Current: p.Cursor == cursorSystem},
		{Value: cursorCustom, Label: p.L.T("cursor.custom"), Current: p.Cursor == cursorCustom},
	}
}

// Accents lista as cores de destaque para o menu, a atual marcada.
func (p page) Accents() []prefOption {
	out := make([]prefOption, 0, len(accents))
	for _, a := range accents {
		out = append(out, prefOption{Value: a, Label: p.L.T("accent." + a), Current: p.Accent == a})
	}
	return out
}

type userView struct {
	Name    string
	Email   string
	Picture string
	Initial string
}

type weekLink struct {
	ID      string
	Name    string
	URL     string
	Updated string
	// UpdatedAt em RFC 3339, para o script reordenar a lista sem recarregar.
	UpdatedAt string
	Current   bool
}

type hubView struct {
	Weeks []weekLink
	Order string
}

// formView serve o formulário de nome de semana, em página e em diálogo,
// para criar e para renomear.
type formView struct {
	L       i18n.Locale
	Heading string
	Lead    string
	Action  string
	Submit  string
	Cancel  string
	Name    string
	Error   string
}

type deleteView struct {
	L      i18n.Locale
	Action string
	Name   string
	Cancel string
}

type messageView struct {
	Heading  string
	Body     string
	LinkText string
	LinkURL  string
}

// boardView é o que o template do quadro recebe. Só strings e valores
// prontos: toda formatação acontece aqui, nunca no template.
type boardView struct {
	ID            string
	Name          string
	TodayLongName string
	Days          [week.Days]dayView
}

type dayView struct {
	Weekday  int
	Name     string
	LongName string
	Short    string
	IsToday  bool
	Tasks    []taskView
}

// taskView é uma tarefa pronta para o template, com o Locale porque a
// parcial também é renderizada sozinha, para o JSON.
type taskView struct {
	L     i18n.Locale
	ID    string
	Title string
	Time  string
	Done  bool
}

func newBoardView(l i18n.Locale, w week.Week, today week.Weekday) boardView {
	v := boardView{ID: w.ID, Name: w.Name, TodayLongName: l.WeekdayLong(today)}
	for i, d := range week.All() {
		day := dayView{
			Weekday:  int(d),
			Name:     l.Weekday(d),
			LongName: l.WeekdayLong(d),
			Short:    l.WeekdayShort(d),
			IsToday:  d == today,
		}
		for _, t := range w.Days[i] {
			day.Tasks = append(day.Tasks, newTaskView(l, t))
		}
		v.Days[i] = day
	}
	return v
}

func newTaskView(l i18n.Locale, t week.Task) taskView {
	return taskView{L: l, ID: t.ID, Title: t.Title, Time: t.Time, Done: t.Done}
}

func newWeekForm(l i18n.Locale) formView {
	return formView{
		L:       l,
		Heading: l.T("new.heading"),
		Lead:    l.T("new.lead"),
		Action:  "/semanas",
		Submit:  l.T("new.submit"),
		Cancel:  "/semanas",
	}
}

func renameForm(l i18n.Locale, w week.Week) formView {
	return formView{
		L:       l,
		Heading: l.T("rename.heading"),
		Lead:    l.T("rename.lead"),
		Action:  weekURL(w.ID) + "/renomear",
		Submit:  l.T("rename.submit"),
		Cancel:  weekURL(w.ID),
		Name:    w.Name,
	}
}

// nav monta a barra para a requisição. currentID é a semana aberta, ou vazio.
func (s *Server) nav(r *http.Request, currentID string) (navView, error) {
	v := currentVisitor(r)
	l := s.locale(r)
	if !v.ok {
		return navView{Order: string(store.OrderRecent)}, nil
	}
	nav := navView{Order: string(v.user.WeekOrder)}
	if !v.user.Anonymous() {
		nav.User = &userView{
			Name:    v.user.Name,
			Email:   v.user.Email,
			Picture: v.user.PictureURL,
			Initial: initial(v.user.Name, v.user.Email),
		}
	}
	list, err := s.opts.Store.ListWeeks(r.Context(), v.user.ID, v.user.WeekOrder)
	if err != nil {
		return navView{}, fmt.Errorf("listando semanas: %w", err)
	}
	now := s.now()
	for _, w := range list {
		link := weekLink{
			ID:        w.ID,
			Name:      w.Name,
			URL:       weekURL(w.ID),
			Updated:   l.Updated(w.UpdatedAt, now, s.opts.Config.Timezone),
			UpdatedAt: w.UpdatedAt.UTC().Format(time.RFC3339Nano),
			Current:   w.ID == currentID,
		}
		nav.Weeks = append(nav.Weeks, link)
		if link.Current {
			nav.Current = &nav.Weeks[len(nav.Weeks)-1]
		}
	}
	return nav, nil
}

func initial(name, email string) string {
	for _, s := range []string{name, email} {
		if r, size := utf8.DecodeRuneInString(strings.TrimSpace(s)); size > 0 && r != utf8.RuneError {
			return string(unicode.ToUpper(r))
		}
	}
	return "?"
}

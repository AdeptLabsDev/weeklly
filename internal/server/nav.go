package server

import (
	"fmt"
	"net/http"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/AdeptLabsDev/weeklly/internal/store"
	"github.com/AdeptLabsDev/weeklly/internal/week"
)

// page é o que todo template recebe: o que a casca precisa (título, classe
// do body, tema, navegação) e os dados da página em Data.
type page struct {
	Title     string
	BodyClass string
	// Theme é "dark" ou "light"; render preenche a partir do cookie.
	Theme string
	Nav   navView
	Data  any
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
	Heading string
	Lead    string
	Action  string
	Submit  string
	Cancel  string
	Name    string
	Error   string
}

type deleteView struct {
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

type taskView struct {
	ID    string
	Title string
	Time  string
	Done  bool
}

func newBoardView(w week.Week, today week.Weekday) boardView {
	v := boardView{ID: w.ID, Name: w.Name, TodayLongName: today.LongName()}
	for i, d := range week.All() {
		day := dayView{
			Weekday:  int(d),
			Name:     d.Name(),
			LongName: d.LongName(),
			Short:    d.Short(),
			IsToday:  d == today,
		}
		for _, t := range w.Days[i] {
			day.Tasks = append(day.Tasks, newTaskView(t))
		}
		v.Days[i] = day
	}
	return v
}

func newTaskView(t week.Task) taskView {
	return taskView{ID: t.ID, Title: t.Title, Time: t.Time, Done: t.Done}
}

func newWeekForm() formView {
	return formView{
		Heading: "Nova semana",
		Lead:    "Dê um nome que diga para que ela serve.",
		Action:  "/semanas",
		Submit:  "Criar semana",
		Cancel:  "/semanas",
	}
}

func renameForm(w week.Week) formView {
	return formView{
		Heading: "Renomear semana",
		Lead:    "O nome novo vale em todo lugar: barra, lista e título.",
		Action:  weekURL(w.ID) + "/renomear",
		Submit:  "Renomear",
		Cancel:  weekURL(w.ID),
		Name:    w.Name,
	}
}

// nav monta a barra para a requisição. currentID é a semana aberta, ou vazio.
func (s *Server) nav(r *http.Request, currentID string) (navView, error) {
	v := currentVisitor(r)
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
			Updated:   updatedLabel(w.UpdatedAt, now, s.opts.Config.Timezone),
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

var monthsPT = [...]string{"", "janeiro", "fevereiro", "março", "abril", "maio", "junho", "julho", "agosto", "setembro", "outubro", "novembro", "dezembro"}

// updatedLabel descreve quando a semana mudou pela última vez, na voz do
// produto: "editada hoje", "editada ontem", "editada há 3 dias", "editada em
// 12 de agosto".
func updatedLabel(at, now time.Time, loc *time.Location) string {
	a, n := at.In(loc), now.In(loc)
	aDay := time.Date(a.Year(), a.Month(), a.Day(), 0, 0, 0, 0, loc)
	nDay := time.Date(n.Year(), n.Month(), n.Day(), 0, 0, 0, 0, loc)
	days := int(nDay.Sub(aDay).Hours() / 24)
	switch {
	case days <= 0:
		return "editada hoje"
	case days == 1:
		return "editada ontem"
	case days < 7:
		return fmt.Sprintf("editada há %d dias", days)
	case a.Year() == n.Year():
		return fmt.Sprintf("editada em %d de %s", a.Day(), monthsPT[a.Month()])
	default:
		return fmt.Sprintf("editada em %d de %s de %d", a.Day(), monthsPT[a.Month()], a.Year())
	}
}

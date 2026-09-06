package store

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/AdeptLabsDev/weeklly/internal/week"
)

func TestWeeksBelongToTheirUser(t *testing.T) {
	s := openTest(t)
	ctx := context.Background()

	alice, err := s.CreateAnonymousUser(ctx)
	if err != nil {
		t.Fatal(err)
	}
	bob, err := s.CreateAnonymousUser(ctx)
	if err != nil {
		t.Fatal(err)
	}

	first, err := s.CreateWeek(ctx, alice.ID, "  Semana   padrão ")
	if err != nil {
		t.Fatal(err)
	}
	if first.Name != "Semana padrão" {
		t.Errorf("nome não normalizado: %q", first.Name)
	}
	if _, err := s.CreateWeek(ctx, alice.ID, "   "); !errors.Is(err, week.ErrEmptyName) {
		t.Errorf("nome vazio: %v", err)
	}
	second, err := s.CreateWeek(ctx, alice.ID, "Ano novo")
	if err != nil {
		t.Fatal(err)
	}

	got, err := s.Week(ctx, alice.ID, first.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Name != "Semana padrão" || got.ID != first.ID {
		t.Errorf("Week = %+v", got)
	}
	for i, tasks := range got.Days {
		if len(tasks) != 0 {
			t.Errorf("dia %d nasceu com tarefas", i)
		}
	}

	if _, err := s.Week(ctx, bob.ID, first.ID); !errors.Is(err, ErrNotFound) {
		t.Errorf("semana de outro usuário: %v, quero ErrNotFound", err)
	}
	if _, err := s.Week(ctx, alice.ID, "nao-e-um-id"); !errors.Is(err, ErrNotFound) {
		t.Errorf("id inválido: %v, quero ErrNotFound", err)
	}

	recent, err := s.ListWeeks(ctx, alice.ID, OrderRecent)
	if err != nil {
		t.Fatal(err)
	}
	if len(recent) != 2 || recent[0].ID != second.ID || recent[1].ID != first.ID {
		t.Errorf("ListWeeks recente = %+v, quero a mais recente primeiro", recent)
	}
	byName, err := s.ListWeeks(ctx, alice.ID, OrderName)
	if err != nil {
		t.Fatal(err)
	}
	if len(byName) != 2 || byName[0].Name != "Ano novo" || byName[1].Name != "Semana padrão" {
		t.Errorf("ListWeeks por nome = %+v", byName)
	}
	if other, _ := s.ListWeeks(ctx, bob.ID, OrderRecent); len(other) != 0 {
		t.Errorf("bob vê %d semanas", len(other))
	}

	// Ordem preferida fica no usuário.
	if err := s.SetWeekOrder(ctx, alice.ID, OrderName); err != nil {
		t.Fatal(err)
	}
	if u, _ := s.UserByID(ctx, alice.ID); u.WeekOrder != OrderName {
		t.Errorf("WeekOrder = %q", u.WeekOrder)
	}

	// Renomear, duplicar e apagar respeitam o dono.
	if err := s.RenameWeek(ctx, bob.ID, first.ID, "Roubada"); !errors.Is(err, ErrNotFound) {
		t.Errorf("renomear semana alheia: %v", err)
	}
	if err := s.RenameWeek(ctx, alice.ID, first.ID, " Semana base "); err != nil {
		t.Fatal(err)
	}
	if w, _ := s.Week(ctx, alice.ID, first.ID); w.Name != "Semana base" {
		t.Errorf("nome depois de renomear = %q", w.Name)
	}
	if _, err := s.DuplicateWeek(ctx, bob.ID, first.ID, "Cópia"); !errors.Is(err, ErrNotFound) {
		t.Errorf("duplicar semana alheia: %v", err)
	}
	if err := s.DeleteWeek(ctx, bob.ID, first.ID); !errors.Is(err, ErrNotFound) {
		t.Errorf("apagar semana alheia: %v", err)
	}
	if err := s.DeleteWeek(ctx, alice.ID, second.ID); err != nil {
		t.Fatal(err)
	}
	if list, _ := s.ListWeeks(ctx, alice.ID, OrderRecent); len(list) != 1 {
		t.Errorf("depois de apagar sobraram %d semanas", len(list))
	}
}

func TestTasksLiveInsideTheirDay(t *testing.T) {
	s := openTest(t)
	ctx := context.Background()
	alice, _ := s.CreateAnonymousUser(ctx)
	bob, _ := s.CreateAnonymousUser(ctx)
	w, err := s.CreateWeek(ctx, alice.ID, "Semana padrão")
	if err != nil {
		t.Fatal(err)
	}

	treino, err := s.AddTask(ctx, alice.ID, w.ID, week.Monday, "  Treino ", "7:00")
	if err != nil {
		t.Fatal(err)
	}
	if treino.Title != "Treino" || treino.Time != "07:00" || treino.Position != 0 || treino.Done {
		t.Errorf("tarefa = %+v", treino)
	}
	leitura, err := s.AddTask(ctx, alice.ID, w.ID, week.Monday, "Leitura", "")
	if err != nil {
		t.Fatal(err)
	}
	if leitura.Position != 1 || leitura.Time != "" {
		t.Errorf("segunda tarefa = %+v", leitura)
	}
	feira, err := s.AddTask(ctx, alice.ID, w.ID, week.Saturday, "Feira", "")
	if err != nil {
		t.Fatal(err)
	}

	if _, err := s.AddTask(ctx, alice.ID, w.ID, week.Monday, "", ""); !errors.Is(err, week.ErrEmptyTitle) {
		t.Errorf("título vazio: %v", err)
	}
	if _, err := s.AddTask(ctx, alice.ID, w.ID, week.Monday, "x", "25:00"); !errors.Is(err, week.ErrInvalidTime) {
		t.Errorf("horário inválido: %v", err)
	}
	if _, err := s.AddTask(ctx, alice.ID, w.ID, week.Weekday(7), "x", ""); !errors.Is(err, ErrNotFound) {
		t.Errorf("dia 7: %v", err)
	}
	if _, err := s.AddTask(ctx, bob.ID, w.ID, week.Monday, "Intruso", ""); !errors.Is(err, ErrNotFound) {
		t.Errorf("tarefa em semana alheia: %v", err)
	}

	loaded, err := s.Week(ctx, alice.ID, w.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(loaded.Days[week.Monday]) != 2 || loaded.Days[week.Monday][0].ID != treino.ID || loaded.Days[week.Monday][1].ID != leitura.ID {
		t.Errorf("segunda = %+v", loaded.Days[week.Monday])
	}
	if len(loaded.Days[week.Saturday]) != 1 || loaded.Days[week.Sunday] != nil {
		t.Errorf("fim de semana = %+v / %+v", loaded.Days[week.Saturday], loaded.Days[week.Sunday])
	}

	// Editar: título, horário, feita. Cada um separadamente e respeitando o dono.
	newTitle, newTime, done := "Treino na academia", "06:30", true
	updated, err := s.UpdateTask(ctx, alice.ID, treino.ID, TaskPatch{Title: &newTitle, Time: &newTime, Done: &done})
	if err != nil {
		t.Fatal(err)
	}
	if updated.Title != newTitle || updated.Time != newTime || !updated.Done {
		t.Errorf("depois do patch = %+v", updated)
	}
	empty := ""
	if u, err := s.UpdateTask(ctx, alice.ID, treino.ID, TaskPatch{Time: &empty}); err != nil || u.Time != "" {
		t.Errorf("tirar horário: %+v, %v", u, err)
	}
	if _, err := s.UpdateTask(ctx, bob.ID, treino.ID, TaskPatch{Done: &done}); !errors.Is(err, ErrNotFound) {
		t.Errorf("editar tarefa alheia: %v", err)
	}
	bad := "24:01"
	if _, err := s.UpdateTask(ctx, alice.ID, treino.ID, TaskPatch{Time: &bad}); !errors.Is(err, week.ErrInvalidTime) {
		t.Errorf("horário inválido no patch: %v", err)
	}

	// Duplicar copia as tarefas, sem o "feita".
	copyWeek, err := s.DuplicateWeek(ctx, alice.ID, w.ID, "Semana padrão (cópia)")
	if err != nil {
		t.Fatal(err)
	}
	if len(copyWeek.Days[week.Monday]) != 2 || copyWeek.Days[week.Monday][0].Title != newTitle || copyWeek.Days[week.Monday][0].Done {
		t.Errorf("cópia = %+v", copyWeek.Days[week.Monday])
	}
	if copyWeek.Days[week.Monday][0].ID == treino.ID {
		t.Error("a cópia reaproveitou o id da tarefa original")
	}

	// Mover dentro do dia e para outro dia mantém as posições contíguas.
	terceira, err := s.AddTask(ctx, alice.ID, w.ID, week.Monday, "Terceira", "")
	if err != nil {
		t.Fatal(err)
	}
	titles := func(d week.Weekday) []string {
		wk, err := s.Week(ctx, alice.ID, w.ID)
		if err != nil {
			t.Fatal(err)
		}
		var out []string
		for i, task := range wk.Days[d] {
			if task.Position != i {
				t.Errorf("dia %d: tarefa %q na posição %d, índice %d", d, task.Title, task.Position, i)
			}
			out = append(out, task.Title)
		}
		return out
	}
	moved, err := s.MoveTask(ctx, alice.ID, terceira.ID, week.Monday, 0)
	if err != nil || moved.Position != 0 {
		t.Fatalf("mover para o início: %+v, %v", moved, err)
	}
	if got := titles(week.Monday); strings.Join(got, ",") != "Terceira,Treino na academia,Leitura" {
		t.Errorf("segunda depois de mover = %v", got)
	}
	if _, err := s.MoveTask(ctx, alice.ID, terceira.ID, week.Monday, 99); err != nil {
		t.Fatal(err)
	}
	if got := titles(week.Monday); strings.Join(got, ",") != "Treino na academia,Leitura,Terceira" {
		t.Errorf("mover além do fim = %v", got)
	}
	if _, err := s.MoveTask(ctx, alice.ID, leitura.ID, week.Saturday, 0); err != nil {
		t.Fatal(err)
	}
	if got := titles(week.Saturday); strings.Join(got, ",") != "Leitura,Feira" {
		t.Errorf("sábado depois de receber = %v", got)
	}
	if got := titles(week.Monday); strings.Join(got, ",") != "Treino na academia,Terceira" {
		t.Errorf("segunda depois de ceder = %v", got)
	}
	if _, err := s.MoveTask(ctx, bob.ID, leitura.ID, week.Monday, 0); !errors.Is(err, ErrNotFound) {
		t.Errorf("mover tarefa alheia: %v", err)
	}

	if _, err := s.MoveTask(ctx, alice.ID, leitura.ID, week.Weekday(7), 0); !errors.Is(err, ErrNotFound) {
		t.Errorf("mover para dia 7: %v", err)
	}
	if _, err := s.MoveTask(ctx, alice.ID, leitura.ID, week.Monday, 1); err != nil {
		t.Fatal(err)
	}
	if got := titles(week.Monday); strings.Join(got, ",") != "Treino na academia,Leitura,Terceira" {
		t.Errorf("segunda depois de voltar = %v", got)
	}

	// Apagar fecha o buraco na ordem.
	if err := s.DeleteTask(ctx, bob.ID, treino.ID); !errors.Is(err, ErrNotFound) {
		t.Errorf("apagar tarefa alheia: %v", err)
	}
	if err := s.DeleteTask(ctx, alice.ID, treino.ID); err != nil {
		t.Fatal(err)
	}
	if got := titles(week.Monday); strings.Join(got, ",") != "Leitura,Terceira" {
		t.Errorf("segunda depois de apagar = %v", got)
	}
	if _, err := s.Task(ctx, alice.ID, treino.ID); !errors.Is(err, ErrNotFound) {
		t.Errorf("tarefa apagada ainda existe: %v", err)
	}
	if got, _ := s.Task(ctx, alice.ID, feira.ID); got.Title != "Feira" {
		t.Errorf("Task = %+v", got)
	}

	// Duplicar: a cópia entra logo abaixo, com título e horário, sem "feita".
	if _, err := s.UpdateTask(ctx, alice.ID, leitura.ID, TaskPatch{Time: ptr("08:15"), Done: ptr(true)}); err != nil {
		t.Fatal(err)
	}
	copyTask, err := s.DuplicateTask(ctx, alice.ID, leitura.ID)
	if err != nil {
		t.Fatal(err)
	}
	if copyTask.ID == leitura.ID || copyTask.Title != "Leitura" || copyTask.Time != "08:15" || copyTask.Done || copyTask.Position != 1 || copyTask.Weekday != week.Monday {
		t.Errorf("cópia = %+v", copyTask)
	}
	if got := titles(week.Monday); strings.Join(got, ",") != "Leitura,Leitura,Terceira" {
		t.Errorf("segunda depois de duplicar = %v", got)
	}
	if stored, _ := s.Task(ctx, alice.ID, copyTask.ID); stored != copyTask {
		t.Errorf("cópia lida = %+v, devolvida = %+v", stored, copyTask)
	}
	if _, err := s.DuplicateTask(ctx, bob.ID, leitura.ID); !errors.Is(err, ErrNotFound) {
		t.Errorf("duplicar tarefa alheia: %v", err)
	}
	if _, err := s.DuplicateTask(ctx, alice.ID, "nao-existe-aqui-00"); !errors.Is(err, ErrNotFound) {
		t.Errorf("duplicar inexistente: %v", err)
	}
}

func TestSessionsAndGoogleLinking(t *testing.T) {
	s := openTest(t)
	ctx := context.Background()

	visitor, err := s.CreateAnonymousUser(ctx)
	if err != nil {
		t.Fatal(err)
	}
	token, created, err := s.CreateSession(ctx, visitor.ID)
	if err != nil {
		t.Fatal(err)
	}
	sess, u, err := s.SessionByToken(ctx, token)
	if err != nil {
		t.Fatal(err)
	}
	if u.ID != visitor.ID || !u.Anonymous() || sess.LastWeekID != "" || sess.ID != created.ID || u.WeekOrder != OrderRecent {
		t.Errorf("sessão = %+v, usuário = %+v", sess, u)
	}
	if sess.ID == token {
		t.Error("o id da sessão é o próprio token: só o hash pode ficar no banco")
	}
	if _, _, err := s.SessionByToken(ctx, "token-inventado"); !errors.Is(err, ErrNotFound) {
		t.Errorf("token inventado: %v", err)
	}

	w, err := s.CreateWeek(ctx, visitor.ID, "Semana padrão")
	if err != nil {
		t.Fatal(err)
	}
	if err := s.SetLastWeek(ctx, sess.ID, w.ID); err != nil {
		t.Fatal(err)
	}
	if sess, _, _ = s.SessionByToken(ctx, token); sess.LastWeekID != w.ID {
		t.Errorf("LastWeekID = %q", sess.LastWeekID)
	}
	// Apagar a semana lembrada não quebra a sessão.
	if err := s.DeleteWeek(ctx, visitor.ID, w.ID); err != nil {
		t.Fatal(err)
	}
	if sess, _, err = s.SessionByToken(ctx, token); err != nil || sess.LastWeekID != "" {
		t.Errorf("depois de apagar a semana: %+v, %v", sess, err)
	}

	// O visitante entra com o Google: o mesmo usuário ganha identidade.
	g := GoogleIdentity{Sub: "google-1", Email: "miguel@example.com", Name: "Miguel", Picture: "https://lh3.googleusercontent.com/x"}
	if err := s.LinkGoogle(ctx, visitor.ID, g); err != nil {
		t.Fatal(err)
	}
	linked, err := s.UserByGoogleSub(ctx, g.Sub)
	if err != nil {
		t.Fatal(err)
	}
	if linked.ID != visitor.ID || linked.Anonymous() || linked.Email != g.Email || linked.Name != "Miguel" {
		t.Errorf("usuário ligado = %+v", linked)
	}
	if _, err := s.UserByGoogleSub(ctx, "ninguem"); !errors.Is(err, ErrNotFound) {
		t.Errorf("sub desconhecido: %v", err)
	}

	// Outro visitante anônimo, com uma semana, entra com a mesma conta Google:
	// as semanas migram e o anônimo some.
	other, err := s.CreateAnonymousUser(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.CreateWeek(ctx, other.ID, "Semana no celular"); err != nil {
		t.Fatal(err)
	}
	otherToken, _, err := s.CreateSession(ctx, other.ID)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.MergeUsers(ctx, other.ID, linked.ID); err != nil {
		t.Fatal(err)
	}
	list, err := s.ListWeeks(ctx, linked.ID, OrderRecent)
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 1 || list[0].Name != "Semana no celular" {
		t.Errorf("depois da fusão a conta tem %+v", list)
	}
	if _, err := s.UserByID(ctx, other.ID); !errors.Is(err, ErrNotFound) {
		t.Errorf("usuário anônimo deveria ter sumido: %v", err)
	}
	if _, moved, err := s.SessionByToken(ctx, otherToken); err != nil || moved.ID != linked.ID {
		t.Errorf("sessão do anônimo deveria apontar para a conta: %+v, %v", moved, err)
	}

	// E-mail já usado por outra conta não pode ser ligado a um terceiro.
	third, err := s.CreateAnonymousUser(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.LinkGoogle(ctx, third.ID, GoogleIdentity{Sub: "google-2", Email: "MIGUEL@example.com"}); !errors.Is(err, ErrEmailTaken) {
		t.Errorf("e-mail repetido: %v, quero ErrEmailTaken", err)
	}

	if err := s.DeleteSession(ctx, token); err != nil {
		t.Fatal(err)
	}
	if _, _, err := s.SessionByToken(ctx, token); !errors.Is(err, ErrNotFound) {
		t.Errorf("sessão encerrada ainda vale: %v", err)
	}
	if n, err := s.DeleteExpiredSessions(ctx); err != nil || n != 0 {
		t.Errorf("DeleteExpiredSessions = %d, %v", n, err)
	}
}

func ptr[T any](v T) *T { return &v }

package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/AdeptLabsDev/weeklly/internal/ids"
	"github.com/AdeptLabsDev/weeklly/internal/week"
)

// MaxTasksPerDay é o limite de tarefas num dia. Um plano com mais que isso
// não é um plano.
const MaxTasksPerDay = 100

// ErrDayFull é devolvido ao passar de MaxTasksPerDay.
var ErrDayFull = fmt.Errorf("um dia comporta até %d tarefas", MaxTasksPerDay)

// ownedTask restringe qualquer escrita em tarefas às semanas do usuário.
const ownedTask = `id = ? AND week_id IN (SELECT id FROM weeks WHERE user_id = ?)`

// AddTask põe uma tarefa no fim do dia. ErrNotFound se a semana não é do usuário.
func (s *Store) AddTask(ctx context.Context, userID, weekID string, day week.Weekday, rawTitle, rawTime string) (t week.Task, err error) {
	title, err := week.CleanTitle(rawTitle)
	if err != nil {
		return week.Task{}, err
	}
	at, err := week.CleanTime(rawTime)
	if err != nil {
		return week.Task{}, err
	}
	if !day.Valid() || !ids.Valid(weekID) {
		return week.Task{}, ErrNotFound
	}

	tx, err := s.writer.BeginTx(ctx, nil)
	if err != nil {
		return week.Task{}, err
	}
	defer func() {
		if err != nil {
			err = errors.Join(err, tx.Rollback())
		}
	}()

	var count, next int
	err = tx.QueryRowContext(ctx, `
		SELECT count(*), coalesce(max(position) + 1, 0)
		  FROM tasks
		 WHERE week_id = (SELECT id FROM weeks WHERE id = ? AND user_id = ?) AND weekday = ?`,
		weekID, userID, int(day)).Scan(&count, &next)
	if err != nil {
		return week.Task{}, err
	}
	if count >= MaxTasksPerDay {
		return week.Task{}, ErrDayFull
	}
	t = week.Task{ID: ids.New(), Weekday: day, Position: next, Title: title, Time: at}
	_, err = tx.ExecContext(ctx, `
		INSERT INTO tasks (id, week_id, weekday, position, title, time)
		SELECT ?, id, ?, ?, ?, ? FROM weeks WHERE id = ? AND user_id = ?`,
		t.ID, int(day), next, title, nullable(at), weekID, userID)
	if err != nil {
		return week.Task{}, fmt.Errorf("criando tarefa: %w", err)
	}
	if err = affectedTx(tx, ctx); err != nil {
		return week.Task{}, err
	}
	return t, tx.Commit()
}

// affectedTx confere que o último INSERT ... SELECT da transação achou a semana.
func affectedTx(tx *sql.Tx, ctx context.Context) error {
	var n int
	if err := tx.QueryRowContext(ctx, `SELECT changes()`).Scan(&n); err != nil {
		return err
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

// TaskPatch é o que pode mudar numa tarefa. Campos nulos ficam como estão.
type TaskPatch struct {
	Title *string
	Time  *string
	Done  *bool
}

// UpdateTask aplica o patch. ErrNotFound se a tarefa não é do usuário.
func (s *Store) UpdateTask(ctx context.Context, userID, taskID string, p TaskPatch) (week.Task, error) {
	if !ids.Valid(taskID) {
		return week.Task{}, ErrNotFound
	}
	if p.Title != nil {
		title, err := week.CleanTitle(*p.Title)
		if err != nil {
			return week.Task{}, err
		}
		if _, err := s.writer.ExecContext(ctx, `UPDATE tasks SET title = ? WHERE `+ownedTask, title, taskID, userID); err != nil {
			return week.Task{}, fmt.Errorf("editando tarefa: %w", err)
		}
	}
	if p.Time != nil {
		at, err := week.CleanTime(*p.Time)
		if err != nil {
			return week.Task{}, err
		}
		if _, err := s.writer.ExecContext(ctx, `UPDATE tasks SET time = ? WHERE `+ownedTask, nullable(at), taskID, userID); err != nil {
			return week.Task{}, fmt.Errorf("editando horário: %w", err)
		}
	}
	if p.Done != nil {
		done := 0
		if *p.Done {
			done = 1
		}
		if _, err := s.writer.ExecContext(ctx, `UPDATE tasks SET done = ? WHERE `+ownedTask, done, taskID, userID); err != nil {
			return week.Task{}, fmt.Errorf("marcando tarefa: %w", err)
		}
	}
	return s.Task(ctx, userID, taskID)
}

// DeleteTask apaga a tarefa e fecha o buraco na ordem do dia.
func (s *Store) DeleteTask(ctx context.Context, userID, taskID string) (err error) {
	if !ids.Valid(taskID) {
		return ErrNotFound
	}
	tx, err := s.writer.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			err = errors.Join(err, tx.Rollback())
		}
	}()
	var weekID string
	var day, position int
	err = tx.QueryRowContext(ctx, `SELECT week_id, weekday, position FROM tasks WHERE `+ownedTask, taskID, userID).Scan(&weekID, &day, &position)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrNotFound
		}
		return err
	}
	if _, err = tx.ExecContext(ctx, `DELETE FROM tasks WHERE id = ?`, taskID); err != nil {
		return fmt.Errorf("apagando tarefa: %w", err)
	}
	if _, err = tx.ExecContext(ctx,
		`UPDATE tasks SET position = position - 1 WHERE week_id = ? AND weekday = ? AND position > ?`,
		weekID, day, position); err != nil {
		return fmt.Errorf("reordenando dia: %w", err)
	}
	return tx.Commit()
}

// MoveTask põe a tarefa na posição dada de um dia (o mesmo ou outro),
// fechando o buraco de onde saiu e abrindo espaço onde entrou. position é o
// índice final na lista do dia; acima do fim, vira o fim.
func (s *Store) MoveTask(ctx context.Context, userID, taskID string, day week.Weekday, position int) (t week.Task, err error) {
	if !ids.Valid(taskID) || !day.Valid() {
		return week.Task{}, ErrNotFound
	}
	if position < 0 {
		position = 0
	}
	tx, err := s.writer.BeginTx(ctx, nil)
	if err != nil {
		return week.Task{}, err
	}
	defer func() {
		if err != nil {
			err = errors.Join(err, tx.Rollback())
		}
	}()

	var weekID string
	var fromDay, fromPos int
	err = tx.QueryRowContext(ctx, `SELECT week_id, weekday, position FROM tasks WHERE `+ownedTask, taskID, userID).Scan(&weekID, &fromDay, &fromPos)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return week.Task{}, ErrNotFound
		}
		return week.Task{}, err
	}

	var others int
	if err = tx.QueryRowContext(ctx,
		`SELECT count(*) FROM tasks WHERE week_id = ? AND weekday = ? AND id != ?`, weekID, int(day), taskID).Scan(&others); err != nil {
		return week.Task{}, err
	}
	if int(day) != fromDay && others >= MaxTasksPerDay {
		return week.Task{}, ErrDayFull
	}
	if position > others {
		position = others
	}
	if int(day) == fromDay && position == fromPos {
		return s.Task(ctx, userID, taskID)
	}

	if _, err = tx.ExecContext(ctx,
		`UPDATE tasks SET position = position - 1 WHERE week_id = ? AND weekday = ? AND position > ?`,
		weekID, fromDay, fromPos); err != nil {
		return week.Task{}, fmt.Errorf("fechando espaço: %w", err)
	}
	if _, err = tx.ExecContext(ctx,
		`UPDATE tasks SET position = position + 1 WHERE week_id = ? AND weekday = ? AND position >= ? AND id != ?`,
		weekID, int(day), position, taskID); err != nil {
		return week.Task{}, fmt.Errorf("abrindo espaço: %w", err)
	}
	if _, err = tx.ExecContext(ctx,
		`UPDATE tasks SET weekday = ?, position = ? WHERE id = ?`, int(day), position, taskID); err != nil {
		return week.Task{}, fmt.Errorf("movendo tarefa: %w", err)
	}
	if err = tx.Commit(); err != nil {
		return week.Task{}, err
	}
	return s.Task(ctx, userID, taskID)
}

// Task lê uma tarefa do usuário.
func (s *Store) Task(ctx context.Context, userID, taskID string) (week.Task, error) {
	row := s.reader.QueryRowContext(ctx,
		`SELECT id, weekday, position, title, time, done FROM tasks WHERE `+ownedTask, taskID, userID)
	t, err := scanTask(row)
	if errors.Is(err, sql.ErrNoRows) {
		return week.Task{}, ErrNotFound
	}
	return t, err
}

type scanner interface {
	Scan(dest ...any) error
}

func scanTask(row scanner) (week.Task, error) {
	var t week.Task
	var day, done int
	var at sql.NullString
	if err := row.Scan(&t.ID, &day, &t.Position, &t.Title, &at, &done); err != nil {
		return week.Task{}, err
	}
	t.Weekday = week.Weekday(day)
	t.Time = at.String
	t.Done = done == 1
	return t, nil
}

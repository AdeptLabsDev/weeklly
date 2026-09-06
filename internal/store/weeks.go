package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/AdeptLabsDev/weeklly/internal/ids"
	"github.com/AdeptLabsDev/weeklly/internal/week"
)

// WeekOrder é como a lista de semanas é ordenada.
type WeekOrder string

// Ordens aceitas. Recent é o padrão.
const (
	OrderRecent WeekOrder = "recent"
	OrderName   WeekOrder = "name"
)

// ParseWeekOrder lê a ordem vinda de um formulário; qualquer outra coisa é Recent.
func ParseWeekOrder(s string) WeekOrder {
	if WeekOrder(s) == OrderName {
		return OrderName
	}
	return OrderRecent
}

// WeekSummary é o que o hub e o seletor de semanas mostram.
type WeekSummary struct {
	ID        string
	Name      string
	UpdatedAt time.Time
}

// CreateWeek cria uma semana vazia com o nome dado.
func (s *Store) CreateWeek(ctx context.Context, userID, rawName string) (week.Week, error) {
	name, err := week.CleanName(rawName)
	if err != nil {
		return week.Week{}, err
	}
	id := ids.New()
	if _, err := s.writer.ExecContext(ctx,
		`INSERT INTO weeks (id, user_id, name) VALUES (?, ?, ?)`, id, userID, name); err != nil {
		return week.Week{}, fmt.Errorf("criando semana: %w", err)
	}
	return week.Week{ID: id, Name: name}, nil
}

// RenameWeek troca o nome. ErrNotFound se a semana não é do usuário.
func (s *Store) RenameWeek(ctx context.Context, userID, weekID, rawName string) error {
	name, err := week.CleanName(rawName)
	if err != nil {
		return err
	}
	res, err := s.writer.ExecContext(ctx,
		`UPDATE weeks SET name = ?, updated_at = ? WHERE id = ? AND user_id = ?`, name, now(), weekID, userID)
	if err != nil {
		return fmt.Errorf("renomeando semana: %w", err)
	}
	return affected(res)
}

// DeleteWeek apaga a semana e suas tarefas. ErrNotFound se não é do usuário.
func (s *Store) DeleteWeek(ctx context.Context, userID, weekID string) error {
	res, err := s.writer.ExecContext(ctx, `DELETE FROM weeks WHERE id = ? AND user_id = ?`, weekID, userID)
	if err != nil {
		return fmt.Errorf("apagando semana: %w", err)
	}
	return affected(res)
}

// DuplicateWeek copia a semana com todas as tarefas (não feitas) para uma
// semana nova com o nome dado. É o gesto central do produto: reaproveitar
// um plano.
func (s *Store) DuplicateWeek(ctx context.Context, userID, weekID, rawName string) (w week.Week, err error) {
	name, err := week.CleanName(rawName)
	if err != nil {
		return week.Week{}, err
	}
	tx, err := s.writer.BeginTx(ctx, nil)
	if err != nil {
		return week.Week{}, err
	}
	defer func() {
		if err != nil {
			err = errors.Join(err, tx.Rollback())
		}
	}()

	var exists int
	if err = tx.QueryRowContext(ctx, `SELECT 1 FROM weeks WHERE id = ? AND user_id = ?`, weekID, userID).Scan(&exists); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return week.Week{}, ErrNotFound
		}
		return week.Week{}, err
	}
	id := ids.New()
	if _, err = tx.ExecContext(ctx, `INSERT INTO weeks (id, user_id, name) VALUES (?, ?, ?)`, id, userID, name); err != nil {
		return week.Week{}, fmt.Errorf("criando cópia: %w", err)
	}
	copies, err := tasksToCopy(ctx, tx, weekID)
	if err != nil {
		return week.Week{}, err
	}
	for _, c := range copies {
		if _, err = tx.ExecContext(ctx,
			`INSERT INTO tasks (id, week_id, weekday, position, title, time) VALUES (?, ?, ?, ?, ?, ?)`,
			ids.New(), id, c.weekday, c.position, c.title, c.t); err != nil {
			return week.Week{}, fmt.Errorf("copiando tarefas: %w", err)
		}
	}
	if err = tx.Commit(); err != nil {
		return week.Week{}, err
	}
	return s.Week(ctx, userID, id)
}

type copyTask struct {
	weekday, position int
	title             string
	t                 sql.NullString
}

// tasksToCopy lê as tarefas de uma semana inteiras antes de inserir: na
// mesma transação, ler e escrever ao mesmo tempo não é seguro.
func tasksToCopy(ctx context.Context, tx *sql.Tx, weekID string) ([]copyTask, error) {
	rows, err := tx.QueryContext(ctx,
		`SELECT weekday, position, title, time FROM tasks WHERE week_id = ? ORDER BY weekday, position`, weekID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var copies []copyTask
	for rows.Next() {
		var c copyTask
		if err := rows.Scan(&c.weekday, &c.position, &c.title, &c.t); err != nil {
			return nil, err
		}
		copies = append(copies, c)
	}
	return copies, rows.Err()
}

// ListWeeks devolve as semanas do usuário na ordem pedida.
func (s *Store) ListWeeks(ctx context.Context, userID string, order WeekOrder) ([]WeekSummary, error) {
	// rowid desempata semanas criadas no mesmo milissegundo.
	orderBy := `updated_at DESC, rowid DESC`
	if order == OrderName {
		orderBy = `name COLLATE NOCASE ASC, rowid ASC`
	}
	rows, err := s.reader.QueryContext(ctx,
		`SELECT id, name, updated_at FROM weeks WHERE user_id = ? ORDER BY `+orderBy, userID)
	if err != nil {
		return nil, fmt.Errorf("listando semanas: %w", err)
	}
	defer rows.Close()

	var out []WeekSummary
	for rows.Next() {
		var w WeekSummary
		var updated string
		if err := rows.Scan(&w.ID, &w.Name, &updated); err != nil {
			return nil, err
		}
		if w.UpdatedAt, err = parseTime(updated); err != nil {
			return nil, err
		}
		out = append(out, w)
	}
	return out, rows.Err()
}

// Week carrega uma semana e as tarefas de cada dia, na ordem manual.
// Devolve ErrNotFound se não existe ou não pertence ao usuário.
func (s *Store) Week(ctx context.Context, userID, weekID string) (week.Week, error) {
	if !ids.Valid(weekID) {
		return week.Week{}, ErrNotFound
	}
	var w week.Week
	err := s.reader.QueryRowContext(ctx,
		`SELECT id, name FROM weeks WHERE id = ? AND user_id = ?`, weekID, userID).Scan(&w.ID, &w.Name)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return week.Week{}, ErrNotFound
		}
		return week.Week{}, fmt.Errorf("lendo semana: %w", err)
	}

	rows, err := s.reader.QueryContext(ctx,
		`SELECT id, weekday, position, title, time, done FROM tasks WHERE week_id = ? ORDER BY weekday, position, rowid`, weekID)
	if err != nil {
		return week.Week{}, fmt.Errorf("lendo tarefas: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		t, err := scanTask(rows)
		if err != nil {
			return week.Week{}, err
		}
		if t.Weekday.Valid() {
			w.Days[t.Weekday] = append(w.Days[t.Weekday], t)
		}
	}
	return w, rows.Err()
}

// SetWeekOrder guarda a preferência de ordem da lista de semanas.
func (s *Store) SetWeekOrder(ctx context.Context, userID string, order WeekOrder) error {
	res, err := s.writer.ExecContext(ctx, `UPDATE users SET week_order = ? WHERE id = ?`, string(order), userID)
	if err != nil {
		return fmt.Errorf("guardando ordem: %w", err)
	}
	return affected(res)
}

func affected(res sql.Result) error {
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

// parseTime lê os instantes gravados pelo SQLite (strftime com %f).
func parseTime(s string) (time.Time, error) {
	t, err := time.Parse(timeLayout, s)
	if err != nil {
		return time.Time{}, fmt.Errorf("instante inválido no banco %q: %w", s, err)
	}
	return t, nil
}

package store

import (
	"context"
	"path/filepath"
	"strings"
	"testing"
)

const (
	userA = "aaaaaaaaaa234567"
	userB = "bbbbbbbbbb234567"
	weekA = "abcdefghij234567"
	weekB = "zyxwvutsrq765432"
	taskA = "tttttttttt234567"
)

func openTest(t *testing.T) *Store {
	t.Helper()
	ctx := context.Background()
	s, err := Open(ctx, filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() {
		if err := s.Close(); err != nil {
			t.Errorf("Close: %v", err)
		}
	})
	if err := s.Migrate(ctx); err != nil {
		t.Fatalf("Migrate: %v", err)
	}
	return s
}

func TestOpenAppliesMandatoryPragmas(t *testing.T) {
	s := openTest(t)
	ctx := context.Background()

	var journal string
	if err := s.reader.QueryRowContext(ctx, `PRAGMA journal_mode`).Scan(&journal); err != nil {
		t.Fatal(err)
	}
	if journal != "wal" {
		t.Errorf("journal_mode = %q, quero wal", journal)
	}

	var fk int
	if err := s.writer.QueryRowContext(ctx, `PRAGMA foreign_keys`).Scan(&fk); err != nil {
		t.Fatal(err)
	}
	if fk != 1 {
		t.Errorf("foreign_keys = %d, quero 1", fk)
	}

	if _, err := s.reader.ExecContext(ctx, `INSERT INTO users (id) VALUES (?)`, userA); err == nil {
		t.Error("pool de leitura aceitou uma escrita")
	}
}

func TestMigrateIsIdempotent(t *testing.T) {
	s := openTest(t)
	ctx := context.Background()
	if err := s.Migrate(ctx); err != nil {
		t.Fatalf("segunda Migrate: %v", err)
	}
	v, err := s.SchemaVersion(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if v != 1 {
		t.Errorf("SchemaVersion = %d, quero 1", v)
	}
	if err := s.Health(ctx); err != nil {
		t.Errorf("Health: %v", err)
	}
}

func TestSchemaEnforcesProductInvariants(t *testing.T) {
	s := openTest(t)
	ctx := context.Background()
	exec := func(q string, args ...any) error {
		_, err := s.writer.ExecContext(ctx, q, args...)
		return err
	}
	mustFail := func(what, q string, args ...any) {
		t.Helper()
		if err := exec(q, args...); err == nil {
			t.Errorf("%s: o banco aceitou", what)
		}
	}
	count := func(q string, args ...any) int {
		t.Helper()
		var n int
		if err := s.reader.QueryRowContext(ctx, q, args...).Scan(&n); err != nil {
			t.Fatal(err)
		}
		return n
	}

	if err := exec(`INSERT INTO users (id, email) VALUES (?, 'miguel@example.com')`, userA); err != nil {
		t.Fatal(err)
	}
	mustFail("e-mail duplicado ignorando caixa", `INSERT INTO users (id, email) VALUES (?, 'MIGUEL@example.com')`, userB)
	mustFail("id de usuário fora do formato", `INSERT INTO users (id) VALUES ('u2')`)
	mustFail("ordem de semanas desconhecida", `UPDATE users SET week_order = 'color' WHERE id = ?`, userA)
	if err := exec(`INSERT INTO users (id) VALUES (?)`, userB); err != nil {
		t.Fatalf("usuário anônimo rejeitado: %v", err)
	}

	if err := exec(`INSERT INTO weeks (id, user_id, name) VALUES (?, ?, 'Semana padrão')`, weekA, userA); err != nil {
		t.Fatalf("semana válida rejeitada: %v", err)
	}
	mustFail("id curto", `INSERT INTO weeks (id, user_id, name) VALUES ('curto', ?, 'x')`, userA)
	mustFail("id com caractere fora de [a-z2-7]", `INSERT INTO weeks (id, user_id, name) VALUES ('ABCDEFGHIJ234567', ?, 'x')`, userA)
	mustFail("nome vazio", `INSERT INTO weeks (id, user_id, name) VALUES (?, ?, '')`, weekB, userA)
	mustFail("nome com espaço na ponta", `INSERT INTO weeks (id, user_id, name) VALUES (?, ?, ' x')`, weekB, userA)
	mustFail("nome longo demais", `INSERT INTO weeks (id, user_id, name) VALUES (?, ?, ?)`, weekB, userA, strings.Repeat("a", 61))
	mustFail("semana de usuário inexistente", `INSERT INTO weeks (id, user_id, name) VALUES (?, 'nope', 'x')`, weekB)

	if err := exec(`INSERT INTO tasks (id, week_id, weekday, position, title, time) VALUES (?, ?, 0, 0, 'Treino', '07:00')`, taskA, weekA); err != nil {
		t.Fatalf("tarefa válida rejeitada: %v", err)
	}
	mustFail("dia 7", `INSERT INTO tasks (id, week_id, weekday, position, title) VALUES ('t2t2t2t2t2234567', ?, 7, 0, 'x')`, weekA)
	mustFail("posição negativa", `INSERT INTO tasks (id, week_id, weekday, position, title) VALUES ('t2t2t2t2t2234567', ?, 0, -1, 'x')`, weekA)
	mustFail("título vazio", `INSERT INTO tasks (id, week_id, weekday, position, title) VALUES ('t2t2t2t2t2234567', ?, 0, 1, '')`, weekA)
	mustFail("título longo demais", `INSERT INTO tasks (id, week_id, weekday, position, title) VALUES ('t2t2t2t2t2234567', ?, 0, 1, ?)`, weekA, strings.Repeat("a", 201))
	mustFail("horário 24:00", `INSERT INTO tasks (id, week_id, weekday, position, title, time) VALUES ('t2t2t2t2t2234567', ?, 0, 1, 'x', '24:00')`, weekA)
	mustFail("horário sem zero à esquerda", `INSERT INTO tasks (id, week_id, weekday, position, title, time) VALUES ('t2t2t2t2t2234567', ?, 0, 1, 'x', '9:30')`, weekA)
	mustFail("done fora de 0/1", `UPDATE tasks SET done = 2 WHERE id = ?`, taskA)
	mustFail("tarefa de semana inexistente", `INSERT INTO tasks (id, week_id, weekday, position, title) VALUES ('t2t2t2t2t2234567', ?, 0, 0, 'x')`, weekB)
	mustFail("título binário em coluna TEXT (STRICT)", `UPDATE tasks SET title = X'00ff' WHERE id = ?`, taskA)

	// Mexer numa tarefa marca a semana como recente e carimba a tarefa.
	if err := exec(`UPDATE weeks SET updated_at = '2000-01-01T00:00:00.000Z' WHERE id = ?`, weekA); err != nil {
		t.Fatal(err)
	}
	if err := exec(`UPDATE tasks SET updated_at = '2000-01-01T00:00:00.000Z' WHERE id = ?`, taskA); err != nil {
		t.Fatal(err)
	}
	if err := exec(`UPDATE tasks SET done = 1 WHERE id = ?`, taskA); err != nil {
		t.Fatal(err)
	}
	var weekTouched, taskTouched string
	if err := s.reader.QueryRowContext(ctx, `SELECT updated_at FROM weeks WHERE id = ?`, weekA).Scan(&weekTouched); err != nil {
		t.Fatal(err)
	}
	if err := s.reader.QueryRowContext(ctx, `SELECT updated_at FROM tasks WHERE id = ?`, taskA).Scan(&taskTouched); err != nil {
		t.Fatal(err)
	}
	if strings.HasPrefix(weekTouched, "2000-") || strings.HasPrefix(taskTouched, "2000-") {
		t.Error("editar uma tarefa não atualizou os carimbos")
	}

	// Apagar o usuário leva semanas, tarefas e sessões junto.
	if err := exec(`INSERT INTO sessions (token_hash, user_id, last_week_id, expires_at) VALUES ('h', ?, ?, '2999-01-01T00:00:00.000Z')`, userA, weekA); err != nil {
		t.Fatal(err)
	}
	if err := exec(`DELETE FROM users WHERE id = ?`, userA); err != nil {
		t.Fatalf("cascade bloqueado: %v", err)
	}
	if n := count(`SELECT count(*) FROM weeks`) + count(`SELECT count(*) FROM tasks`) + count(`SELECT count(*) FROM sessions`); n != 0 {
		t.Errorf("apagar o usuário deixou %d linhas para trás", n)
	}
}

func TestParseMigrationName(t *testing.T) {
	v, name, err := parseMigrationName("0001_schema.sql")
	if err != nil || v != 1 || name != "schema" {
		t.Errorf("parseMigrationName = %d %q %v", v, name, err)
	}
	for _, bad := range []string{"schema.sql", "abc_schema.sql", "0000_x.sql"} {
		if _, _, err := parseMigrationName(bad); err == nil {
			t.Errorf("%q aceito", bad)
		}
	}
}

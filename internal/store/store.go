// Package store abre o SQLite com a configuração obrigatória do projeto
// (D5) e aplica as migrações embutidas no boot.
//
// Dois pools sobre o mesmo arquivo: um escritor com uma única conexão, porque
// o SQLite só admite um escritor por vez e serializar aqui evita SQLITE_BUSY;
// e leitores concorrentes, que o modo WAL permite sem bloquear o escritor.
package store

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	// Driver SQLite em Go puro: sem cgo, sem toolchain C, cross-compile trivial.
	_ "modernc.org/sqlite"
)

//go:embed migrations/*.sql
var migrationFS embed.FS

// Store é o acesso ao banco.
type Store struct {
	writer *sql.DB
	reader *sql.DB
}

// Open abre (ou cria) o arquivo em dbPath com os pragmas obrigatórios.
func Open(ctx context.Context, dbPath string) (*Store, error) {
	if err := os.MkdirAll(filepath.Dir(dbPath), 0o750); err != nil {
		return nil, fmt.Errorf("criando diretório do banco: %w", err)
	}

	writer, err := open(ctx, dsn(dbPath), 1)
	if err != nil {
		return nil, fmt.Errorf("pool de escrita: %w", err)
	}
	reader, err := open(ctx, dsn(dbPath, "_pragma=query_only(ON)"), 4)
	if err != nil {
		return nil, errors.Join(fmt.Errorf("pool de leitura: %w", err), writer.Close())
	}
	return &Store{writer: writer, reader: reader}, nil
}

// dsn monta a URI do SQLite. Os pragmas viajam na URI porque o driver os
// aplica em cada conexão nova do pool, não só na primeira.
func dsn(dbPath string, extra ...string) string {
	p := filepath.ToSlash(dbPath)
	if filepath.IsAbs(dbPath) && !strings.HasPrefix(p, "/") {
		p = "/" + p // Windows: C:/x vira /C:/x, forma que o SQLite aceita em URIs
	}
	escaped := (&url.URL{Path: p}).EscapedPath()

	params := append([]string{
		"_pragma=journal_mode(WAL)",
		"_pragma=synchronous(NORMAL)",
		"_pragma=foreign_keys(ON)",
		"_pragma=busy_timeout(5000)",
	}, extra...)
	return "file:" + escaped + "?" + strings.Join(params, "&")
}

func open(ctx context.Context, dsn string, maxConns int) (*sql.DB, error) {
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(maxConns)
	db.SetMaxIdleConns(maxConns)
	db.SetConnMaxLifetime(0)
	if err := db.PingContext(ctx); err != nil {
		return nil, errors.Join(err, db.Close())
	}
	return db, nil
}

// Close fecha os dois pools.
func (s *Store) Close() error {
	return errors.Join(s.reader.Close(), s.writer.Close())
}

// Health confirma que o banco responde a uma consulta.
func (s *Store) Health(ctx context.Context) error {
	var one int
	if err := s.reader.QueryRowContext(ctx, `SELECT 1`).Scan(&one); err != nil {
		return fmt.Errorf("banco não responde: %w", err)
	}
	return nil
}

// Migrate aplica, em ordem e dentro de transações, as migrações ainda não
// registradas em schema_migrations. É idempotente e roda em todo boot.
func (s *Store) Migrate(ctx context.Context) error {
	const table = `CREATE TABLE IF NOT EXISTS schema_migrations (
		version    INTEGER PRIMARY KEY,
		name       TEXT    NOT NULL,
		applied_at TEXT    NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
	) STRICT`
	if _, err := s.writer.ExecContext(ctx, table); err != nil {
		return fmt.Errorf("criando schema_migrations: %w", err)
	}

	files, err := fs.Glob(migrationFS, "migrations/*.sql")
	if err != nil {
		return err
	}
	sort.Strings(files)

	for _, file := range files {
		version, name, err := parseMigrationName(path.Base(file))
		if err != nil {
			return err
		}
		var applied bool
		if err := s.writer.QueryRowContext(ctx,
			`SELECT EXISTS (SELECT 1 FROM schema_migrations WHERE version = ?)`, version,
		).Scan(&applied); err != nil {
			return fmt.Errorf("consultando migração %d: %w", version, err)
		}
		if applied {
			continue
		}
		body, err := fs.ReadFile(migrationFS, file)
		if err != nil {
			return err
		}
		if err := s.apply(ctx, version, name, string(body)); err != nil {
			return fmt.Errorf("aplicando migração %s: %w", file, err)
		}
	}
	return nil
}

// SchemaVersion devolve a última migração aplicada (0 se nenhuma).
func (s *Store) SchemaVersion(ctx context.Context) (int, error) {
	var v int
	err := s.reader.QueryRowContext(ctx, `SELECT COALESCE(MAX(version), 0) FROM schema_migrations`).Scan(&v)
	return v, err
}

func (s *Store) apply(ctx context.Context, version int, name, body string) (err error) {
	tx, err := s.writer.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			err = errors.Join(err, tx.Rollback())
		}
	}()
	if _, err = tx.ExecContext(ctx, body); err != nil {
		return err
	}
	if _, err = tx.ExecContext(ctx, `INSERT INTO schema_migrations (version, name) VALUES (?, ?)`, version, name); err != nil {
		return err
	}
	return tx.Commit()
}

// parseMigrationName lê "0001_schema.sql" como (1, "schema").
func parseMigrationName(file string) (int, string, error) {
	base := strings.TrimSuffix(file, ".sql")
	prefix, name, ok := strings.Cut(base, "_")
	if !ok {
		return 0, "", fmt.Errorf("migração %q: nome precisa ser NNNN_descricao.sql", file)
	}
	version, err := strconv.Atoi(prefix)
	if err != nil || version <= 0 {
		return 0, "", fmt.Errorf("migração %q: prefixo numérico inválido", file)
	}
	return version, name, nil
}

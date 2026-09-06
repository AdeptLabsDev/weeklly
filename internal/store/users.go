package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/AdeptLabsDev/weeklly/internal/ids"
)

// ErrNotFound é devolvido quando a linha não existe ou não pertence a quem
// pediu. As duas situações são indistinguíveis de propósito.
var ErrNotFound = errors.New("não encontrado")

// ErrEmailTaken é devolvido ao ligar uma conta Google cujo e-mail já pertence
// a outro usuário.
var ErrEmailTaken = errors.New("este e-mail já está em outra conta")

// User é um usuário. Nasce anônimo (só o id) e ganha identidade ao entrar
// com o Google.
type User struct {
	ID         string
	Email      string
	Name       string
	PictureURL string
	GoogleSub  string
	// WeekOrder é como a pessoa prefere ver a lista de semanas.
	WeekOrder WeekOrder
}

// Anonymous informa se o usuário ainda não entrou com uma conta.
func (u User) Anonymous() bool { return u.GoogleSub == "" }

// GoogleIdentity é o que o Google devolve sobre uma pessoa.
type GoogleIdentity struct {
	Sub     string
	Email   string
	Name    string
	Picture string
}

// CreateAnonymousUser cria um usuário sem identidade. Acontece na primeira
// semana criada por um visitante.
func (s *Store) CreateAnonymousUser(ctx context.Context) (User, error) {
	id := ids.New()
	if _, err := s.writer.ExecContext(ctx, `INSERT INTO users (id) VALUES (?)`, id); err != nil {
		return User{}, fmt.Errorf("criando usuário: %w", err)
	}
	return User{ID: id, WeekOrder: OrderRecent}, nil
}

const selectUser = `SELECT id, email, google_sub, name, picture_url, week_order FROM users`

// UserByID busca um usuário pelo id.
func (s *Store) UserByID(ctx context.Context, id string) (User, error) {
	return scanUser(s.reader.QueryRowContext(ctx, selectUser+` WHERE id = ?`, id))
}

// UserByGoogleSub busca o usuário ligado a uma conta Google.
func (s *Store) UserByGoogleSub(ctx context.Context, sub string) (User, error) {
	return scanUser(s.reader.QueryRowContext(ctx, selectUser+` WHERE google_sub = ?`, sub))
}

func scanUser(row *sql.Row) (User, error) {
	var u User
	var email, sub, name, picture sql.NullString
	var order string
	if err := row.Scan(&u.ID, &email, &sub, &name, &picture, &order); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return User{}, ErrNotFound
		}
		return User{}, err
	}
	u.Email, u.GoogleSub, u.Name, u.PictureURL = email.String, sub.String, name.String, picture.String
	u.WeekOrder = ParseWeekOrder(order)
	return u, nil
}

// LinkGoogle dá identidade a um usuário anônimo ou atualiza nome e foto de
// quem já entrou antes.
func (s *Store) LinkGoogle(ctx context.Context, userID string, g GoogleIdentity) error {
	res, err := s.writer.ExecContext(ctx,
		`UPDATE users SET google_sub = ?, email = ?, name = ?, picture_url = ? WHERE id = ?`,
		g.Sub, g.Email, nullable(g.Name), nullable(g.Picture), userID)
	if err != nil {
		if isUniqueViolation(err) {
			return ErrEmailTaken
		}
		return fmt.Errorf("ligando conta Google: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

// MergeUsers move as semanas e sessões de from para into e apaga from. É o
// que acontece quando um visitante anônimo entra com uma conta Google que já
// existe: nada se perde.
func (s *Store) MergeUsers(ctx context.Context, from, into string) (err error) {
	if from == into {
		return nil
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
	if _, err = tx.ExecContext(ctx, `UPDATE weeks SET user_id = ? WHERE user_id = ?`, into, from); err != nil {
		return err
	}
	if _, err = tx.ExecContext(ctx, `UPDATE sessions SET user_id = ? WHERE user_id = ?`, into, from); err != nil {
		return err
	}
	if _, err = tx.ExecContext(ctx, `DELETE FROM users WHERE id = ?`, from); err != nil {
		return err
	}
	return tx.Commit()
}

func nullable(s string) any {
	if s == "" {
		return nil
	}
	return s
}

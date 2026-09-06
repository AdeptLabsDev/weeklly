package store

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"
)

// SessionTTL é a validade de uma sessão. Renova a cada uso (deslizante).
const SessionTTL = 90 * 24 * time.Hour

// touchInterval limita a frequência da renovação: uma escrita por sessão a
// cada intervalo, não uma por requisição.
const touchInterval = time.Hour

// Session é a sessão de um navegador. ID é o hash do token, nunca o token.
type Session struct {
	ID         string
	UserID     string
	LastWeekID string
}

// CreateSession abre uma sessão para o usuário e devolve o token que vai no
// cookie e a sessão criada. Só o hash do token fica no banco.
func (s *Store) CreateSession(ctx context.Context, userID string) (string, Session, error) {
	var raw [32]byte
	if _, err := rand.Read(raw[:]); err != nil {
		return "", Session{}, err
	}
	token := base64.RawURLEncoding.EncodeToString(raw[:])
	hash := hashToken(token)
	expires := time.Now().UTC().Add(SessionTTL).Format(timeLayout)
	if _, err := s.writer.ExecContext(ctx,
		`INSERT INTO sessions (token_hash, user_id, expires_at) VALUES (?, ?, ?)`,
		hash, userID, expires); err != nil {
		return "", Session{}, fmt.Errorf("criando sessão: %w", err)
	}
	return token, Session{ID: hash, UserID: userID}, nil
}

// SessionByToken devolve a sessão viva e seu usuário, renovando a validade
// no máximo uma vez por hora. Token desconhecido ou vencido é ErrNotFound.
func (s *Store) SessionByToken(ctx context.Context, token string) (Session, User, error) {
	if token == "" {
		return Session{}, User{}, ErrNotFound
	}
	hash := hashToken(token)
	var sess Session
	var lastWeek, lastSeen sql.NullString
	var u User
	var email, sub, name, picture sql.NullString
	var order string
	err := s.reader.QueryRowContext(ctx, `
		SELECT s.token_hash, s.user_id, s.last_week_id, s.last_seen_at,
		       u.email, u.google_sub, u.name, u.picture_url, u.week_order
		  FROM sessions s JOIN users u ON u.id = s.user_id
		 WHERE s.token_hash = ? AND s.expires_at > ?`,
		hash, now()).Scan(&sess.ID, &sess.UserID, &lastWeek, &lastSeen, &email, &sub, &name, &picture, &order)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Session{}, User{}, ErrNotFound
		}
		return Session{}, User{}, fmt.Errorf("lendo sessão: %w", err)
	}
	sess.LastWeekID = lastWeek.String
	u = User{ID: sess.UserID, Email: email.String, GoogleSub: sub.String, Name: name.String, PictureURL: picture.String, WeekOrder: ParseWeekOrder(order)}

	if seen, err := parseTime(lastSeen.String); err == nil && time.Since(seen) > touchInterval {
		_, err := s.writer.ExecContext(ctx,
			`UPDATE sessions SET last_seen_at = ?, expires_at = ? WHERE token_hash = ?`,
			now(), time.Now().UTC().Add(SessionTTL).Format(timeLayout), hash)
		if err != nil {
			return Session{}, User{}, fmt.Errorf("renovando sessão: %w", err)
		}
	}
	return sess, u, nil
}

// DeleteSession encerra a sessão do token. Token desconhecido não é erro.
func (s *Store) DeleteSession(ctx context.Context, token string) error {
	_, err := s.writer.ExecContext(ctx, `DELETE FROM sessions WHERE token_hash = ?`, hashToken(token))
	return err
}

// SetLastWeek lembra a última semana aberta na sessão.
func (s *Store) SetLastWeek(ctx context.Context, sessionID, weekID string) error {
	_, err := s.writer.ExecContext(ctx, `UPDATE sessions SET last_week_id = ? WHERE token_hash = ?`, weekID, sessionID)
	return err
}

// DeleteExpiredSessions apaga sessões vencidas. Para rodar de tempos em tempos.
func (s *Store) DeleteExpiredSessions(ctx context.Context) (int64, error) {
	res, err := s.writer.ExecContext(ctx, `DELETE FROM sessions WHERE expires_at <= ?`, now())
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

const timeLayout = "2006-01-02T15:04:05.000Z"

func now() string { return time.Now().UTC().Format(timeLayout) }

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

// isUniqueViolation reconhece a violação de UNIQUE do SQLite sem depender
// dos tipos internos do driver.
func isUniqueViolation(err error) bool {
	return err != nil && strings.Contains(err.Error(), "UNIQUE constraint failed")
}

package server

import (
	"context"
	"crypto/rand"
	"errors"
	"log/slog"
	"net/http"
	"runtime/debug"
	"strings"
	"time"

	"github.com/AdeptLabsDev/weeklly/internal/config"
)

type middleware func(http.Handler) http.Handler

type ctxKey int

const requestIDKey ctxKey = iota

// requestID devolve o identificador da requisição, para correlacionar logs.
func requestID(ctx context.Context) string {
	id, _ := ctx.Value(requestIDKey).(string)
	return id
}

// securityHeaders aplica a política de segurança do navegador em toda resposta.
//
// A CSP parte de 'none' e libera só o que o produto usa: scripts, estilos,
// fontes e imagens do próprio domínio, mais a foto de perfil servida pelo
// Google. Sem estilos inline (nem em atributo style) e sem outros terceiros.
// HSTS só em produção, onde há TLS na frente.
func securityHeaders(cfg config.Config) middleware {
	csp := strings.Join([]string{
		"default-src 'none'",
		"script-src 'self'",
		"style-src 'self'",
		"img-src 'self' data: https://lh3.googleusercontent.com",
		"font-src 'self'",
		"connect-src 'self'",
		"manifest-src 'self'",
		"base-uri 'none'",
		"form-action 'self'",
		"frame-ancestors 'none'",
	}, "; ")

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			h := w.Header()
			h.Set("Content-Security-Policy", csp)
			h.Set("X-Content-Type-Options", "nosniff")
			h.Set("X-Frame-Options", "DENY")
			h.Set("Referrer-Policy", "strict-origin-when-cross-origin")
			h.Set("Permissions-Policy", "camera=(), microphone=(), geolocation=(), payment=(), usb=()")
			h.Set("Cross-Origin-Opener-Policy", "same-origin")
			h.Set("Cross-Origin-Resource-Policy", "same-origin")
			if !cfg.IsDev() {
				h.Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
			}
			next.ServeHTTP(w, r)
		})
	}
}

// logging registra uma linha por requisição com status, bytes e duração.
// /healthz vai em nível debug para não inundar o log com o polling do container.
func logging(logger *slog.Logger) middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			id := rand.Text()
			ctx := context.WithValue(r.Context(), requestIDKey, id)
			sw := &statusWriter{ResponseWriter: w}

			next.ServeHTTP(sw, r.WithContext(ctx))

			level := slog.LevelInfo
			if r.URL.Path == "/healthz" {
				level = slog.LevelDebug
			}
			logger.LogAttrs(ctx, level, "http",
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.Int("status", sw.status),
				slog.Int("bytes", sw.bytes),
				slog.Duration("duration", time.Since(start)),
				slog.String("request_id", id),
			)
		})
	}
}

// recoverer transforma panic em 500 com stack no log, em vez de derrubar a conexão.
func recoverer(logger *slog.Logger) middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				rec := recover()
				if rec == nil {
					return
				}
				if err, ok := rec.(error); ok && errors.Is(err, http.ErrAbortHandler) {
					panic(rec) // sinal do net/http para abortar silenciosamente
				}
				logger.Error("panic no handler",
					"err", rec,
					"path", r.URL.Path,
					"request_id", requestID(r.Context()),
					"stack", string(debug.Stack()),
				)
				http.Error(w, "erro interno", http.StatusInternalServerError)
			}()
			next.ServeHTTP(w, r)
		})
	}
}

// statusWriter captura status e tamanho da resposta para o log.
type statusWriter struct {
	http.ResponseWriter
	status int
	bytes  int
}

// WriteHeader guarda o primeiro status escrito e repassa.
func (w *statusWriter) WriteHeader(code int) {
	if w.status == 0 {
		w.status = code
	}
	w.ResponseWriter.WriteHeader(code)
}

// Write conta os bytes e assume 200 quando o handler não chamou WriteHeader.
func (w *statusWriter) Write(b []byte) (int, error) {
	if w.status == 0 {
		w.status = http.StatusOK
	}
	n, err := w.ResponseWriter.Write(b)
	w.bytes += n
	return n, err
}

// Unwrap expõe o writer original para http.ResponseController.
func (w *statusWriter) Unwrap() http.ResponseWriter { return w.ResponseWriter }

// Command weeklly é o servidor web do planejador semanal.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	// Fusos embutidos no binário: o container final (distroless) e o Windows
	// não têm banco de fusos do sistema.
	_ "time/tzdata"

	"github.com/AdeptLabsDev/weeklly/internal/config"
	"github.com/AdeptLabsDev/weeklly/internal/server"
	"github.com/AdeptLabsDev/weeklly/internal/store"
	"github.com/AdeptLabsDev/weeklly/web"
)

// version é injetada no build: -ldflags "-X main.version=...".
var version = "dev"

func main() {
	os.Exit(realMain())
}

// realMain devolve o código de saída em vez de chamar os.Exit, para que os
// defers rodem.
func realMain() int {
	healthcheck := flag.Bool("healthcheck", false, "consulta /healthz do processo local e sai com 0 ou 1")
	showVersion := flag.Bool("version", false, "imprime a versão e sai")
	flag.Parse()

	if *showVersion {
		fmt.Println("weeklly", version)
		return 0
	}

	cfg, err := config.Load(os.Getenv)
	if err != nil {
		fmt.Fprintln(os.Stderr, "configuração inválida:", err)
		return 2
	}

	if *healthcheck {
		return runHealthcheck(cfg.Addr)
	}

	logger := newLogger(cfg)
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := run(ctx, cfg, logger); err != nil {
		logger.Error("encerrando com erro", "err", err)
		return 1
	}
	return 0
}

func run(ctx context.Context, cfg config.Config, logger *slog.Logger) (err error) {
	st, err := store.Open(ctx, cfg.DBPath)
	if err != nil {
		return fmt.Errorf("abrindo banco: %w", err)
	}
	defer func() { err = errors.Join(err, st.Close()) }()

	if err := st.Migrate(ctx); err != nil {
		return fmt.Errorf("migrando banco: %w", err)
	}

	var webFS fs.FS = web.FS
	if cfg.IsDev() {
		webFS = os.DirFS("web")
	}

	srv, err := server.New(server.Options{
		Config:  cfg,
		Logger:  logger,
		Store:   st,
		Web:     webFS,
		Version: version,
	})
	if err != nil {
		return fmt.Errorf("montando servidor: %w", err)
	}

	httpServer := &http.Server{
		Addr:              cfg.Addr,
		Handler:           srv,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       120 * time.Second,
		MaxHeaderBytes:    64 << 10,
		ErrorLog:          slog.NewLogLogger(logger.Handler(), slog.LevelWarn),
	}

	serveErr := make(chan error, 1)
	go func() {
		logger.Info("weeklly no ar", "addr", cfg.Addr, "env", cfg.Env, "version", version, "db", cfg.DBPath)
		serveErr <- httpServer.ListenAndServe()
	}()

	select {
	case err := <-serveErr:
		return fmt.Errorf("servidor http: %w", err)
	case <-ctx.Done():
	}

	logger.Info("sinal recebido, encerrando")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("encerrando servidor: %w", err)
	}
	if err := <-serveErr; !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("servidor http: %w", err)
	}
	return nil
}

func newLogger(cfg config.Config) *slog.Logger {
	if cfg.IsDev() {
		return slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug}))
	}
	return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
}

// runHealthcheck atende o HEALTHCHECK do container: a imagem final não tem
// shell nem curl, então o próprio binário faz a consulta.
func runHealthcheck(addr string) int {
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		fmt.Fprintln(os.Stderr, "healthcheck:", err)
		return 1
	}
	if host == "" || host == "0.0.0.0" || host == "::" {
		host = "127.0.0.1"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://"+net.JoinHostPort(host, port)+"/healthz", nil)
	if err != nil {
		fmt.Fprintln(os.Stderr, "healthcheck:", err)
		return 1
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Fprintln(os.Stderr, "healthcheck:", err)
		return 1
	}
	_ = resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		fmt.Fprintln(os.Stderr, "healthcheck:", resp.Status)
		return 1
	}
	return 0
}

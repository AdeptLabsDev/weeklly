// Package server monta o http.Handler do weeklly: rotas, middlewares de
// segurança, sessão e observabilidade, templates e arquivos estáticos.
package server

import (
	"fmt"
	"html/template"
	"io/fs"
	"log/slog"
	"net/http"
	"time"

	"github.com/AdeptLabsDev/weeklly/internal/auth"
	"github.com/AdeptLabsDev/weeklly/internal/config"
	"github.com/AdeptLabsDev/weeklly/internal/store"
)

// Options é tudo que o servidor precisa de fora.
type Options struct {
	Config config.Config
	Logger *slog.Logger
	Store  *store.Store
	// Web é a raiz com templates/ e static/ (embutida ou lida do disco).
	Web     fs.FS
	Version string
	// Google é o cliente do login. Nil e credenciais configuradas: New monta
	// o cliente real. Nil sem credenciais: o login fica desligado.
	Google *auth.Google
}

// Server é o handler HTTP raiz.
type Server struct {
	opts    Options
	views   *views
	assets  *assets
	handler http.Handler
	// now é injetável para que os testes fixem "hoje".
	now func() time.Time
}

// New valida as opções, faz o parse dos templates e monta a cadeia de handlers.
func New(opts Options) (*Server, error) {
	if opts.Logger == nil || opts.Store == nil || opts.Web == nil {
		return nil, fmt.Errorf("server.New: Logger, Store e Web são obrigatórios")
	}
	if opts.Google == nil && opts.Config.Google.Configured() {
		g := opts.Config.Google
		opts.Google = auth.NewGoogle(g.ClientID, g.ClientSecret, opts.Config.BaseURL+callbackPath, auth.GoogleEndpoints, nil)
	}

	assets, err := newAssets(opts.Web, opts.Config.IsDev())
	if err != nil {
		return nil, fmt.Errorf("arquivos estáticos: %w", err)
	}
	views, err := newViews(opts.Web, opts.Config.IsDev(), template.FuncMap{
		"asset": assets.URL,
		"dict":  dict,
	})
	if err != nil {
		return nil, fmt.Errorf("templates: %w", err)
	}

	s := &Server{opts: opts, views: views, assets: assets, now: time.Now}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /{$}", s.handleHome)
	mux.HandleFunc("GET /semanas", s.handleHub)
	mux.HandleFunc("GET /semanas/nova", s.handleNewWeekPage)
	mux.HandleFunc("POST /semanas", s.handleCreateWeek)
	mux.HandleFunc("POST /semanas/ordem", s.handleOrder)
	mux.HandleFunc("GET /semana/{id}", s.handleBoard)
	mux.HandleFunc("GET /semana/{id}/renomear", s.handleRenamePage)
	mux.HandleFunc("POST /semana/{id}/renomear", s.handleRename)
	mux.HandleFunc("POST /semana/{id}/duplicar", s.handleDuplicate)
	mux.HandleFunc("GET /semana/{id}/excluir", s.handleDeletePage)
	mux.HandleFunc("POST /semana/{id}/excluir", s.handleDelete)
	mux.HandleFunc("POST /semana/{id}/tarefas", s.handleAddTask)
	mux.HandleFunc("POST /tarefas/{id}/editar", s.handleEditTask)
	mux.HandleFunc("POST /tarefas/{id}/concluir", s.handleDoneTask)
	mux.HandleFunc("POST /tarefas/{id}/mover", s.handleMoveTask)
	mux.HandleFunc("POST /tarefas/{id}/excluir", s.handleDeleteTask)
	mux.HandleFunc("POST /tema", s.handleTheme)
	mux.HandleFunc("POST /idioma", s.handleLanguage)
	mux.HandleFunc("POST /cursor", s.handleCursor)
	mux.HandleFunc("POST /cor", s.handleAccent)
	mux.HandleFunc("GET /entrar/google", s.handleLoginStart)
	mux.HandleFunc("GET "+callbackPath, s.handleLoginCallback)
	mux.HandleFunc("POST /sair", s.handleLogout)
	mux.HandleFunc("GET /healthz", s.handleHealth)
	mux.Handle("GET /static/", assets)
	mux.HandleFunc("/", s.handleNotFound)

	// A cadeia lê de baixo para cima: recover envolve tudo, logging mede tudo,
	// os headers de segurança valem para toda resposta (inclusive 404), a
	// proteção cross-origin bloqueia escritas vindas de outros sites, e a
	// sessão é carregada uma vez por requisição.
	var h http.Handler = mux
	h = s.withSession(h)
	h = http.NewCrossOriginProtection().Handler(h)
	h = securityHeaders(opts.Config)(h)
	h = logging(opts.Logger)(h)
	h = recoverer(opts.Logger)(h)
	s.handler = h

	return s, nil
}

// ServeHTTP implementa http.Handler.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.handler.ServeHTTP(w, r)
}

// dict monta um mapa dentro do template, para passar mais de um valor a uma
// parcial: {{template "x" (dict "A" 1 "B" 2)}}.
func dict(pairs ...any) (map[string]any, error) {
	if len(pairs)%2 != 0 {
		return nil, fmt.Errorf("dict: número ímpar de argumentos")
	}
	m := make(map[string]any, len(pairs)/2)
	for i := 0; i < len(pairs); i += 2 {
		key, ok := pairs[i].(string)
		if !ok {
			return nil, fmt.Errorf("dict: chave %v não é string", pairs[i])
		}
		m[key] = pairs[i+1]
	}
	return m, nil
}

package server

import (
	"bytes"
	"fmt"
	"html/template"
	"io/fs"
	"net/http"
	"path"
	"strings"
)

// views guarda um template por página. Cada página é o layout clonado mais o
// arquivo em templates/pages/, então "title" e "content" de uma página nunca
// vazam para outra.
//
// html/template faz escape contextual: o conteúdo do usuário nunca vira HTML.
type views struct {
	fsys  fs.FS
	dev   bool
	funcs template.FuncMap
	base  *template.Template
	pages map[string]*template.Template
}

func newViews(fsys fs.FS, dev bool, funcs template.FuncMap) (*views, error) {
	v := &views{fsys: fsys, dev: dev, funcs: funcs}
	base, pages, err := v.parse()
	if err != nil {
		return nil, err
	}
	v.base, v.pages = base, pages
	return v, nil
}

func (v *views) parse() (*template.Template, map[string]*template.Template, error) {
	base, err := template.New("layout.html").Funcs(v.funcs).ParseFS(v.fsys, "templates/layout.html", "templates/partials/*.html")
	if err != nil {
		return nil, nil, fmt.Errorf("layout: %w", err)
	}
	files, err := fs.Glob(v.fsys, "templates/pages/*.html")
	if err != nil {
		return nil, nil, err
	}
	if len(files) == 0 {
		return nil, nil, fmt.Errorf("nenhuma página em templates/pages")
	}

	pages := make(map[string]*template.Template, len(files))
	for _, file := range files {
		name := strings.TrimSuffix(path.Base(file), ".html")
		clone, err := base.Clone()
		if err != nil {
			return nil, nil, err
		}
		page, err := clone.ParseFS(v.fsys, file)
		if err != nil {
			return nil, nil, fmt.Errorf("página %s: %w", name, err)
		}
		pages[name] = page
	}
	return base, pages, nil
}

// partial renderiza um template das parciais (por exemplo, uma tarefa) para
// devolver em JSON: o mesmo markup do servidor, encaixado pelo script.
func (v *views) partial(name string, data any) (string, error) {
	base := v.base
	if v.dev {
		var err error
		if base, _, err = v.parse(); err != nil {
			return "", err
		}
	}
	var buf bytes.Buffer
	if err := base.ExecuteTemplate(&buf, name, data); err != nil {
		return "", fmt.Errorf("renderizando parcial %s: %w", name, err)
	}
	return buf.String(), nil
}

// render executa a página em memória e só então escreve a resposta: um erro
// de template vira 500 limpo, nunca meia página.
func (v *views) render(w http.ResponseWriter, status int, page string, data any) error {
	pages := v.pages
	if v.dev {
		var err error
		if _, pages, err = v.parse(); err != nil {
			return err
		}
	}
	t, ok := pages[page]
	if !ok {
		return fmt.Errorf("página desconhecida: %s", page)
	}

	var buf bytes.Buffer
	if err := t.ExecuteTemplate(&buf, "layout.html", data); err != nil {
		return fmt.Errorf("renderizando %s: %w", page, err)
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	// Páginas dependem da sessão: nenhum cache compartilhado pode guardá-las.
	w.Header().Set("Cache-Control", "private, no-cache")
	w.WriteHeader(status)
	_, err := buf.WriteTo(w)
	return err
}

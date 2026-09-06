package server

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/fs"
	"mime"
	"net/http"
	"strings"
)

func init() {
	// A tabela embutida do Go não conhece woff2, e a imagem distroless não
	// tem /etc/mime.types para consultar.
	if err := mime.AddExtensionType(".woff2", "font/woff2"); err != nil {
		panic(err)
	}
}

const staticPrefix = "/static/"

// assets serve web/static com cache correto: em produção cada arquivo tem um
// hash do conteúdo calculado no boot, a URL leva o hash em ?v= e a resposta
// pode ser guardada para sempre. Em desenvolvimento nada é guardado.
type assets struct {
	fsys   fs.FS
	dev    bool
	hashes map[string]string
}

func newAssets(fsys fs.FS, dev bool) (*assets, error) {
	a := &assets{fsys: fsys, dev: dev, hashes: map[string]string{}}
	if dev {
		return a, nil
	}
	err := fs.WalkDir(fsys, "static", func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		data, err := fs.ReadFile(fsys, p)
		if err != nil {
			return err
		}
		sum := sha256.Sum256(data)
		a.hashes[strings.TrimPrefix(p, "static/")] = hex.EncodeToString(sum[:6])
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("indexando static: %w", err)
	}
	return a, nil
}

// URL é a função "asset" dos templates: "app.css" vira "/static/app.css?v=…".
func (a *assets) URL(name string) string {
	if h, ok := a.hashes[name]; ok {
		return staticPrefix + name + "?v=" + h
	}
	return staticPrefix + name
}

func (a *assets) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimPrefix(r.URL.Path, staticPrefix)
	if name == "" || !fs.ValidPath(name) {
		http.NotFound(w, r)
		return
	}
	full := "static/" + name
	info, err := fs.Stat(a.fsys, full)
	if err != nil || info.IsDir() {
		http.NotFound(w, r)
		return
	}

	switch h := a.hashes[name]; {
	case a.dev:
		w.Header().Set("Cache-Control", "no-cache")
	case h != "" && r.URL.Query().Get("v") == h:
		w.Header().Set("ETag", `"`+h+`"`)
		w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	case h != "":
		w.Header().Set("ETag", `"`+h+`"`)
		w.Header().Set("Cache-Control", "public, max-age=3600")
	}
	// name passou por fs.ValidPath (sem ".." nem caminho absoluto) e a.fsys é
	// um FS raiz em web/: não há como sair de static/.
	http.ServeFileFS(w, r, a.fsys, full) //nolint:gosec // G703: caminho validado acima
}

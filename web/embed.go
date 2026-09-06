// Package web embute os templates e os arquivos estáticos no binário.
//
// Em desenvolvimento o servidor lê os mesmos diretórios do disco
// (os.DirFS("web")) para recarregar templates e CSS sem recompilar.
package web

import "embed"

// FS contém templates/ e static/. O prefixo all: inclui arquivos iniciados
// por ponto ou sublinhado, que o embed ignora por padrão.
//
//go:embed all:templates all:static
var FS embed.FS

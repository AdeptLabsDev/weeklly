package server

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"html/template"
	"net/http"
	"strings"

	"github.com/AdeptLabsDev/weeklly/internal/config"
	"github.com/AdeptLabsDev/weeklly/internal/i18n"
	"github.com/AdeptLabsDev/weeklly/internal/week"
)

const (
	privateRobots = "noindex, nofollow"
	publicRobots  = "index, follow, max-image-preview:large"
)

// publicPage é uma página de apresentação: um template e a URL de cada
// idioma. Um idioma novo entra em i18n.All e nos mapas de caminhos; a rota,
// o seletor de idiomas, o hreflang e o sitemap seguem daí.
type publicPage struct {
	template    string
	paths       map[i18n.Lang]string
	title       string // chave i18n do <title>
	description string // chave i18n da meta description
}

var (
	landingPaths = map[i18n.Lang]string{i18n.PT: "/planejador-semanal", i18n.EN: "/en/weekly-planner"}
	faqPaths     = map[i18n.Lang]string{i18n.PT: "/perguntas-frequentes", i18n.EN: "/en/faq"}

	landingPage = publicPage{template: "landing", paths: landingPaths, title: "seo.title", description: "seo.description"}
	faqPage     = publicPage{template: "faq", paths: faqPaths, title: "faq.title", description: "faq.description"}
	publicPages = []publicPage{landingPage, faqPage}

	// Locale do Open Graph e imagem social por idioma.
	ogLocales = map[i18n.Lang]string{i18n.PT: "pt_BR", i18n.EN: "en_US"}
	ogImages  = map[i18n.Lang]string{i18n.PT: "og-planner-pt.png", i18n.EN: "og-planner-en.png"}

	// faqTopics é a ordem das perguntas; cada uma tem faq.<tópico>.question e
	// .answer nos catálogos. A página e o JSON-LD vêm da mesma lista.
	faqTopics = []string{"free", "account", "dates", "saved", "repeat", "time", "mobile", "language"}

	// playStoreURL é a página do app no Google Play. Vazia até o app ser
	// publicado: o selo aparece marcado "em breve", sem link.
	playStoreURL = ""
)

type alternateLink struct {
	Lang string
	URL  string
}

// langLink é um idioma no seletor da página pública: nome no próprio idioma
// e a URL da tradução desta mesma página.
type langLink struct {
	Tag     string
	Name    string
	URL     string
	Current bool
}

type faqItem struct {
	Question string
	Answer   string
}

// sceneDay é uma coluna da cena de "como funciona" (viewBox de 560 por 380):
// sete colunas de 64 com 8 de vão, o nome centralizado em cada uma.
type sceneDay struct {
	Name  string
	X     int
	Label int
}

type seoView struct {
	Description  string
	Canonical    string
	Robots       string
	Locale       string
	OtherLocales []string
	Image        string
	Alternates   []alternateLink
	Languages    []langLink
	// Landing e FAQ são as URLs, no idioma da página, para a barra e o rodapé.
	Landing string
	FAQ     string
	// PlayStore é o link do app no Google Play; vazio enquanto não há app.
	PlayStore string
	// Days são os dias da cena de "como funciona": nome curto e posição.
	Days []sceneDay
	// Questions preenche a página de perguntas frequentes.
	Questions []faqItem
	JSONLD    template.JS
}

// publicPath identifica somente documentos independentes da sessão.
func publicPath(path string) bool {
	if path == "/robots.txt" || path == "/sitemap.xml" {
		return true
	}
	for _, p := range publicPages {
		for _, candidate := range p.paths {
			if path == candidate {
				return true
			}
		}
	}
	return false
}

// publicPaths lista as URLs públicas na ordem das páginas e dos idiomas,
// para o roteador e o sitemap.
func publicPaths() []string {
	var out []string
	for _, p := range publicPages {
		for _, lang := range i18n.All {
			out = append(out, p.paths[lang])
		}
	}
	return out
}

func (s *Server) searchIndexing() bool {
	cfg := s.opts.Config
	return cfg.SearchIndexing && cfg.Env == config.Production && strings.HasPrefix(cfg.BaseURL, "https://")
}

// searchHeaders fecha a indexação por padrão, incluindo redirects, JSON e
// erros do mux. Só o handler de uma página pública pode abri-la.
func (s *Server) searchHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Robots-Tag", privateRobots)
		w.Header().Set("Cache-Control", "private, no-store")
		// Imagens e fontes públicas precisam continuar disponíveis aos robôs.
		if s.searchIndexing() && strings.HasPrefix(r.URL.Path, staticPrefix) {
			w.Header().Del("X-Robots-Tag")
		}
		next.ServeHTTP(w, r)
	})
}

// registerPublic liga cada URL pública ao seu template e idioma.
func (s *Server) registerPublic(mux *http.ServeMux) {
	for _, p := range publicPages {
		for _, lang := range i18n.All {
			mux.HandleFunc("GET "+p.paths[lang], s.handlePublic(p, lang))
		}
	}
}

// handlePublic renderiza uma página de apresentação. O idioma vem da URL;
// o tema segue o cookie, como no app, por ser preferência de apresentação,
// não dado pessoal. Cursor, cor e sessão ficam de fora.
func (s *Server) handlePublic(p publicPage, lang i18n.Lang) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := i18n.L(lang)
		meta := seoView{
			Description: l.T(p.description), Robots: privateRobots, Locale: ogLocales[lang],
			Landing: landingPaths[lang], FAQ: faqPaths[lang], PlayStore: playStoreURL,
		}
		for _, other := range i18n.All {
			if other != lang {
				meta.OtherLocales = append(meta.OtherLocales, ogLocales[other])
			}
		}
		for i, d := range week.All() {
			x := 32 + i*72
			meta.Days = append(meta.Days, sceneDay{Name: l.WeekdayShort(d), X: x, Label: x + 32})
		}
		for _, o := range l.Languages() {
			meta.Languages = append(meta.Languages, langLink{Tag: string(o.Tag), Name: o.Name, URL: p.paths[o.Tag], Current: o.Current})
		}
		if p.template == faqPage.template {
			for _, topic := range faqTopics {
				meta.Questions = append(meta.Questions, faqItem{Question: l.T("faq." + topic + ".question"), Answer: l.T("faq." + topic + ".answer")})
			}
		}
		if s.searchIndexing() {
			meta.Robots = publicRobots
		}
		if base := s.opts.Config.BaseURL; base != "" {
			meta.Canonical = base + p.paths[lang]
			meta.Image = base + s.assets.URL(ogImages[lang])
			for _, tag := range i18n.All {
				meta.Alternates = append(meta.Alternates, alternateLink{Lang: string(tag), URL: base + p.paths[tag]})
			}
			meta.Alternates = append(meta.Alternates, alternateLink{Lang: "x-default", URL: base + p.paths[i18n.PT]})
			data, err := json.Marshal(map[string]any{"@context": "https://schema.org", "@graph": s.publicGraph(p, l, meta, base)})
			if err != nil {
				s.serverError(w, r, err)
				return
			}
			// encoding/json escapa <, >, & e separadores Unicode. A única
			// conversão para JS confiável recebe esses bytes, nunca HTML livre.
			meta.JSONLD = template.JS(data) //nolint:gosec // G203: somente JSON serializado e escapado pela stdlib
			// Autoriza só este bloco de dados, sem unsafe-inline nem scripts de terceiros.
			sum := sha256.Sum256(data)
			hash := base64.StdEncoding.EncodeToString(sum[:])
			csp := w.Header().Get("Content-Security-Policy")
			w.Header().Set("Content-Security-Policy", strings.Replace(csp, "script-src 'self'", "script-src 'self' 'sha256-"+hash+"'", 1))
		}
		s.render(w, r, http.StatusOK, p.template, page{
			Title: l.T(p.title), Public: true, SEO: meta, L: l,
			BodyClass: "is-landing",
		})
	}
}

// publicGraph monta o JSON-LD: a página, o aplicativo que ela apresenta e,
// na página de perguntas, as perguntas e respostas (FAQPage).
func (s *Server) publicGraph(p publicPage, l i18n.Locale, meta seoView, base string) []any {
	webPage := map[string]any{
		"@type": "WebPage", "@id": meta.Canonical + "#webpage",
		"url": meta.Canonical, "name": l.T(p.title),
		"description": meta.Description, "inLanguage": string(l.Lang()),
		"mainEntity": map[string]string{"@id": base + "/#app"},
	}
	if len(meta.Questions) > 0 {
		questions := make([]any, 0, len(meta.Questions))
		for _, q := range meta.Questions {
			questions = append(questions, map[string]any{
				"@type": "Question", "name": q.Question,
				"acceptedAnswer": map[string]any{"@type": "Answer", "text": q.Answer},
			})
		}
		webPage["@type"] = []string{"WebPage", "FAQPage"}
		webPage["mainEntity"] = questions
	}
	return []any{
		webPage,
		map[string]any{
			"@type": "WebApplication", "@id": base + "/#app",
			"name": l.T("app.name"), "url": base + "/",
			"description": l.T("seo.description"), "inLanguage": langTags(),
			"applicationCategory": "ProductivityApplication",
			"operatingSystem":     "Any", "image": meta.Image,
		},
	}
}

func langTags() []string {
	out := make([]string, 0, len(i18n.All))
	for _, lang := range i18n.All {
		out = append(out, string(lang))
	}
	return out
}

func (s *Server) handleRobots(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	if !s.searchIndexing() {
		_, _ = fmt.Fprint(w, "User-agent: *\nDisallow: /\n")
		return
	}
	// O crawler precisa ler noindex nas páginas privadas. Disallow não é
	// controle de acesso; o isolamento dos quadros continua na sessão.
	_, _ = fmt.Fprintf(w, "User-agent: *\nDisallow:\n\nSitemap: %s/sitemap.xml\n", s.opts.Config.BaseURL)
}

type sitemapURL struct {
	Location string `xml:"loc"`
}

type sitemap struct {
	XMLName xml.Name     `xml:"urlset"`
	XMLNS   string       `xml:"xmlns,attr"`
	URLs    []sitemapURL `xml:"url"`
}

func (s *Server) handleSitemap(w http.ResponseWriter, r *http.Request) {
	doc := sitemap{XMLNS: "http://www.sitemaps.org/schemas/sitemap/0.9"}
	if s.searchIndexing() {
		for _, path := range publicPaths() {
			doc.URLs = append(doc.URLs, sitemapURL{Location: s.opts.Config.BaseURL + path})
		}
	}
	data, err := xml.Marshal(doc)
	if err != nil {
		s.serverError(w, r, err)
		return
	}
	w.Header().Set("Content-Type", "application/xml; charset=utf-8")
	_, _ = w.Write(append([]byte(xml.Header), data...))
}

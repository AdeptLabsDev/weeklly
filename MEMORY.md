# MEMORY.md

Estado vivo do projeto. Limite: 200 linhas. Detalhe longo vai para `docs/topics/`.

## Estado (2026-09-06, guia visual e imagem social)

- Estilo implementado documentado em `design/ESTILO-WEEKLLY.md`, com tokens, tipografia, composição, fontes e distinção entre a silhueta atual e o cartoon futuro. Capturas locais da landing nos dois temas em `design/referencias/`.
- Prompt reutilizável em `design/PROMPTS-IMAGENS.md`; prompt executado em `design/PROMPT-TESTE-WEEKLLY.txt`; imagem explicativa gerada e revisada em `design/social/weeklly-o-que-e.png`. Sem alteração de interface.
- Nesta entrega, `task check` passou por vet e parou no lint com `context loading failed: no go files to analyze`; as etapas seguintes não foram executadas. Nenhuma dependência alterada para contornar o problema.

## Estado (2026-09-06, SEO)

- D23: páginas públicas `/planejador-semanal` e `/en/weekly-planner`, com idioma fixo por URL e sem sessão; `/` continua abrindo onde a pessoa parou. CTAs mantêm idioma ao entrar no app.
- D29: faixa dos aplicativos na landing (`.landing-apps`, token `--color-band`; `<main>` sem `landing-width`, cada seção tem a sua). Selo do Google Play = `.store-badge`; `playStoreURL` vazia em `seo.go` → `<span>` "em breve".
- D28: "como funciona" é guiado pela rolagem sem prender a página: o script lê a posição do centro do `.how` na tela (0 embaixo, 1 em cima; fronteiras 0,38 e 0,6), marca o rádio e põe `--how-progress` no `.how`. Tremor do botão = "Shake" do CSS-Tricks (horizontal, em ciclos). FAQ anima com `::details-content` + `interpolate-size` (só CSS).
- D27: abertura da landing em duas colunas (`.hero-copy` + `.hero-art` com a partial `mascot.html`, silhueta em `currentColor`); `.landing-cta` é o único botão na cor da marca, com `cta-wiggle` no hover. Cartoon da mascote: briefing e prompts em `design/brand/MASCOTE.md`, troca por `<img>` descrita lá.
- D26 (landing definitiva, parte 1): páginas públicas são a lista `publicPages` em `seo.go`; barra e rodapé na partial `public.html`; "como funciona" = rádios `#how-1..3` + cena SVG movida por `.how:has(#how-N:checked)` em `app.css`, avanço automático no `app.js` (IntersectionObserver + `animationend` do `.how-rail-fill`); FAQ em `pages/faq.html` (`/perguntas-frequentes`, `/en/faq`), perguntas da lista `faqTopics` (chaves `faq.<tópico>.question/answer`).
- D25: concluir tarefa anima (classes `is-just-done`/`is-just-undone` postas por `replaceTask`; risco é `background-size` no `.task-title-text`), duplicar tarefa (`POST /tarefas/{id}/duplicar`, ícone à direita do x), landing com botão de idioma abrindo o painel `#language-menu` (posicionado a partir do invocador).
- Mascote (D24): esquilo em `design/brand/squirrel-icon-menu.svg`, na cor principal da marca `#727cf5` (token `--color-brand`, só para marca; começou coral); virou o favicon (SVG + PNG 32 + iPhone 180). Referência da Duolingo seção a seção em `docs/LANDING-REFERENCE.md`, com o plano de landing v3 aguardando três decisões do Miguel (esquilo na abertura, faixa dos sete dias, um botão só na barra).
- Landing refeita pelo Miguel (rodada 7): barra com sol/lua e PT/EN, headline "grátis e sem cadastro", sem demonstração, passos + seis pontos fortes, FAQ por último, rodapé com seis grupos ("em breve" onde não há página). Tema segue cookie; `app.js` carregado como melhoria. Pendente: páginas Sobre/Contato/Termos/Privacidade e links das redes.
- Metadados, canonical/hreflang, JSON-LD com hash CSP e PNGs de compartilhamento. `seo.go` centraliza rotas de descoberta. App e erros recebem noindex; sitemap só tem páginas públicas.
- Indexação desligada por padrão (`WEEKLLY_SEARCH_INDEXING=false`); habilitar exige produção + BaseURL HTTPS oficial. Domínio, deploy, Search Console/Bing e métricas reais seguem pendentes. Pesquisa de skills e plano em `docs/SEO.md`.
- Validação: `task check` completo passou; Playwright completo: 10 aprovados, 2 casos só de desktop ignorados no mobile. Inclui SEO sem JS e em 320px; capturas revisadas. PNGs regeneráveis com `node e2e/scripts/generate-social.mjs` após `task css`.

## Estado (2026-09-06, rodada 5)

- Cursor próprio (D22) é opcional: "Tipo de mouse" nas configurações (cookie `weeklly_cursor`, `data-cursor` no html), desligado por padrão porque o navegador mostra o cursor do sistema em navegações, na troca de tema e no duplo clique. Elemento `partials/cursor.html` (`data-cursor-el`, popover manual no top layer), `html.has-cursor` esconde o do sistema.
- Cor de destaque escolhida pela pessoa (D22): "Cor" nas configurações, cookie `weeklly_accent`, `data-accent` no html, paleta `--accent-*` por tema, `--color-display` nos textos principais e `--color-cursor` no cursor. Padrão "mono".
- Outra sessão do Claude trabalhou em paralelo nesta árvore (SEO: landing, robots, sitemap). Edições nos mesmos arquivos coexistiram; conferir `git status` antes de mexer.
- Idioma nas configurações: linha "Idioma · Português ⌄" abre o popover `#language-menu` ao lado (ou abaixo, no celular), lista de `Locale.Languages()`: idioma novo = constante em `All` + `lang.<tag>` nos catálogos + `Parse`.

## Estado (2026-09-06, rodada 4)

- Dois idiomas (D21): `internal/i18n` com catálogos pt-BR e en, cookie `weeklly_lang`, `Accept-Language` na primeira visita. Menu de configurações na barra (idioma; mais opções depois). Desfazer/refazer na linha de "Hoje é...", por aba, com atalhos.
- Miguel commitou tudo até a rodada 3 ("feat: fundação do projeto e núcleo da Fase 1").

## Estado (2026-09-06, rodada 3)

- Movimento e mecânica (D20): tema varrendo com View Transitions, alça de arrasto para reordenar e mover tarefas entre dias (Alt+setas no teclado), clique em área vazia foca o campo, horário em texto com seletor leve, Recentes/A–Z animado sem recarregar, logo para o hub.
- Rota nova `POST /tarefas/{id}/mover`. e2e cobre arrasto no desktop.

## Estado (2026-09-06, rodada 2)

- Monocromático em dois temas (D17), sol/lua na barra com cookie, Nunito nos nomes e títulos, sistema no resto.
- Tarefas por dia (D18): título, horário opcional, feita; adicionar/editar/concluir/excluir sem recarregar, com formulários que funcionam sem script. Tabela `days` saiu, `tasks` entrou.
- Ações da semana no seletor (D19): renomear, duplicar, excluir; ordem Recentes/A–Z por usuário.
- Credenciais do Google já estão no `.env` do Miguel (não testadas de ponta a ponta com o Google real).
- Pendente do Miguel: cor de destaque. Próximo natural: "hoje" pelo fuso do navegador, reordenar tarefas arrastando, fila offline.

## Estado (2026-09-06, rodada 1)

- Barra de navegação entregue: logo, semana aberta como botão com seletor (Popover API), casinha para o hub, nova semana (diálogo nativo com página de fallback), entrar com Google ou foto de perfil com menu de sair.
- Persistência de semanas, sessão anônima no primeiro "Nova semana", login com Google por OIDC (D14, D15). Credenciais do Google ainda não cadastradas: o botão leva à página "Login ainda não configurado".
- Playwright em `e2e/` sobe o servidor e percorre o fluxo em desktop e Pixel 7; `task e2e`. Capturas em `e2e/screenshots` (D16).
- Próximo: Miguel quer organizar a informação antes das cores. Depois, editor por dia com autosave (Fase 1).

## Estado anterior (2026-09-05)

- Fase 0 construída: repositório, CI, tokens, modelo de dados, tela do quadro com dados fictícios. `task check` verde localmente.
- Primeira tela reprovada pelo Miguel (fim de semana menor, moldura fechada). Refeita como faixa horizontal de sete cartões iguais (D11), com o primeiro JS do projeto.
- Decidido: semana é espaço com nome, sem datas (D12). Esquema 0001 reescrito; `internal/week` sem calendário.
- Segunda rodada de feedback: roda e arrasto travados (culpa do scroll-snap, removido), visual fraco. Skills `frontend-design` e `web-design-guidelines` instaladas e aplicadas (D13): papel quente sobre noite fria, nomes dos dias em serifa, paginador, inércia.
- Falta para fechar a Fase 0: primeiro push com CI verde no GitHub e aprovação visual do Miguel.
- Depois: Fase 1 (editor por dia, autosave, persistência local, fuso do navegador).
- Preview local: `task dev` e abrir http://127.0.0.1:8080.

## Decisões recentes

- D6 Go. D7 stack completa. D8 tokens e tipografia. D9 semana começa na segunda, garantido no banco. D10 layout desktop 5 + 2.
- Migrador próprio em vez de goose: sequencial, up-only, 60 linhas (D7).

## Lições e gotchas

- Windows com Git Bash: heredoc acima de uns 8 KB estoura o limite de linha de comando. Arquivo grande vai pela ferramenta Write.
- Tailwind standalone é binário Bun/glibc: não roda em Alpine (existe variante musl). O estágio de build usa Debian.
- Tailwind v4 emite só as variáveis de `@theme` que detecta em uso; `@theme static` emite todas.
- modernc.org/sqlite: pragmas viajam na URI (`file:…?_pragma=…`) e valem por conexão do pool. Caminho absoluto do Windows vira `/C:/…` na URI.
- SQLite STRICT converte inteiro para TEXT sem erro; o que ele rejeita é BLOB em coluna TEXT.
- `go test -race` no Windows precisa de cgo (gcc). Local: `task test`. CI Linux: `task test:race`.
- Go não conhece `.woff2` na tabela de MIME; registrado em `internal/server/assets.go`.
- Subconjunto latin do Google Fonts cobre o português inteiro; woff2 de 20 a 32 KB por face.
- Docker Desktop fica em `%LOCALAPPDATA%\Programs\DockerDesktop`, não em Program Files, e não sobe sozinho com o Windows.
- Screenshot sem Playwright: `msedge.exe --headless=new --screenshot=… --window-size=…`. Largura mínima real da janela é cerca de 500px: pedir 390 gera a página em 500 e corta a imagem. Para mobile de verdade, Playwright com emulação de dispositivo (Fase 4).
- Grid da casca usa `minmax(0, 1fr)`: coluna `auto`/`1fr` cresce além da viewport se um filho tiver largura mínima intrínseca grande.
- `scroll-snap` numa faixa com rolagem programática anula a roda do mouse (cada passo volta ao encaixe) e puxa o arrasto ao soltar. Faixa livre, sem snap; inércia por rAF.
- Gatilho BEFORE DELETE em `days` não bloqueia o cascade de `weeks`: o SQLite apaga a linha pai antes de propagar.
- `npx skills add <owner/repo@skill> -g -y` instala em `~/.agents/skills` com symlink em `~/.claude/skills`. Avisos "PromptScript" são de outro cliente, não afetam o Claude Code.
- `httptest.Server.Client()` devolve sempre o mesmo objeto: para simular dois navegadores, crie `http.Client` novos compartilhando só o Transport.
- Servidor de preview em segundo plano sobrevive entre rodadas: antes de subir outro, `taskkill //F //IM weeklly.exe`, senão o novo falha no bind e o antigo (com esquema velho) responde.
- Em `/semanas/nova` há dois formulários iguais (página e diálogo fechado): nos testes, restrinja ao `main`.
- Google Fonts, cookies Secure e Playwright: o servidor de teste precisa ser `httptest.NewTLSServer`, senão o jar do Go descarta cookies Secure em http.
- Capturas logo após trocar o tema pegam as transições de 120ms no meio: esperar ~300ms antes do screenshot.
- Numa transação do SQLite, ler (rows abertas) e escrever ao mesmo tempo não é seguro: colete tudo, feche, depois insira (ver `tasksToCopy`).
- Placeholder "Semana padrão…" existe em todo formulário de semana: testes que checam "o nome antigo sumiu" precisam olhar o elemento certo, não a página inteira.
- `setPointerCapture` na faixa faz o `click` ser entregue à faixa (alvo comum de down e up), não ao elemento sob o ponteiro. Cliques "parados" dentro da faixa são tratados no `pointerup` com o alvo guardado no `pointerdown`.
- View Transitions: `::view-transition-new(root)` com `clip-path: inset(0 100% 0 0 → 0)` dá a varredura; `view-transition-name` num botão o anima à parte. `mix-blend-mode: normal` nos snapshots evita o crossfade padrão.
- html/template escapa apóstrofos como `&#39;` no texto: testes que procuram frases em inglês com "isn't" precisam de outra frase ou da forma escapada.
- `data-cursor` no `<html>` é a preferência: o elemento do cursor usa `data-cursor-el`. Um `querySelector("[data-cursor]")` pegava o `<html>` e transladava a página inteira.
- `cursor: url()` pisca ao navegar: o Chromium troca cursor de imagem pelo padrão enquanto a página carrega (`is_loading_`) e só recoloca o da página nova quando o mouse se mexe. Cursor por elemento com `cursor: none` não sofre disso. Chromium computa `cursor: text` (não `auto`) em `<input>`.
- Elemento `position: fixed` nunca fica acima do top layer (diálogos, popovers): para isso ele precisa ser popover, e o último a abrir fica por cima, daí o `raise()` do cursor a cada `toggle`/`showModal`.
- Playwright sem JavaScript: clicar num popover que entra com transição (`@starting-style`) fica em "element is not stable"; esperar ~300 ms antes do clique.
- Popover aninhado (invocador dentro de outro popover) não fecha o pai ao abrir; fechar o pai fecha o filho. `@starting-style` dá a animação de entrada.
- Playwright fora de `e2e/`: `require("C:/.../e2e/node_modules/@playwright/test")` num `.cjs` no scratchpad serve para capturas avulsas (prévias ampliadas com `deviceScaleFactor`).
- Frases do script vão em `data-i18n` (JSON no atributo): um `<script type="application/json">` seria escapado como texto pelo html/template e quebraria.

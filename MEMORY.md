# MEMORY.md

Estado vivo do projeto. Limite: 200 linhas. Detalhe longo vai para `docs/topics/`.

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

# weeklly

Contexto do produto para agentes. O CLAUDE.md global define como o time opera; este define este projeto.

## O que é

Planejador semanal: sete dias em um quadro fixo, texto livre por dia, salvo sozinho, aberto em menos de um segundo. Site primeiro, app depois. Projeto pessoal do Miguel, padrão world-class.

## Documentos que mandam

- `docs/ROADMAP.md`: fases, invariantes do produto, critérios de saída.
- `docs/DECISIONS.md`: toda escolha técnica com razão e data. Dependência nova, mudança de stack ou de arquitetura exige nova entrada.
- `docs/BACKLOG.md`: só tópicos, `[x]` ou `[ ]`. Marcar ao concluir; adicionar linha ao criar etapa.
- `ARCHITECTURE.md`: estrutura, fluxo, segurança, como rodar.
- `MEMORY.md`: estado vivo, lições e gotchas. Atualizar ao fechar um bloco de trabalho.

## Stack cravada

Go com biblioteca padrão (net/http, html/template, slog, embed) · SQLite WAL via modernc, sem cgo · Tailwind v4 CLI standalone, sem Node · JS nativo, sem framework e sem bibliotecas de animação · dark-first com tema claro · monocromático até a cor de destaque ser definida (só `--color-accent` muda) · Nunito só em nomes e títulos, fonte do sistema no resto.

## Regras deste repositório

- Biblioteca padrão primeiro. Dependência nova só com entrada em `docs/DECISIONS.md`.
- Nenhum JavaScript de terceiros no navegador. Nenhum `style=""` inline: a CSP bloqueia.
- A semana não tem datas. O único uso de calendário é "hoje" em `internal/week`. Handler nunca chama `time.Now()` direto: usa `Server.now`.
- Migrações: SQL puro em `internal/store/migrations`, numeradas. Nunca editar uma já aplicada.
- Invariantes do produto vivem no banco (CHECK, gatilhos) e em `internal/week`, não só no handler.
- Segurança: headers em `internal/server/middleware.go`; toda rota nova herda. Sem segredos no repositório.
- Testes acompanham o código: datas, schema, handlers e headers. `task check` verde antes de qualquer entrega.
- Interface em português do Brasil e inglês: todo texto visível vem de `internal/i18n` (chave nos dois catálogos, nunca literal em template, handler ou script). Identificadores em inglês, comentários em português.
- Design: tokens só em `web/styles/app.css` (`@theme`, com o tema claro em `:root[data-theme="light"]`). Componentes em CSS com classes semânticas; utilitários do Tailwind para layout pontual. Nenhuma cor fora dos tokens.
- Toda ação de tarefa e de semana é um formulário que funciona sem script; o script pede JSON no mesmo POST e recebe a parcial renderizada pelo servidor. Markup só nos templates.

## Comandos

`task dev` · `task check` · `task e2e` (Playwright, Node só em `e2e/`) · `task build` · `task docker:up` · `task --list`

## Glossário

Semana: espaço de planejamento com nome, sete dias de segunda a domingo, sem datas, id opaco. Dia: uma posição na semana e sua lista de tarefas. Tarefa: título, horário opcional, feita ou não, em ordem manual. Quadro: faixa horizontal dos sete cartões iguais. Hub: página com as semanas por uso recente; o seletor na barra é o atalho. Hoje: o dia da semana atual no fuso do usuário. Visitante: quem tem sessão anônima (criou semana sem entrar). Entrar: ligar a sessão a uma conta Google; só quem entrou vê a foto na barra.

## Skills de design

`frontend-design` (Anthropic) antes de desenhar tela nova; `web-design-guidelines` (Vercel) para revisar template e CSS. As duas estão instaladas globalmente. O `/hm-designer` do Miguel continua sendo a validação final.

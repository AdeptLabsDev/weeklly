# Arquitetura

Como o weeklly está organizado, como rodar e por que cada peça existe. As razões completas de cada escolha estão em [docs/DECISIONS.md](docs/DECISIONS.md).

## Stack

| Camada | Escolha | Por quê, em uma linha |
|---|---|---|
| Servidor | Go 1.27, biblioteca padrão (`net/http`, `html/template`, `log/slog`, `embed`) | Um binário estático, escape contextual de HTML, CSRF nativo, compatibilidade por décadas (D6) |
| Banco | SQLite em modo WAL via `modernc.org/sqlite` (Go puro) | Um arquivo, um processo, sem cgo nem toolchain C (D5, D7) |
| CSS | Tailwind v4 pelo CLI standalone, sem Node no repositório. Tokens monocromáticos em `@theme`, tema claro por `data-theme` no `<html>` | Binário com versão e SHA-256 fixados (D3); paleta e temas em D17 |
| JS | Um módulo ES nativo (`web/static/app.js`): diálogos, tema, faixa (centralizar hoje, roda, arrasto com inércia, teclado, paginador) e tarefas sem recarregar. Sem framework | D4, D11, D18 |
| Container | Multi-stage, imagem final distroless static, usuário 65532 | Sem shell, sem root, binário estático |

## Como rodar

Pré-requisitos: Go 1.27 ou mais novo e as ferramentas abaixo (todas via `go install`, ficam em `~/go/bin`):

```
go install github.com/go-task/task/v3/cmd/task@latest
go install github.com/air-verse/air@latest
go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest
go install golang.org/x/vuln/cmd/govulncheck@latest
```

```
cp .env.example .env   # WEEKLLY_ENV=development
task dev               # http://127.0.0.1:8080 com recarga automática de Go, templates e CSS
task check             # vet, lint, testes, vulnerabilidades e build: o mesmo que o CI
task run               # binário de produção em bin/
task docker:up         # imagem de produção em http://127.0.0.1:8080, dados no volume weeklly-data
task --list            # tudo
```

A primeira execução baixa o CLI do Tailwind para `.tools/` (fora do git) e confere o SHA-256 antes de usar.

Testes ponta a ponta, uma vez `task e2e:install` (Node 18 ou mais novo, só para `e2e/`) e depois `task e2e`. As capturas de tela ficam em `e2e/screenshots`.

### Login com o Google

1. No Google Cloud Console, em APIs e serviços, crie um "ID do cliente OAuth" do tipo aplicativo da Web.
2. Cadastre a URL de retorno: `{WEEKLLY_BASE_URL}/entrar/google/callback`. Para desenvolvimento, `http://127.0.0.1:8080/entrar/google/callback`.
3. Coloque `WEEKLLY_GOOGLE_CLIENT_ID` e `WEEKLLY_GOOGLE_CLIENT_SECRET` no `.env` (ou no ambiente do container). Em produção, `WEEKLLY_BASE_URL` precisa ser https.

Sem as credenciais o produto funciona sem login: as semanas ficam no navegador pela sessão anônima, e entrar depois reivindica tudo (D14).

## Portas e serviços

| Serviço | Porta | Onde |
|---|---|---|
| weeklly | 8080 | `127.0.0.1:8080` local. No container escuta em `:8080`, publicado só em `127.0.0.1` |

Não há outros serviços: o banco é um arquivo. Em produção um proxy com TLS (Caddy) fica na frente; decisão final na Fase 5.

## Estrutura

```
cmd/weeklly/                entrypoint: flags, config, boot do banco, servidor HTTP, shutdown gracioso
internal/config/            leitura e validação do ambiente (WEEKLLY_*)
internal/ids/               identificadores opacos (16 caracteres) para URLs e chaves
internal/week/              modelo do produto: Week com nome, Weekday (segunda a domingo), hoje por fuso
internal/auth/              login com o Google: OpenID Connect, PKCE, verificação do id_token, provedor falso para testes
internal/store/             SQLite: pools, pragmas, migrações; usuários, semanas, tarefas, sessões
internal/store/migrations/  SQL puro, numerado, aplicado no boot
internal/server/            http.Handler: rotas, middlewares, sessão, login, templates, assets, view models
web/templates/              layout.html, partials/ (barra, logo, formulário) e pages/*.html
e2e/                        Playwright: testes ponta a ponta e capturas de tela (Node só aqui)
web/styles/app.css          fonte do Tailwind: fontes, tokens (@theme), componentes
web/static/                 fontes woff2 e favicon; app.css é gerado e fica fora do git
tools/tailwind/             downloader do CLI do Tailwind com versão e checksum fixados
docs/                       ROADMAP, DECISIONS, BACKLOG
```

Fronteiras: `week` não conhece HTTP nem banco. `store` não conhece HTTP. `server` usa os dois só pela API pública. `cmd` apenas monta as peças.

## Fluxo de uma requisição

recover → log (request id, status, duração) → headers de segurança → CrossOriginProtection → sessão (cookie → usuário no contexto) → mux → handler → view model → template renderizado em buffer → resposta.

Rotas da Fase 0:

| Rota | Faz |
|---|---|
| `GET /` | Abre onde a pessoa parou: redireciona para a última semana aberta (ou a mais recente); sem semanas, mostra o hub |
| `GET /semanas` | Hub: as semanas do usuário, as usadas mais recentemente primeiro |
| `GET /semanas/nova` | Formulário de nova semana em página (o mesmo abre em diálogo com JavaScript) |
| `POST /semanas` | Cria a semana; um visitante sem sessão ganha usuário anônimo e sessão aqui |
| `GET /semana/{id}` | O quadro. Semana de outra pessoa, inexistente ou sem sessão: 404 |
| `GET`/`POST /semana/{id}/renomear` | Renomear (página sem script e POST) |
| `POST /semana/{id}/duplicar` | Cria "Nome (cópia)" com as tarefas e abre a cópia |
| `GET`/`POST /semana/{id}/excluir` | Confirmação (página sem script) e exclusão |
| `POST /semana/{id}/tarefas` | Cria uma tarefa (`weekday`, `title`, `time`). Com `Accept: application/json` devolve a tarefa renderizada; sem, volta ao dia |
| `POST /tarefas/{id}/editar`, `/concluir`, `/mover`, `/excluir` | Título e horário, feita ou não, posição e dia (`weekday`, `position`), apagar. Mesmo contrato JSON ou redirecionamento |
| `POST /semanas/ordem` | Ordem da lista de semanas (`recent` ou `name`), guardada no usuário |
| `POST /tema` | Tema (`dark` ou `light`) em cookie legível pelo script |
| `GET /entrar/google` | Início do login: cookie curto com state, nonce e PKCE; redireciona ao Google. Sem credenciais, página explicando |
| `GET /entrar/google/callback` | Retorno do Google: confere state, troca o código, verifica o `id_token`, liga ou funde a conta, abre a sessão |
| `POST /sair` | Encerra a sessão deste navegador |
| `GET /healthz` | JSON com status e versão; consulta real ao banco; 503 se o banco não responde |
| qualquer outra | 404 com a voz do produto |
| `GET /static/…` | Arquivos estáticos com hash de conteúdo em `?v=` e cache imutável |

## Segurança

- **CSP `default-src 'none'`** com liberação explícita só para `'self'`. Sem estilos inline, sem terceiros. Um `style=""` ou um script externo quebra a página: é intencional.
- **CSRF** por `http.CrossOriginProtection` (baseado em Sec-Fetch-Site), da biblioteca padrão.
- **Headers** em toda resposta, inclusive 404: nosniff, X-Frame-Options DENY, Referrer-Policy, Permissions-Policy, COOP, CORP. HSTS fora de development.
- **Escape de HTML** contextual pelo `html/template`: texto do usuário nunca vira markup.
- **Static**: caminho validado por `fs.ValidPath` sobre um FS raiz embutido; não existe caminho para fora de `web/static`.
- **Timeouts** no `http.Server` (header, read, write, idle) e limite de tamanho de header.
- **Banco**: tabelas STRICT, CHECKs e gatilhos garantem as invariantes do produto dentro do SQLite; foreign keys ligadas; pool de leitura em `query_only`.
- **Padrões seguros**: escuta em localhost e modo production por padrão; development é opt-in.
- **Container**: multi-stage, distroless static, `USER 65532`, `read_only`, `cap_drop: ALL`, `no-new-privileges`, porta publicada só em 127.0.0.1.
- **Supply chain**: `go mod verify` e `govulncheck` no CI; Tailwind com SHA-256 fixado por plataforma; zero dependência de JavaScript.
- **Sessões**: token aleatório de 32 bytes no cookie, só o hash no banco; HttpOnly, SameSite=Lax, Secure fora de development; 90 dias renovados no uso. Toda consulta de semana filtra pelo usuário da sessão.
- **Login**: OpenID Connect com PKCE, state e nonce em cookie curto; `id_token` verificado contra o JWKS do Google (assinatura, emissor, audiência, validade, nonce); só e-mail confirmado. Cancelamento e falhas viram páginas explicando, nunca sessão.
- **Segredos**: só as credenciais do Google, por ambiente, com placeholders em `.env.example`.

## Custo

A Fase 0 não usa nenhuma API externa. Em produção (Fase 5) a estimativa é um VPS pequeno (Hetzner, cerca de €4 por mês), armazenamento S3-compatível para o Litestream (Cloudflare R2, faixa gratuita) e e-mail transacional para links mágicos (Resend, faixa gratuita). Total esperado: abaixo de €5 por mês.

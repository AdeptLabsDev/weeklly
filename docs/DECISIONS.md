# Decisões

Registro de decisões técnicas e de produto. Cada entrada tem escolha, razão e data. Uma decisão só muda com uma nova entrada que a substitua explicitamente.

## Stack decidida (2026-09-05)

| Camada | Escolha |
|---|---|
| Linguagem do servidor | **Go** |
| HTTP, rotas, templates | Biblioteca padrão do Go (`net/http`, `html/template`) |
| CSS | Tailwind v4, CLI standalone |
| JS | Nativo, módulos ES, sem framework e sem bibliotecas de animação |
| Banco | SQLite, modo WAL, backup contínuo com Litestream |
| Tema | Dark-first, tema claro na Fase 4 |
| Usuários | Contas multiusuário desde o início |

Detalhes e razões em D1 a D7.

## D1 — Tema dark-first (2026-09-05)

**Escolha.** Dark-first. Tema claro entra na Fase 4.

**Razão.** Padrão Higher Mind. O Miro entra como gramática visual (grid de pontos, cartões, toolbar flutuante), não como paleta.

## D2 — Contas multiusuário (2026-09-05)

**Escolha.** Contas desde o início. Toda semana tem um dono.

**Razão.** O site é público, o app virá e o produto pode expandir. Dono desde o início é uma coluna; adicionar depois é migração de dados.

## D3 — Tailwind v4 (2026-09-05)

**Escolha.** Tailwind v4 via CLI standalone (binário único, sem Node no projeto). O CSS gerado é embutido no binário do servidor com hash no nome do arquivo.

**Razão.** Exigência do projeto. O CLI standalone elimina Node e npm do repositório, o que reduz superfície de ataque na cadeia de dependências e simplifica o build.

## D4 — Frontend sem framework e sem bibliotecas de animação (2026-09-05)

**Escolha.** HTML renderizado no servidor. JavaScript nativo em módulos ES, sem bundler, sem dependências. Sem React ou equivalente. Animação com CSS e View Transitions API.

**Razão.** Exigência do projeto, e o produto justifica: a interatividade é pequena (autosave, troca de semana, tema, hub). O navegador moderno cobre tudo nativamente:

- `fetch` com `keepalive` para salvar ao fechar a aba
- `localStorage` e IndexedDB para o rascunho local que fica à frente do servidor
- View Transitions API para a transição entre semanas
- Speculation Rules para pré-renderizar as semanas vizinhas
- Service Worker para PWA e leitura/escrita offline

**Consequência.** O app mobile pode ser o mesmo código web empacotado (Trusted Web Activity no Android, WebView no iOS). Decisão formal na Fase 6.

## D5 — SQLite (2026-09-05)

**Escolha.** SQLite, modo WAL, um arquivo, um processo.

**Razão.** O domínio é pequeno (semanas por usuário), o servidor é um só, e leitura local em SQLite é mais rápida que qualquer banco em rede. Não há latência de conexão, pool ou serviço separado para operar.

**Configuração obrigatória.** `journal_mode=WAL`, `foreign_keys=ON`, `busy_timeout=5000`, `synchronous=NORMAL`. Um único pool de escrita.

**Backup.** Litestream replicando o WAL continuamente para armazenamento S3-compatível (Cloudflare R2 ou Backblaze B2). Restauração testada antes do lançamento.

**Limite conhecido.** Um servidor, uma região. Se um dia precisar de várias regiões, o caminho é LiteFS ou Postgres. Não é o caso deste produto.

## D6 — Linguagem do servidor: Go (2026-09-05)

**Escolha.** Go. Decidido pelo Miguel em 2026-09-05 após a avaliação abaixo.

**Critérios.**

1. Nova para o Miguel (já usou Next.js e Node.js).
2. Comunidade ativa e documentação profunda, para que agentes consultem bem e errem pouco.
3. Ótima com SQLite e com HTML renderizado no servidor.
4. Deploy simples, alinhado a um processo e um arquivo de banco.
5. Estabilidade de longo prazo: código de hoje precisa compilar daqui a dez anos.
6. Ciclo rápido de iteração com agente: compilar, testar, corrigir.

**Opções avaliadas.**

| Linguagem | Por que faria sentido | Custo | Agentes |
|---|---|---|---|
| **Go** | Biblioteca padrão cobre HTTP, templates com escape automático, CSRF, crypto, logs e embed. Um binário estático. Compila em segundos. Promessa de compatibilidade do Go 1. | Menos "baterias": auth e sessões usam bibliotecas pequenas ou código próprio. Mais verboso. | Excelente. Documentação canônica única (pkg.go.dev), um só estilo (gofmt), poucas versões de API para confundir. |
| **Rust (Axum)** | Performance máxima, segurança de memória, um binário. Ecossistema web maduro. | Compilação lenta e muito atrito por iteração para um app deste tamanho. Overkill. | Muito bom, mas com mais rodadas de erro de compilação por tarefa. |
| **Ruby (Rails 8)** | SQLite é o padrão de produção do Rails 8. Auth, migrações, jobs e cache prontos. Tailwind integrado. | Runtime pesado no deploy. Hotwire traz JS importado (Turbo, Stimulus), o que conflita com D4. Muita deriva de versão. | Bom, mas agentes misturam APIs de Rails 6, 7 e 8 com frequência. |
| **Elixir (Phoenix)** | LiveView elimina JS próprio. PubSub daria sync em tempo real quase de graça. | Depende de WebSocket e do cliente JS do LiveView. Conflita com offline e PWA do roadmap. SQLite é cidadão de segunda classe no Ecto. | Bom, comunidade menor, mais erros sutis. |
| **C# (.NET)** | Razor Pages, EF Core com SQLite, Identity pronto. Comunidade enorme. | Tooling pesado, imagens grandes, menos alinhado a "um binário e um arquivo". | Muito bom. |

**Razão.**

1. A biblioteca padrão resolve o servidor inteiro: roteamento por método e caminho em `net/http`, `html/template` com escape contextual (XSS coberto por padrão), `http.CrossOriginProtection` para CSRF (Go 1.25+), `crypto`, `log/slog`, `embed`. Menos dependências é menos superfície de ataque e menos versões para o agente confundir.
2. Um binário estático. Deploy é copiar um arquivo. Cross-compile do Windows para Linux com uma variável de ambiente. Combina com SQLite: um processo, uma máquina, um arquivo.
3. Compila em segundos e o compilador verifica erros não tratados. O ciclo agente, compilador, teste é rápido e com feedback claro.
4. Promessa de compatibilidade do Go 1. Código escrito hoje compila em dez anos. Combina com pensar em décadas.
5. Documentação canônica em um só lugar e um único estilo de código. Agentes produzem Go idiomático com consistência alta.
6. Dezenas de MB de RAM e latência de handler abaixo de um milissegundo com SQLite local. O orçamento de 100 ms fica inteiro para a rede.

**Alternativas descartadas.** Rust: atrito de iteração alto para o tamanho do app. Rails 8: exige Hotwire, o que conflita com D4. Phoenix: conflita com offline e PWA do roadmap. .NET: tooling pesado para um binário e um arquivo.

## D7 — Stack completa (2026-09-05)

**Escolha.** A stack abaixo, fechada junto com D6. Cada linha pode ser substituída por uma nova decisão com razão escrita.

| Camada | Escolha | Razão |
|---|---|---|
| Linguagem | Go, versão estável mais recente, toolchain fixada no `go.mod` | Ver D6 |
| HTTP e rotas | `net/http` da biblioteca padrão, padrões de método e caminho | Sem router de terceiros; suficiente e estável |
| Templates | `html/template`, layouts e parciais | Escape contextual automático; zero dependência |
| CSS | Tailwind v4 CLI standalone; saída embutida com `embed` e hash no nome | D3 |
| JS | Módulos ES nativos, sem bundler, embutidos no binário | D4 |
| Banco | SQLite via `modernc.org/sqlite` (Go puro, sem cgo) com `database/sql` | Sem toolchain C no Windows; cross-compile trivial. Desempenho suficiente para este domínio |
| Migrações | SQL puro em arquivos embutidos, aplicadas no boot por migrador próprio (`internal/store`, cerca de 60 linhas) | Sequencial e só para frente. O goose traria uma árvore de dependências para um trabalho de 60 linhas |
| Sessões | Tabela própria em SQLite, cookie com token aleatório e só o hash no banco (D15, substitui o `scs` previsto) | 80 linhas com STRICT e chave estrangeira valem mais que uma dependência com blobs gob |
| Autenticação | Login com o Google por OpenID Connect na biblioteca padrão (D14, substitui o link mágico previsto); outras formas depois | Decisão do Miguel; sem senha para vazar |
| CSRF | `http.CrossOriginProtection` | Biblioteca padrão |
| Logs | `log/slog`, JSON em produção | Biblioteca padrão |
| Testes | `testing` e `httptest` para handlers e serviços; Playwright para ponta a ponta | Playwright é dependência de desenvolvimento, fora do produto |
| Lint | `gofmt`, `go vet`, `staticcheck`, `golangci-lint` | Bloqueiam CI |
| Dev local | `air` para live reload; Taskfile para comandos | Funciona nativamente no Windows |
| Container | Multi-stage, imagem final `scratch` ou distroless, non-root, binário estático | Segurança day-1 do padrão Higher Mind |
| Backup | Litestream para Cloudflare R2 ou Backblaze B2 | D5 |
| Hospedagem | Hetzner VPS com Caddy (HTTPS automático, cabeçalhos de segurança) e Docker Compose; alternativa Fly.io com volume | Um servidor, custo baixo, controle total. Fechar na Fase 5 |
| API para o app | `/api/v1` em JSON, mesma camada de serviço dos handlers HTML, versionada | Fase 6 |

## D8 — Direção de design: tokens e tipografia (2026-09-05)

**Escolha.** Gramática do Miro em paleta noturna: quadro com grid de pontos, cartões elevados com sombra profunda, toolbar flutuante em pílula. Tipografia editorial: Instrument Serif itálico no título da semana e nos números dos dias, Instrument Sans na interface e no texto. Uma única cor de destaque, âmbar, reservada para "hoje" e para o foco. Tokens em `web/styles/app.css` (`@theme`), em OKLCH.

**Razão.** O Miro é referência de vocabulário, não de paleta (D1). O âmbar único evita a estética de dashboard multicolorido e faz do "hoje" o único ponto de luz do quadro. A serif itálica dá voz própria ao título sem custo de legibilidade, e os números em serif ligam os cartões ao título. Fontes servidas do próprio domínio, subconjunto latin, para respeitar a CSP e não depender de terceiros.

## D9 — Semana começa na segunda-feira (2026-09-05)

**Escolha.** Segunda a domingo. A posição 0 é segunda e a 6 é domingo, com CHECK no banco e tipo próprio no código (`week.Weekday`). Com D12 a semana deixou de ter datas; a ordem dos dias continua esta.

**Razão.** Convenção brasileira e padrão ISO, alinhada à matemática de semanas da biblioteca padrão. Se um dia virar configurável, muda a apresentação, não a identidade dos dados.

## D10 — Layout do quadro no desktop: 5 + 2 (2026-09-05)

**Escolha.** Seis colunas: segunda a sexta ocupam a altura inteira; sábado e domingo dividem a sexta coluna. Abaixo de 1024px, duas colunas; abaixo de 640px, uma. No desktop o quadro ocupa exatamente a viewport e o conteúdo rola dentro do cartão.

**Razão.** Sete colunas iguais ficam estreitas em notebook. O fim de semana costuma ter menos texto e a hierarquia comunica isso. Validado na tela da Fase 0.

## D11 — Quadro como faixa horizontal de sete cartões iguais (2026-09-05)

**Escolha.** Substitui D10. Os sete dias têm o mesmo tamanho, quase quadrados e grandes: o menor entre 640px, 84% da largura da tela e a altura disponível. Ficam numa faixa horizontal que o usuário desliza para o lado (roda do mouse, trackpad, toque, arrastar ou setas), com encaixe por cartão. Não há moldura em volta do quadro: o grid de pontos cobre a página inteira, os cartões flutuam e as bordas laterais esmaecem para sugerir continuidade. Ao abrir, o cartão de hoje vem para o centro. Vale igual no desktop e no mobile.

**Razão.** Feedback do Miguel na primeira tela: sábado e domingo menores criavam uma hierarquia que o produto não tem, e a moldura fechada contradizia a sensação de espaço aberto que é a referência do Miro. Sete cartões iguais numa faixa tratam cada dia como o mesmo espaço de planejamento. O quadro continua finito (sete cartões, sem zoom, sem rolagem vertical): invariante 6.

**Consequência.** Primeiro JavaScript do projeto, `web/static/app.js`, nativo, sem dependências (D4): centralizar hoje, roda vertical vira horizontal, arrastar com o mouse com inércia, setas do teclado, paginador. Sem `scroll-snap`: a primeira versão usava encaixe e ele anulava a roda do mouse e puxava o arrasto de volta ao soltar; a faixa agora para exatamente onde a mão parou, e só desliza com inércia se soltar em movimento. Na Fase 1 o arrastar precisa ignorar o editor de texto.

## D12 — Semana é um espaço com nome, sem datas (2026-09-05)

**Escolha.** Uma semana é um espaço de planejamento nomeado pelo usuário ("Semana padrão", "Semana de provas"), com sete dias de segunda a domingo e sem datas. Identificada por um id opaco de 16 caracteres gerado pelo sistema. O hub lista as semanas pela última edição. "Hoje" é só o dia da semana atual no fuso do usuário. Substitui a semana ancorada em datas do calendário (D9 continua valendo para a ordem dos dias).

**Razão.** A dor do produto, na palavra do Miguel: o planejamento semanal costuma ser o mesmo por várias semanas, e calendários obrigam a repetir cada tarefa em cada data. Uma semana que se monta uma vez e se consulta sempre resolve isso na raiz. Datas podem voltar como recurso opcional se fizerem falta.

**Consequência.** O esquema `0001_schema.sql` foi reescrito em vez de receber uma migração nova: nenhum banco fora de desenvolvimento existia, e uma migração de "datas para espaços" seria história falsa. O pacote `internal/week` perdeu toda a matemática de calendário; ficou o dia da semana por fuso, a validação do nome e o gerador de id. O banco cria os sete dias no gatilho de inserção da semana e impede apagar um dia avulso.

## D13 — Revisão visual guiada por skills (2026-09-05)

**Escolha.** Duas skills instaladas globalmente com autorização do Miguel: `anthropics/skills@frontend-design` (direção estética, contra padrões de página gerada) e `vercel-labs/agent-skills@web-design-guidelines` (regras de acessibilidade, foco, animação, toque, conteúdo). Aplicadas na segunda versão do quadro:

- Paleta: quadro noturno frio (azul-grafite) com cartões em papel quente, como notas sob uma luminária. Sai o "quase-preto com um acento", entra um contraste de temperatura deliberado. Âmbar segue como única cor de destaque, só para hoje.
- Tipografia: o nome do dia é o herói de cada cartão, em serifa itálica e caixa normal. Saem os rótulos em caixa alta espaçada e os pontos médios entre metadados.
- Copy em frases: "Hoje é sábado.", "Ir para hoje", "Sem plano ainda". Sem rótulos acima de conteúdo.
- Usabilidade: paginador com os sete dias abreviados que acende os visíveis e leva ao clicado; setas, Home e End no teclado; "Ir para hoje" como âncora que funciona sem JavaScript; `touch-action`, área segura, foco visível, movimento reduzido, botões com `aria-label`, links para navegação e botões para ações.

**Razão.** A primeira faixa tinha os sinais que a skill da Anthropic lista como marca de página gerada, e o Miguel pediu mais qualidade visual e de uso. As duas skills têm origem oficial e centenas de milhares de instalações; ficam disponíveis para as próximas telas.

## D14 — Login com o Google e sessão anônima antes do login (2026-09-06)

**Escolha.** A primeira forma de entrar é a conta Google, por OpenID Connect (authorization code com PKCE), implementada em `internal/auth` só com a biblioteca padrão: o `id_token` é verificado contra as chaves públicas do Google (assinatura RS256, emissor, audiência, validade, nonce) e só e-mails confirmados entram. Outras formas de entrar vêm depois.

Antes de entrar, o visitante já usa o produto: ao criar a primeira semana ele ganha um usuário anônimo e uma sessão neste navegador. Entrar com o Google reivindica esse usuário (ele ganha e-mail, nome e foto) ou, se a conta Google já existe, move as semanas anônimas para ela. Nada se perde.

**Razão.** Decisão do Miguel pelo Google como primeiro provedor. A sessão anônima existe porque a barra de navegação precisa listar e criar semanas de verdade, e porque um planejador que exige conta antes de mostrar valor perde a pessoa na porta. Sem as credenciais (`WEEKLLY_GOOGLE_CLIENT_ID` e `WEEKLLY_GOOGLE_CLIENT_SECRET`), o botão leva a uma página que explica que o login está desligado neste ambiente: nunca um botão que não faz nada.

**Consequência.** `users.email` passou a aceitar nulo e ganhou `google_sub`, `name` e `picture_url`. A CSP libera imagens de `lh3.googleusercontent.com` para a foto de perfil. O fluxo inteiro é testado contra um provedor falso (`auth.FakeProvider`) que assina tokens com chave própria. Em produção o login exige `WEEKLLY_BASE_URL` em https.

## D15 — Sessões próprias em SQLite (2026-09-06)

**Escolha.** Substitui o `alexedwards/scs` previsto em D7. Tabela `sessions` STRICT com chave estrangeira para o usuário; o cookie leva um token de 32 bytes aleatórios e o banco guarda só o SHA-256 dele. HttpOnly, SameSite=Lax, Secure fora de development. Validade de 90 dias, renovada no máximo uma vez por hora de uso. A sessão lembra a última semana aberta, para o app abrir onde a pessoa parou.

**Razão.** O que o produto precisa de uma sessão cabe em 80 linhas, e assim as invariantes ficam no banco como o resto (cascade ao apagar o usuário, `last_week_id` que vira nulo quando a semana some). O `scs` guardaria blobs gob opacos e traria uma dependência para isso.

## D16 — Playwright para testes ponta a ponta, em `e2e/` (2026-09-06)

**Escolha.** Adiantado da Fase 5: `e2e/` tem o Playwright com Chromium, sobe o servidor sozinho num banco temporário e percorre os fluxos reais (criar semanas, trocar pelo seletor, hub, login desligado, roda e arrasto da faixa) em desktop e num Pixel 7 emulado. Também tira as capturas de tela usadas nas revisões de design.

**Razão.** O Edge headless não envia formulários nem guarda sessão, e o quadro só existe para quem tem sessão. Sem um navegador de verdade não dá para verificar o produto de ponta a ponta nem fotografá-lo. O Node fica confinado a `e2e/`, fora do produto e da imagem Docker (D3, D4 continuam valendo).

## D17 — Monocromático em dois temas, Nunito só no que tem nome (2026-09-06)

**Escolha.** Enquanto a cor de destaque não é definida, o sistema é monocromático no vocabulário da Vercel: no escuro, fundo quase preto (`#0a0a0a`), superfícies em cinza médio (`#171717`, `#222`), bordas discretas, texto claro; no claro, fundo `#fafafa`, superfícies brancas, texto quase preto. "Hoje" é marcado por contraste: borda mais forte e etiqueta invertida. Vermelho fica só em ações destrutivas. O token `--color-accent` existe e hoje é a cor do texto: a cor de destaque entra em uma linha.

O tema é escolhido no botão sol/lua da barra (o ícone mostra o tema atual) e vai num cookie legível pelo script: o servidor entrega a página já no tema certo, sem piscar, e o botão troca sem recarregar. Padrão: escuro.

Tipografia: Nunito (variável, subconjunto latin, servida do próprio domínio) em peso 800 só nos papéis de exibição: nome da semana, nomes dos dias, títulos de página e de diálogo, itens do seletor e do hub. Todo o resto usa a fonte do sistema da pessoa. Substitui D8 (Instrument Serif e Sans) e D13 no que diz respeito a cores e fontes.

**Razão.** Decisão do Miguel: cor de destaque fica para depois, fonte da marca é a Nunito, corpo em fonte do sistema. Monocromático evita gastar o acento antes da hora e deixa a hierarquia se sustentar só em contraste e peso.

## D18 — O dia é uma lista de tarefas (2026-09-06)

**Escolha.** Substitui o texto livre por dia. Cada dia tem tarefas na ordem em que a pessoa as põe, como colunas de um quadro. Uma tarefa tem título (1 a 200 caracteres), horário opcional (`HH:MM`) e o estado "feita". Adicionar, editar título e horário no lugar, concluir e excluir acontecem sem recarregar; sem JavaScript, os mesmos formulários funcionam e voltam para o dia. O servidor responde a tarefa já renderizada (JSON com o HTML da parcial), então o markup tem uma só fonte.

**Razão.** Pedido do Miguel: organizar por tarefas facilita mais que um bloco de texto, e o horário por tarefa é o gancho para o dia ganhar forma. A ordem é manual porque é um plano, não uma agenda; ordenar por horário pode entrar como opção. Limite de 100 tarefas por dia.

**Consequência.** A tabela `days` saiu; entrou `tasks` com `weekday`, `position`, `title`, `time`, `done`, e gatilhos que carimbam a tarefa e marcam a semana como recente. Duplicar semana copia as tarefas sem o "feita". Arrastar a faixa ignora campos, botões e formulários.

## D19 — Ações da semana e ordem da lista (2026-09-06)

**Escolha.** No seletor da barra, sob "Esta semana": renomear (diálogo, com página de fallback), duplicar (cria "Nome (cópia)" com as tarefas e abre) e excluir (diálogo de confirmação, foco em "Cancelar"; página de fallback). A lista de semanas, no seletor e no hub, alterna entre "Recentes" e "A–Z"; a escolha fica no usuário, então vale em qualquer dispositivo depois de entrar.

**Razão.** Duplicar é o gesto central de um produto sobre semanas reutilizáveis. Renomear e excluir são o mínimo para a lista não virar lixo. A ordem alfabética foi pedida pelo Miguel; "recentes" continua sendo o padrão porque é a que responde "onde eu estava".

## D20 — Movimento com propósito e mecânica das tarefas (2026-09-06)

**Escolha.**
- A troca de tema usa a View Transitions API: o tema novo varre a página da esquerda para a direita (560 ms) e o botão sai girando e encolhendo enquanto o ícone novo entra girando ao contrário. Sem a API, ou com "reduzir movimento", a troca é imediata. Só CSS e o `startViewTransition` nativo.
- Reordenar tarefas é arrastar pela alça (seis pontos, visível ao passar o mouse; sempre visível no toque). O fantasma segue o ponteiro, a tarefa mostra o lugar ao vivo e soltar salva; soltar em outro cartão move a tarefa de dia. Alt com seta para cima ou para baixo faz o mesmo pelo teclado.
- Clicar em qualquer espaço vazio de um dia foca o campo de nova tarefa daquele dia.
- O horário é um campo de texto que aceita "7", "930", "9:30" e "21h", com um seletor leve abaixo (meia em meia hora e "Sem horário"). Substitui o `<input type="time">` nativo, ruim de usar e feio no escuro.
- "Recentes" e "A–Z" trocam sem recarregar: a pílula desliza e as listas do seletor e do hub se reordenam com animação de posição (FLIP), salvando a preferência em segundo plano.
- A logo leva ao hub; `/` continua abrindo a última semana.

**Razão.** Pedidos do Miguel nesta rodada. A View Transitions API dá a varredura sem biblioteca e sem tocar no DOM; a alça de arrasto evita conflito com a edição no lugar (o título é um botão) e com a rolagem por toque; o campo de horário em texto é mais rápido de digitar que qualquer seletor.

**Consequência.** Nova rota `POST /tarefas/{id}/mover` (`weekday`, `position`), com `MoveTask` no banco fechando e abrindo espaço nas posições. Com captura de ponteiro na faixa, o clique chega à própria faixa: o toque em área vazia é tratado no `pointerup`, guardando o alvo do `pointerdown`.

## Pendentes para fases futuras

- **Outras formas de entrar** depois do Google (Fase 3): passkeys via `go-webauthn` é a candidata.
- **Hospedagem final** (Fase 5). Recomendação em D7.
- **Monitoramento de erros e uptime** (Fase 5). Recomendação: Sentry e Better Stack.
- **Empacotamento do app** (Fase 6). Ver consequência em D4.

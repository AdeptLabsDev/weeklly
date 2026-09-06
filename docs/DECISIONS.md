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

## D21 — Dois idiomas, configurações rápidas e histórico de desfazer (2026-09-06)

**Escolha.**
- Todo texto da interface vem de `internal/i18n`: catálogos em português do Brasil e inglês como mapas Go, com um teste que exige as mesmas chaves e o mesmo número de argumentos nos dois. Isso inclui nomes dos dias e meses, rótulos de data ("editada há 3 dias"), erros de domínio (o pacote `week` passou a ter erros como identificadores, sem texto de interface), páginas de login e as frases do script, que chegam ao navegador num atributo `data-i18n` do body.
- O idioma é escolhido no menu de configurações (botão entre a casinha e "Nova semana", um popover sem página própria), guardado em cookie. Na primeira visita, o `Accept-Language` do navegador decide; o padrão é português. A troca recarrega a página, porque tudo muda.
- Desfazer e refazer ficam na linha de "Hoje é...", desktop e celular, com Ctrl+Z, Ctrl+Shift+Z e Ctrl+Y. O histórico é por aba, em memória: cada ação de tarefa (adicionar, editar, concluir, mover, excluir) guarda sua inversa. Desfazer uma exclusão recria a tarefa com o mesmo texto, horário, estado e posição; como o id muda, o registro se atualiza para o refazer.

**Razão.** Pedidos do Miguel. Catálogo em Go, sem biblioteca, mantém a stack (D3, D4) e faz o compilador e o teste vigiarem as traduções. Histórico no cliente é o padrão de editores (por aba, some ao recarregar) e não exige tabela de eventos no servidor; se um dia houver várias abas ou dispositivos editando junto, o caminho é o histórico no servidor.

**Consequência.** Controle segmentado genérico (`.segmented`) usado pela ordem das semanas e pelo idioma. `POST /idioma`. `internal/week` sem nomes de dias.

## D22 — Cursor do produto opcional, painel de idioma e cor à escolha (2026-09-06)

**Escolha.**
- O site tem um cursor próprio, desligado por padrão: uma seta de contorno com cantos arredondados que, enquanto o botão principal está pressionado, encolhe de leve e ganha dois arcos na ponta (referência visual do Miguel). Liga em "Tipo de mouse" nas configurações, um controle de dois ícones (a seta do sistema e a do weeklly), guardado em cookie. É um elemento da página (`partials/cursor.html`), não uma imagem em `cursor: url()`: com a opção ligada, script e ponteiro fino, o `<html>` recebe `has-cursor`, o cursor do sistema some e o elemento segue o ponteiro por `transform`. Ele é um popover manual, então vive no top layer e o script o promove de novo sempre que um diálogo ou menu abre. Sobre campos de texto, áreas de escrita e barras de rolagem ele some e o cursor do sistema volta. A posição atravessa a navegação pela `sessionStorage`. No toque e sem script, valem os cursores do sistema.
- A cor de destaque é da pessoa: "Cor" nas configurações mostra oito amostras (sem cor, laranja, âmbar, verde, azul, violeta, rosa, vermelho), em cookie, aplicadas pelo servidor como `data-accent` no `<html>`. Sem cor, o produto segue monocromático (D17). Com uma cor, ela vai para `--color-accent` (etiqueta "hoje", anel do dia, tarefa feita, botão primário, foco), para os textos principais (`--color-display`: nome da semana, nomes dos dias, títulos de página e de diálogo, itens do hub) e para o cursor (`--color-cursor`). Cada cor tem um tom por tema, escolhido para ler bem como texto sobre o fundo. A troca entra com a mesma varredura do tema, sem girar o ícone.
- O idioma nas configurações é uma linha "Idioma · Português ⌄" que abre um painel próprio, sem esticar o menu: um popover aninhado no de configurações, ao lado da linha no desktop e abaixo do menu quando não cabe. A lista vem de `Locale.Languages()`: um idioma novo entra no catálogo e aparece sozinho.

**Razão.** A primeira versão do cursor usava `cursor: url()` e piscava ao clicar em links e formulários; a causa está no Chromium: enquanto uma navegação carrega, ele troca cursor de imagem pelo padrão (`is_loading_` em `RenderWidgetHostViewAura::UpdateCursorIfOverSelf`) e só recoloca o da página nova quando o mouse se mexe. O cursor como elemento resolve o caso da imagem, mas o Miguel ainda viu o cursor do sistema em navegações, na troca de tema (View Transitions) e no duplo clique (seleção de texto): o navegador nunca entrega um cursor 100% próprio. Daí a decisão dele: o do sistema é o padrão, o nosso é uma escolha. A casca do app (barra, cabeçalhos dos dias, menus) ficou sem seleção de texto, o que tira o caso do duplo clique. A cor à escolha também é decisão do Miguel: em vez de uma cor de destaque única, cada pessoa tem a sua, e o cursor a acompanha.

**Consequência.** Rotas `POST /cursor` e `POST /cor`, cookies `weeklly_cursor` e `weeklly_accent`, atributos `data-cursor` e `data-accent` no `<html>`. Paleta em `@theme` (`--accent-*`, um tom por tema) e mapeamento em `@layer base`. `.segmented` fica para a ordem das semanas e para o tipo de mouse. O elemento do cursor usa `data-cursor-el`, porque `data-cursor` no `<html>` é a preferência. Testes: cookies e fallback dos valores, e no navegador ligar o cursor, escolher a cor, persistir depois de recarregar e o comportamento do cursor ligado.

## D23 — Páginas públicas e fundação de SEO (2026-09-06)

**Escolha.** A apresentação do produto fica em `/planejador-semanal` (pt-BR) e `/en/weekly-planner` (en), com idioma estável pela URL, conteúdo completo no HTML e links entre traduções. `/` preserva a consulta instantânea à última semana (D20). O hub vazio tem um link para conhecer o produto. Links de entrada guardam o idioma escolhido como preferência do app.

Canonical, hreflang recíproco, metadados Open Graph/Twitter, imagens PNG locais e JSON-LD `WebPage` + `WebApplication` descrevem apenas conteúdo público real. O JSON é serializado por `encoding/json` e autorizado por hash específico na CSP. Não são declarados preços, reviews, avaliações nem recursos futuros. As páginas públicas ignoram sessões e não dependem de JavaScript.

**Revisão do Miguel (mesmo dia).** A landing foi refeita em cima dessa fundação:
- A barra tem o botão sol/lua e o seletor PT/EN, nas mesmas posições do app. O tema segue o cookie (é apresentação, não dado pessoal) e troca por formulário sem script, ou com a varredura quando o `app.js` está carregado; o idioma continua fixo pela URL, e o seletor são links entre as duas traduções. A página passou a carregar o `app.js` do produto, só como melhoria: todo o conteúdo já vem do servidor.
- A headline diz o que o produto é e seus pontos fortes concretos, sem frase abstrata: "Planejador semanal online, grátis e sem cadastro." Grátis e sem cadastro são decisão de produto do Miguel.
- Saíram a demonstração com tarefas de exemplo (refazia o produto e passava outra impressão) e a seção de rotinas; entraram "Como funciona" em três passos e "O que vem com a sua semana", seis pontos concretos: sete dias iguais, grátis e sem cadastro, sem datas, horário opcional, arrastar/mover/desfazer, duplicar.
- O FAQ é a última seção antes do rodapé, com oito perguntas (grátis, conta, datas, salvamento, reutilizar, horário, celular, idiomas). Sem bloco de encerramento depois dele.
- O rodapé lista Quem somos, Produtos, Aplicativos, Ajuda e suporte, Termos e privacidade e Redes sociais. O que ainda não tem página aparece como texto com a marca "em breve", nunca como link morto; as páginas entram no backlog.

**Razão.** Aquisição precisa de conteúdo rastreável sem comprometer a volta imediata ao quadro nem expor planos pessoais. URLs por idioma evitam que cookies e negociação de idioma escondam traduções. A infraestrutura usa só a biblioteca padrão do Go e a stack atual; nenhuma dependência nova de produção.

**Consequência.** `WEEKLLY_SEARCH_INDEXING=false` por padrão; `true` exige produção e `WEEKLLY_BASE_URL` HTTPS. A origem configurada alimenta todas as URLs absolutas. O sitemap só inclui as duas páginas públicas; com indexação desligada fica vazio. Em produção habilitada, robots permite o rastreamento para que buscadores leiam `noindex` nas rotas do app; o controle de acesso continua independente. Roteiro de lançamento, skills pesquisadas, testes e métricas estão em `docs/SEO.md`.

## D24 — Mascote: o esquilo, e o ícone do site (2026-09-06)

**Escolha.** A weeklly tem uma mascote, um esquilo, desenhado pelo Miguel em silhueta (`design/brand/squirrel-icon-menu.svg`). A primeira aplicação é o ícone do site: o esquilo na cor principal da weeklly, `#727cf5`, sobre fundo transparente, em `web/static/favicon.svg` (a partir dos dois caminhos preenchidos do original, centralizados numa caixa quadrada), com `favicon-32.png` para navegadores sem favicon SVG (Safari) e `apple-touch-icon.png` de 180 px sobre ladrilho `#0a0a0a` para a tela inicial do iPhone, que não aceita transparência. Os PNGs são renderizados do SVG pelo Chromium do Playwright; refazer com `node <scratch>/icons.cjs` ou qualquer conversor, sempre a partir do SVG.

**Razão.** Decisão de marca do Miguel, na linha da Duolingo (ver `docs/LANDING-REFERENCE.md`): a mascote dá rosto ao produto e aparece primeiro onde a pessoa mais olha, a aba do navegador. A cor principal é decisão do Miguel; ela vive em `app.css` como `--color-brand`, um token só da marca (ícone, landing), e não muda o app: a interface segue monocromática, com o destaque escolhido por cada pessoa (D22).

**Consequência.** O favicon de pontos (D8) sai. `layout.html` declara os três ícones com hash de conteúdo. Próximos usos previstos: abertura da landing e assinatura no rodapé.

**Revisão (2026-09-06).** O ícone nasceu em coral `#FF6766`; no mesmo dia o Miguel definiu `#727cf5` como a cor principal da weeklly. O SVG e os dois PNGs foram refeitos nessa cor e o token `--color-brand` entrou em `app.css`.

## D25 — Concluir com movimento, duplicar tarefa e idioma da landing (2026-09-06)

**Escolha.**
- Concluir uma tarefa anima: a caixa pulsa, o traço do check se desenha e uma linha atravessa o título da esquerda para a direita. Desmarcar recolhe a linha. O risco deixou de ser `text-decoration` e virou um gradiente de fundo no texto (`background-size` animado, `box-decoration-break: clone`), porque decoração de texto não anima e o fundo por fragmento de linha acompanha títulos que quebram. O script marca `is-just-done`/`is-just-undone` quando a versão nova da tarefa chega do servidor com o "feita" diferente; sem script, ou com "reduzir movimento", vale o estado final, idêntico.
- Duplicar tarefa: botão só com ícone, à direita do x, aparece com a tarefa como os demais. `POST /tarefas/{id}/duplicar` cria a cópia logo abaixo, no mesmo dia, com título e horário e sem o "feita" (uma cópia é um a fazer novo), respeitando o limite do dia. Entra no histórico: desfazer apaga a cópia, refazer a recria.
- Na landing, o idioma é um botão com o nome do idioma atual que abre um painel com todos os idiomas do sistema, o mesmo painel das configurações do app, com links para a tradução (o idioma da página pública é fixo pela URL, D23). A lista vem de `Locale.Languages()` e de `landingPaths` em `seo.go`: um idioma novo entra nos dois e aparece sozinho.

**Razão.** Pedidos do Miguel. A animação dá o retorno do gesto mais frequente do produto sem esconder o estado final. Duplicar é o mesmo gesto de reutilizar que já existe nas semanas, agora na tarefa. O painel único de idiomas escala para vários idiomas sem pesar a barra.

**Consequência.** `DuplicateTask` no store, rota nova, chaves `task.duplicate`, `task.duplicateTitle` e `js.duplicated`. O posicionamento do painel de idioma passou a partir do invocador: ao lado da linha nas configurações, abaixo do botão na landing. Testes: store, handler e navegador (animação, duplicar com desfazer, painel na landing sem script).

## D26 — Landing definitiva: cena de "como funciona" e página de perguntas (2026-09-06)

**Escolha.**
- A barra da landing fica com logo, tema, idioma e "Abrir planejador". O atalho "Como funciona" saiu: a página é curta e o botão competia com o único caminho que importa.
- "Como funciona" deixou de ser três textos numa linha e virou um elemento só: os três passos à esquerda, ligados por um trilho, e uma cena à direita que muda com o passo. A cena é um SVG com a forma do quadro (o nome sendo digitado, sete colunas com os dias no idioma da página, tarefas como barras sem texto) e não uma cópia do produto: barras, e não tarefas de exemplo, por causa da decisão anterior do Miguel contra a demonstração. Os passos são um grupo de rádios; a cena e o texto aberto reagem ao rádio marcado só com CSS (`:has`), então tudo funciona sem script, com teclado (setas) e leitor de tela. O script acrescenta o avanço automático: com a seção na tela, o trilho do passo atual enche em 5,5 s e, no `animationend`, o próximo é marcado; termina no passo 3, a semana pronta. Passar o mouse nos passos ou focar um pausa; escolher um passo encerra o avanço. Com "reduzir movimento", nada anda sozinho.
- A cor da marca (`--color-brand`, D24) aparece só no que está vivo na cena e no trilho: o cursor de texto, o horário, a tarefa feita, a tarefa que muda de dia, o passo atual. O resto segue monocromático.
- As perguntas frequentes saíram da landing e ganharam página própria, `/perguntas-frequentes` e `/en/faq`, ligada em "Ajuda e suporte" no rodapé. A página tem H1, as oito perguntas em `<details>` do mesmo grupo (`name="faq"`: abrir uma fecha a outra, sem script) e um fechamento com a chamada para criar a semana. O JSON-LD dela é `WebPage` + `FAQPage` com as perguntas e respostas do catálogo, a mesma lista que renderiza a página.
- `seo.go` passou a descrever as páginas públicas como uma lista (`publicPages`: template e URL por idioma). Rotas, `publicPath`, seletor de idiomas, hreflang e sitemap saem dela; barra e rodapé viraram a partial `public.html`, compartilhada. Uma página pública nova é uma entrada na lista, um template e as chaves de texto.

**Razão.** Pedidos do Miguel para a versão definitiva da landing: barra sem o atalho, "como funciona" como elemento visual e interativo, perguntas numa página dedicada. A cena mostra a forma do produto sem fingir ser ele; rádios e `:has` mantêm o contrato da página pública (conteúdo completo sem script) e ganham interação de graça. A lista de páginas evita repetir metadados a cada página nova.

**Consequência.** Chaves `landing.nav.how` removida, `landing.faq.*` renomeadas para `faq.*`, novas `faq.title`, `faq.description`, `faq.more.*`, `landing.steps.legend`. `seoView` ganhou `Landing`, `FAQ`, `Days` (dias da cena, com posição) e `Questions`. O sitemap lista quatro URLs. Testes: estrutura das duas páginas nos dois idiomas, FAQPage no grafo, e no navegador os passos sem script, o avanço automático parando ao clicar e o caminho rodapé → perguntas → tradução → landing.

## D27 — Abertura da landing: convencer, a mascote ao lado e o botão da marca (2026-09-06)

**Escolha.**
- A abertura passa a convencer, não a explicar. A headline continua "Planejador semanal online, grátis e sem cadastro."; o parágrafo abaixo virou "Planeje sua semana em segundos com a weeklly, o jeito mais rápido e simples de organizar os seus dias. Grátis, sem login." e a nota sob o botão responde às objeções: "Sem conta, sem cartão, nada para instalar." Como funciona fica para a seção seguinte.
- A abertura tem duas colunas: texto e botão à esquerda, a mascote à direita. A headline caiu de 76 para 64 px e ocupa uma proporção menor da tela; a mascote compensa o lado direito. Por enquanto é a silhueta da marca (partial `mascot.html`, `currentColor` na cor da marca, um halo suave atrás); o cartoon será gerado pelo Miguel num agente de imagem a partir do briefing em `design/brand/MASCOTE.md`, que já diz como trocar a silhueta pela imagem. No celular a mascote fica pequena, acima da headline.
- "Criar minha semana" é o único botão na cor da marca, na landing e no fechamento da página de perguntas; "Abrir planejador" na barra continua monocromático para não competir. No hover o botão treme (`cta-wiggle`, rotação de até 2,5 graus, 800 ms em laço) e a sombra na cor da marca sobe; com "reduzir movimento", só a sombra. Texto branco sobre `#727cf5` dá contraste 3,5:1: passa como componente e texto grande, não como texto comum. Se for preciso AA estrito, o texto do botão vira `#0a0a0a` (5,6:1).

**Razão.** Pedidos do Miguel: a abertura deve convencer de que a weeklly é o jeito mais rápido, prático e fácil de planejar a semana; a headline ocupava demais e quebrava sem proporção; a mascote entra para equilibrar; o botão principal na cor do sistema, com um tremor no hover para convidar ao clique.

**Consequência.** Chaves `landing.lead` e `landing.startNote` reescritas nos dois idiomas; a headline inglesa ganhou um espaço inseparável em "No sign-up" para quebrar em três linhas parelhas. CSS: `.landing-hero` em grade, `.hero-copy`, `.hero-art`, `.landing-cta` na marca com `cta-wiggle`. Testes de navegador: mascote visível, botão principal na cor da marca e o da barra não.

**Revisão (2026-09-06).** Depois do selo da loja (D29), o Miguel aprovou o botão principal no mesmo estilo cartoon: cantos de 16 px, borda de 2 px e sombra sólida de 4 px no tom escuro da marca (`--color-brand-deep`, `#4b53c7`), afundando ao clicar. A cápsula com sombra difusa saiu. O tremor no hover (D28) ficou.

## D28 — Como funciona pela rolagem, tremor do botão e a página de perguntas (2026-09-06)

**Escolha.**
- "Como funciona" avança com a rolagem, não com um relógio. Com o script, a posição do bloco na tela, enquanto a página rola normalmente, escolhe o passo e enche o trilho do passo atual (`--how-progress`, posto pelo script via CSSOM); ver a revisão abaixo para as fronteiras. Clicar num passo rola a página até a posição daquele passo, para os dois ficarem de acordo. Sem script, os rádios são clicáveis e a cena fica parada no passo marcado. O avanço automático por tempo (D26) saiu.
- O tremor do botão "Criar minha semana" segue o "Shake" do CSS-Tricks (Cristian Cortez): só deslocamento horizontal, de 1 a 3 px, com a curva `cubic-bezier(.36,.07,.19,.97)`, sem rotação. Roda em ciclos de 1,4 s enquanto o mouse está em cima: cerca de meio segundo tremendo e o resto parado, um toque de atenção que se repete, não um zumbido. Com "reduzir movimento", só a sombra sobe. Alternativa avaliada e descartada: CSShake (`shake-little`, vibração aleatória contínua a cada 100 ms), forte demais para um botão principal.
- Página de perguntas: coluna centralizada, título "Perguntas frequentes" na cor da marca, respostas descendo suave. A animação é só CSS, sobre `::details-content` com `interpolate-size: allow-keywords` (altura de 0 a `auto`) e `content-visibility` com `allow-discrete`, o que anima também o fechar; navegador sem suporte abre na hora, como antes. Sem script e sem tocar no `name="faq"` que fecha a pergunta anterior.

**Razão.** Pedidos do Miguel: a nota do botão ("Abra em segundos, no navegador."), outro tremor, o avanço dos passos guiado pela rolagem porque a pessoa pode não perceber que os passos são clicáveis e ficar esperando, e a página de perguntas centralizada, mais suave e com a cor da marca no título.

**Consequência.** `landing.startNote` reescrita nos dois idiomas. O script de "como funciona" lê a posição do bloco na tela a cada rolagem e marca o rádio por ela. Teste de navegador: a seção tem a altura do conteúdo, rolar por ela marca 1, 2 e 3, e clicar num passo leva a rolagem até ele.

**Revisão (2026-09-06, mesmo dia).** A primeira versão prendia a seção na tela (`sticky` numa seção de `250svh`) até os três passos passarem. O Miguel pediu que a página role normalmente, sem a seção ter espaço próprio: a pessoa passa por "como funciona" como pelas outras seções e os passos trocam no caminho. Agora a posição do centro do bloco na tela (0 na borda de baixo, 1 na de cima) escolhe o passo: até 0,38 é o passo 1, até 0,6 o passo 2, acima o passo 3, e o trecho de rolagem até a próxima fronteira enche o trilho. Sem altura extra, sem `sticky`. Clicar num passo rola a página até a posição daquele passo.

## D29 — Faixa dos aplicativos na landing (2026-09-06)

**Escolha.** Entre "como funciona" e os pontos fortes entra uma faixa de largura total, como se fosse outra área do produto: fundo na tonalidade da marca (token `--color-band`: `#101228` no tema escuro, `#eceeff` no claro), tudo centralizado, título "Use onde e quando quiser" na cor da marca, uma linha de apoio e o selo do Google Play. O selo é um botão em estilo cartoon (borda grossa, sombra sólida embaixo, afunda ao clicar), monocromático: ícone do Google Play e texto "Disponível no Google Play" em cinza. O ícone é o desenho monocromático do Simple Icons (CC0), não o selo oficial colorido. Enquanto não há app publicado, `playStoreURL` em `seo.go` fica vazia e o selo é um `<span>` com a etiqueta "em breve", sem link; com a URL, vira `<a>`. Para a faixa ocupar a largura toda, o `<main>` da landing deixou de ter a largura de container: cada seção carrega `landing-width`, e a faixa põe o container só no conteúdo.

**Razão.** Pedido do Miguel: uma seção de versões para celular, diferente do resto, com mudança de cor, título na cor principal e o selo cinza em cartoon. A largura total sem `100vw` evita a barra de rolagem horizontal no Windows.

**Consequência.** Chaves `landing.apps.heading`, `landing.apps.lead`, `landing.apps.play`; `seoView.PlayStore`. Testes: ordem das seções e selo sem link enquanto não há URL (Go); fundo diferente do corpo, título na cor da marca e texto do selo (navegador). Pendência: as diretrizes do Google pedem o selo oficial para links à loja; ao publicar o app, decidir se o selo monocromático fica ou se entra o oficial.

## Pendentes para fases futuras

- **Outras formas de entrar** depois do Google (Fase 3): passkeys via `go-webauthn` é a candidata.
- **Hospedagem final** (Fase 5). Recomendação em D7.
- **Monitoramento de erros e uptime** (Fase 5). Recomendação: Sentry e Better Stack.
- **Empacotamento do app** (Fase 6). Ver consequência em D4.

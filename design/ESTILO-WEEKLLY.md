# Estilo visual da weeklly

Análise do projeto local em 06/09/2026. Este guia registra o visual implementado e define como transportá-lo para imagens de comunicação. Não representa um redesenho do site.

## Essência

A weeklly é um planejador semanal online: espaços com nome, sete dias de segunda a domingo, tarefas com horário opcional e salvamento automático. A semana não depende de datas. É possível começar gratuitamente, sem cadastro; entrar com Google é opcional.

O visual combina **uma base monocromática discreta com tipografia arredondada e a silhueta de um esquilo azul-pervinca**. A malha pontilhada sugere um espaço contínuo de planejamento. Os cartões organizam o conteúdo; a marca traz simpatia. A sensação é de clareza, calma e facilidade, sem pressão por produtividade.

## Evidências e precedência

| Fonte local | O que fundamenta |
|---|---|
| [CSS de origem](../web/styles/app.css) | Tokens, tipografia, componentes, temas, responsividade e movimento |
| [Landing](../web/templates/pages/landing.html) | Hierarquia, mascote, cena dos sete dias e benefícios |
| [Textos em português](../internal/i18n/pt.go) | Categoria do produto, funcionalidades e chamadas |
| [Logo aplicado](../web/templates/partials/logo.html) | Desenho do logotipo usado na interface |
| [Mascote aplicado](../web/templates/partials/mascot.html) | Silhueta atual, em `currentColor` |
| [Decisões](../docs/DECISIONS.md) | D17, D22, D24, D26 e D27 |
| [Briefing do cartoon](brand/MASCOTE.md) | Proposta futura de personagem; não confundir com implementação |
| [Landing escura](referencias/landing-desktop-escuro.png) / [clara](referencias/landing-desktop-claro.png) | Capturas locais realizadas nesta análise, viewport de 1440 px |
| [Quadro escuro](../e2e/screenshots/desktop-02-quadro-escuro.png) | Captura existente dos cartões e tarefas, inspecionada nesta análise |

Em caso de divergência, conferir os tokens e templates atuais. Documentos históricos incluem direções já substituídas: papel quente, títulos serifados e coral não descrevem o visual vigente. `web/static/app.css` é CSS compilado; a fonte editável é `web/styles/app.css`.

## Paleta

| Token / função | Escuro | Claro |
|---|---|---|
| `--color-bg`: fundo | `#0a0a0a` | `#fafafa` |
| `--color-surface`: cartões | `#171717` | `#ffffff` |
| `--color-surface-2`: superfície secundária | `#222222` | `#f2f2f2` |
| `--color-border`: borda discreta | `#2e2e2e` | `#e6e6e6` |
| `--color-border-strong`: borda forte | `#454545` | `#c9c9c9` |
| `--color-ink`: texto principal | `#ededed` | `#171717` |
| `--color-ink-muted`: texto secundário | `#a1a1a1` | `#666666` |
| `--color-ink-faint`: apoio discreto | `#6b6b6b` | `#999999` |
| `--color-brand`: marca | `#727cf5` | `#727cf5` |
| `--color-brand-contrast`: texto sobre marca | `#ffffff` | `#ffffff` |
| `--color-danger`: ações destrutivas | `#ff6166` | `#e5484d` |

A cor da marca é fixa; a cor de destaque do aplicativo é uma preferência individual. O padrão do app é monocromático: `--color-accent` acompanha o texto principal. Laranja, âmbar, verde, azul, violeta, rosa e vermelho são opções de personalização, não uma paleta para usar simultaneamente em peças institucionais.

Para comunicação, adotar o tema escuro como base e o azul-pervinca como único acento cromático. Usar cinzas para estrutura e conteúdo; reservar a cor para o esquilo, a chamada principal e poucos estados ativos. Essa seleção para redes sociais é uma orientação deste guia.

## Tipografia e marca

| Uso implementado | Família e tratamento |
|---|---|
| Títulos e nomes | Nunito; títulos principais e nomes dos dias em peso 800 |
| Corpo e controles | `ui-sans-serif, system-ui, -apple-system, Segoe UI, Roboto, Helvetica Neue, Arial, sans-serif` |
| Corpo base | 15 px, entrelinha 1,5 |
| Título da landing | `clamp(38px, 4.6vw, 64px)`, entrelinha 1,06, tracking −0,04 em |
| Apoio da landing | 17–20 px, entrelinha 1,6 |
| Títulos de seção | 28–40 px, entrelinha 1,15, tracking −0,03 em |
| Nome do dia no quadro | 28 px, peso 800, entrelinha 1 |
| CTA principal da landing | 16 px, peso 600; altura de 52 px |

Nunito é servida localmente por `web/static/fonts/nunito.woff2`. Não substituir títulos por serifa, fonte condensada ou monoespaçada. Evitar caixa alta em blocos inteiros e excesso de pesos.

O logotipo é um desenho próprio em SVG, com inicial gestual, diferente dos títulos em Nunito. Reutilizar o arquivo aplicado no site em trabalhos que exijam reprodução exata. Em imagem gerada, uma assinatura textual `weeklly` deve ser tratada como assinatura tipográfica, não como reprodução oficial do logotipo. A grafia tem dois “l”: **w-e-e-k-l-l-y**.

## Geometria e composição

- Fundo pontilhado: pontos de raio aproximado de 1 px, brancos ou pretos a 9% de opacidade, com passo de 24 px. É uma malha regular discreta, não ruído fotográfico.
- Cartões: fundo sólido, borda de 1 px e raio de 18 px (`--radius-card`). Sombras escuras suaves; no tema claro, sombras mais leves.
- Sombra de cartão no escuro: `0 1px 2px rgb(0 0 0 / 40%), 0 24px 48px -28px rgb(0 0 0 / 80%)`.
- Controles: botões em cápsula (`999px`), campos com raios menores, caixas de tarefa arredondadas. Não aplicar o mesmo raio a tudo.
- Landing: largura máxima de 1200 px, margem interna de 24 px; abertura em colunas 7:5, texto à esquerda e esquilo à direita. Intervalos de seção entre 80 e 128 px.
- Quadro real: sete cartões iguais, grandes e quase quadrados, em faixa horizontal rolável com intervalo de 24 px. Não são sete minicolunas que precisam caber na tela. Sábado e domingo têm a mesma importância dos demais dias.
- Cena explicativa da landing: os sete dias cabem em uma faixa de colunas estreitas com barras abstratas de tarefas. Essa simplificação é a referência adequada para peças sociais.
- No celular, a landing passa a uma coluna; a silhueta fica pequena acima do título. Benefícios passam de três colunas para duas e depois uma.

## Ícones, ilustração e efeitos

Ícones funcionais são lineares, monocromáticos, com pontas e junções arredondadas; a landing usa traço de 1,6 no viewBox 20 × 20. Os checks, barras de tarefa e chips de horário têm função informativa.

O esquilo atual é uma **silhueta plana em perfil, voltada para a esquerda, sentado, com cauda grande curvada sobre as costas**. É aplicado em azul-pervinca e pode receber um halo radial suave atrás. O halo e a sombra do CTA são exceções localizadas à superfície predominantemente plana.

O cartoon com olhos e barriga clara de [MASCOTE.md](brand/MASCOTE.md) ainda é uma direção futura. Os tons `#c9cdfb` e `#4b53c7` pertencem a esse briefing, não aos tokens da interface atual. Não apresentar uma nova geração cartoon como personagem oficial já usado no site.

Movimento existente: transição de tema de 560 ms, conclusão de tarefa com check e risco, cena de três passos com avanço de 5,5 s e CTA com tremor de 800 ms no hover. Há tratamento de `prefers-reduced-motion`. Em peças estáticas, representar o estado final de uma ação; não tentar simular movimento com efeitos excessivos.

## Leitura crítica

O conjunto é consistente: a mesma linguagem de cartões, títulos e cinzas conecta o aplicativo à landing. A diferença entre marca fixa e destaque pessoal permite reconhecimento sem tirar a personalização do usuário. A cena dos sete dias comunica melhor a estrutura do produto que uma ilustração genérica de produtividade.

Os riscos para futuras peças são perder a identidade ao usar apenas “SaaS escuro”, transformar tudo em roxo, inventar um esquilo diferente ou desenhar um calendário com datas. Preservar conjuntamente a malha, a Nunito, o esquilo e os sete dias evita essa descaracterização.

Há textos de apoio muito discretos nas capturas, especialmente no rodapé e junto ao CTA. Nas imagens sociais, não usar `ink-faint` para informações essenciais. O CTA do site usa branco sobre pervinca; para a peça, ampliar o texto e usar peso semibold. Uma reprodução em imagem não garante equivalência tipográfica ou cromática exata: revisar sempre o arquivo gerado.

## Aplicação em redes sociais

As medidas a seguir são uma adaptação editorial proposta, não tokens CSS do site nem uma especificação oficial de plataforma.

| Elemento | Padrão proposto para o teste |
|---|---|
| Formato | Vertical 4:5, alvo 1080 × 1350 px |
| Margens | 72 px; todo texto e marca dentro da área segura |
| Título | Cerca de 72 px; duas ou três linhas curtas |
| Definição e benefícios | 30–34 px; sem parágrafos longos |
| Dias | Pelo menos 24 px; segunda a domingo em ordem |
| CTA | Uma cápsula, texto de aproximadamente 32 px semibold |
| Cena | Sete cartões iguais, inteiramente visíveis; tarefas como barras |

A peça deve responder: o que é, como se organiza e por que começar. Voz direta em português brasileiro, acolhedora, sem urgência artificial. Preferir “organizar seus dias”, “grátis e sem cadastro”, “salvas automaticamente” e “Criar minha semana”. Não prometer IA, notificações, funcionamento offline, apps publicados ou integrações não confirmadas. Não inventar domínio, QR code, avaliação ou número de usuários.

## Prompts e entrega

- [Prompt reutilizável](PROMPTS-IMAGENS.md): bloco fixo de identidade e campos para novas peças.
- [Prompt completo do teste](PROMPT-TESTE-WEEKLLY.txt): texto efetivamente usado para apresentar a weeklly.
- [Imagem do teste](social/weeklly-o-que-e.png): resultado da geração integrada de imagens, revisado visualmente.

Antes de aceitar uma imagem: conferir grafia e acentos, legibilidade em tamanho reduzido, sete dias na ordem correta, ausência de datas, uso contido de cor, consistência do esquilo e fidelidade das afirmações. As capturas de referência documentam o site; a arte social é uma adaptação, não um screenshot do produto.

# Referência de landing: Duolingo

Levantamento feito em 2026-09-06 abrindo `https://www.duolingo.com/` em português, num navegador real (a página é renderizada por script; um `fetch` simples devolve só o título). Os textos abaixo são os exatos da página naquele dia. Serve de referência para a landing da weeklly (`/planejador-semanal`, `/en/weekly-planner`).

Não existe skill nem documento pronto que descreva a landing da Duolingo. O que existe e vale usar:

| Fonte | O que dá | Onde |
|---|---|---|
| Skill `copywriting` (instalada) | Esqueleto de landing: headline, subheadline, CTA, prova, problema, benefícios, como funciona, objeções, CTA final; fórmulas de headline e regras de texto | `~/.claude/skills/copywriting` |
| Skill `copy-editing` (instalada) | Revisão linha a linha do texto pronto | idem |
| Skill `ab-testing` (instalada) | Testar duas headlines ou duas ordens de seção quando houver tráfego | idem |
| Mobbin, "Duolingo Web Landing page Flow" | Capturas da landing e do fluxo de cadastro, atualizadas por eles | https://mobbin.com/explore/flows/0daa42cb-13de-4f86-8889-639821bec0d5 |
| Figma Community, "Duolingo.com → Web Pages UI" | Telas recriadas em Figma, para medir espaçamentos e tipografia | https://www.figma.com/community/file/1349313332454280331/duolingo-com-web-pages-ui |
| Landing Metrics, "Duolingo Landing Page Design Inspiration" | Análise curta do acima da dobra (headline, hierarquia de botões, seletor de cursos) | https://www.landingmetrics.com/landing-page-design-example/duolingo |
| Este documento | A estrutura inteira, seção a seção, e o que aplicar na weeklly | `docs/LANDING-REFERENCE.md` |

## A página, seção a seção

Página de 7.864 px de altura em 1440 de largura. Fundo branco, uma cor de marca (verde `#58cc02`), ilustrações grandes em todas as seções, muito espaço vazio. Nenhuma foto, nenhum depoimento, nenhum número.

### 1. Barra

Logo (ícone da coruja + "duolingo") à esquerda. Um único botão à direita, verde, texto em caixa alta: **COMECE AGORA**. Um botão discreto de **IDIOMA DO SITE: PORTUGUÊS** perto do logo. Nada mais: sem menu, sem "recursos", sem "preços".

### 2. Abertura (hero)

Duas colunas. Esquerda: ilustração com a mascote (Duo) e personagens em volta de um celular. Direita: headline centralizada e dois botões empilhados.

- H1: **"O jeito mais divertido de aprender idiomas, xadrez e muito mais!"**
- Sem subheadline.
- Botão primário: **COMECE AGORA** (`/register`). Botão secundário, contornado: **JÁ TENHO UMA CONTA**.

A headline é uma promessa em uma linha, com a categoria dentro ("aprender idiomas"). Os dois botões separam quem chega pela primeira vez de quem volta.

### 3. Faixa de cursos

Uma faixa horizontal com setas, logo abaixo da abertura, listando os pontos de entrada: **INGLÊS, XADREZ, ESPANHOL, FRANCÊS, ALEMÃO, ITALIANO, JAPONÊS, COREANO, CHINÊS (SIMPLIFICADO)**, cada um com uma bandeira. Cada item é um link para a página do curso. Funciona como prova de amplitude e como segundo caminho de entrada, sem cadastro.

### 4. Seções em ziguezague (quatro)

Cada uma ocupa a tela inteira: ilustração de um lado, título grande do outro, um parágrafo curto abaixo do título. A ilustração alterna de lado a cada seção. Os títulos são em minúsculas, na fonte da marca, em verde; o parágrafo é cinza.

| Título (H2) | Parágrafo |
|---|---|
| **grátis. divertido. eficaz.** | "Aprender com o Duolingo é divertido, e pesquisas comprovam que funciona mesmo! Com lições rápidas e curtinhas, você ganha pontos e desbloqueia novos níveis enquanto aprende como se comunicar na vida real." |
| **baseado na ciência** | "Combinamos metodologias baseadas em pesquisas com um conteúdo encantador para criar cursos eficazes que ensinam leitura, escrita, escuta e fala!" |
| **mantenha a motivação** | "Fica fácil criar o hábito de aprender idiomas com recursos que parecem de jogo, desafios divertidos e lembretes do nosso mascote simpático, a coruja Duo." |
| **aprendizado feito para você** | "As lições combinam o melhor da inteligência artificial e da ciência da linguagem e são feitas sob medida para ajudar você a aprender no nível e ritmo certos." |

A primeira seção repete os três adjetivos da marca. Cada seção defende uma ideia só. Nenhuma tem botão: a página confia no botão da barra e no do fechamento.

### 5. Apps

Título centralizado, **"aprenda onde e quando quiser"**, com os dois selos de loja (**Baixe na App Store**, **DISPONÍVEL NO Google Play**) e celulares e ícones flutuando ao redor. É a única seção com fundo "cheio" de elementos.

### 6. Faixa escura: Super Duolingo

Fundo azul-marinho, tipografia diferente, um celular ilustrado e o botão branco **TESTE UMA SEMANA GRÁTIS**. É a venda do plano pago, isolada visualmente do resto para não contaminar o "grátis" da abertura.

### 7. Duolingo English Test

Volta ao fundo branco: título verde, parágrafo, botão contornado **CERTIFIQUE O SEU INGLÊS**. Um segundo produto, em posição de "se você já chegou até aqui".

### 8. Fechamento

Título grande centralizado, **"aprenda um idioma com o duolingo"**, botão **COMECE AGORA** e a mascote saindo de um celular, com moedas e ícones ao redor. A ilustração invade o rodapé verde: a transição entre a última seção e o rodapé é uma curva, não uma linha.

### 9. Rodapé

Fundo verde. Cinco colunas de links brancos em negrito, mais uma sexta abaixo da quinta:

- **Quem somos**: Cursos, Missão, Método, Eficácia, Cultura Duolingo, Pesquisa, Carreiras, Guia da marca, Loja, Imprensa, Investidores, Entre em contato
- **Produtos**: Duolingo, Duolingo for Schools, Duolingo English Test, Podcast, Duolingo for Business, Super Duolingo, Dê o Super Duolingo de presente, Duolingo Max
- **Aplicativos**: Duolingo para Android, Duolingo para iOS
- **Ajuda e suporte**: Dúvidas: Duolingo, Dúvidas: Escolas, Dúvidas: Duolingo English Test, Status
- **Termos e privacidade**: Normas da comunidade, Termos de uso, Privacidade
- **Redes sociais**: Blog, Instagram, TikTok, Twitter, YouTube, LinkedIn

Abaixo, "Idioma do site:" com os 30 idiomas como links para subdomínios (`pt.duolingo.com`, `en.duolingo.com`).

## O que faz a página funcionar

1. **Um botão, repetido.** "Comece agora" na barra, na abertura e no fechamento, sempre o mesmo texto e a mesma cor. Não há segundo caminho competindo com ele.
2. **Headline com a categoria e a promessa.** "O jeito mais divertido de aprender idiomas": quem chega sabe o que é e por que ficar. Nada de metáfora.
3. **Os três adjetivos da marca viram título.** "grátis. divertido. eficaz." é a proposta inteira em três palavras, e a primeira seção depois da dobra.
4. **Uma ideia por tela.** Cada seção do ziguezague tem título de duas ou três palavras e um parágrafo de três linhas. A ilustração faz metade do trabalho.
5. **Mascote como fio condutor.** A coruja aparece na abertura, numa seção ("lembretes do nosso mascote simpático, a coruja Duo"), no fechamento e no ícone. É personagem, não decoração.
6. **Ponto de entrada sem cadastro.** A faixa de cursos leva direto a uma página de curso; a pessoa começa antes de decidir criar conta.
7. **Cor de marca em tudo que é clicável ou é título.** Verde nos botões, nos títulos e no rodapé; o resto é preto, cinza e branco. O que é pago fica numa faixa de outra cor.
8. **Tipografia própria e minúsculas.** Títulos arredondados em minúsculas dão o tom brincalhão sem uma palavra a mais.
9. **Botões "físicos".** Sombra inferior sólida que some ao clicar: o botão parece uma tecla. Texto em caixa alta com espaçamento.
10. **Rodapé como mapa da empresa.** Seis grupos, sem enfeite, e o idioma do site no fim.

## O que não copiar

- Ilustrações em cada seção exigem um sistema de ilustração próprio. Sem ele, o resultado vira clipart. A weeklly tem uma mascote em silhueta; ela pode ser o elemento gráfico, em escala grande, no lugar de cenas ilustradas.
- Faixas de produto pago (Super Duolingo) e produtos irmãos (English Test) não existem na weeklly. A página deve terminar no FAQ e no rodapé, como o Miguel pediu.
- "Divertido" é promessa da Duolingo, não nossa. Os nossos adjetivos estão na decisão do Miguel: grátis, sem cadastro, sem datas.
- Caixa alta nos botões conflita com a tipografia da weeklly (Nunito nos títulos, sistema no resto). Manter os botões como no app.

## Aplicação na weeklly

Estrutura proposta, com os nossos textos, mantendo a fundação de SEO (D23) e as decisões da revisão anterior (barra com tema e idioma, headline concreta, FAQ por último, rodapé com seis grupos):

| Seção | Como a Duolingo faz | Como fica na weeklly |
|---|---|---|
| Barra | Logo, idioma, um botão | Logo, sol/lua, PT/EN, **Criar minha semana** (um botão só, o mesmo da abertura; "Abrir planejador" sai da barra e vai para o FAQ ou para o rodapé, em "Produtos") |
| Abertura | Ilustração à esquerda, headline e dois botões à direita | Esquilo em silhueta grande à esquerda (na cor da marca, `--color-brand` `#727cf5`, sobre o fundo), headline à direita: "Planejador semanal online, grátis e sem cadastro." Botão primário **Criar minha semana**; secundário contornado **Já tenho semanas** (abre `/`) |
| Faixa de entrada | Cursos com bandeira | Os sete dias como uma faixa de cartões pequenos (seg a dom) rolando, sem ser uma demonstração com tarefas: só os nomes, para mostrar a forma do produto. Alternativa: pular esta faixa |
| Ziguezague | Quatro seções, título + parágrafo + ilustração | Três seções, cada uma com um ponto forte e o esquilo em pose diferente (ou um recorte do próprio quadro, sem tarefas de exemplo): **grátis. sem cadastro. sem datas.** / **sete dias do mesmo tamanho** / **horário só quando faz sentido** |
| Apps | Selos das lojas | Ainda não há app: fica fora até existir (backlog, Fase 6) |
| Fechamento | Título + botão + mascote invadindo o rodapé | Os pontos fortes fecham a página (o FAQ virou página própria, D26); o esquilo pode aparecer num fechamento com botão ou no rodapé, ao lado da marca, como assinatura |
| Rodapé | Seis grupos + idiomas | Já está assim; falta ligar as páginas (Sobre, Contato, Termos, Privacidade) e as redes |

Feito em 2026-09-06 (D26): a barra ficou sem o atalho "Como funciona"; "como funciona" virou três passos ligados por um trilho e uma cena em SVG que muda com o passo (o nome sendo digitado, sete colunas, barras entrando, uma feita e uma mudando de dia); as perguntas frequentes ganharam página própria, ligada no rodapé.

Feito em 2026-09-06 (D27): abertura em duas colunas, texto que convence à esquerda ("Planeje sua semana em segundos…"), a mascote à direita (silhueta até o cartoon existir; briefing e prompts em `design/brand/MASCOTE.md`), botão "Criar minha semana" na cor da marca com tremor no hover.

Ainda em aberto, o Miguel decide: (a) o esquilo em cartoon substitui a silhueta na abertura assim que for gerado; (b) se a faixa dos sete dias entra; (c) se a barra fica com um botão só. Depois disso, aplicar a skill `copywriting` para as três seções do ziguezague e a `copy-editing` no texto final.

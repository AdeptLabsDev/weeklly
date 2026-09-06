# Mascote da weeklly: o esquilo

O esquilo é a mascote da weeklly (D24). A silhueta em `squirrel-icon-menu.svg` já é o ícone do site e ocupa a abertura da landing enquanto o cartoon não existe. Este documento é o briefing para gerar o cartoon num agente de imagem, com os prompts prontos, e o passo a passo para colocá-lo na página.

## Por que um esquilo

O esquilo planeja: guarda hoje o que vai precisar amanhã. É o animal que pensa na semana que vem, sem pressa e sem drama. A personalidade da mascote sai daí: rápida, organizada, simpática, um pouco esperta. Nunca ansiosa, nunca "produtividade tóxica".

## Identidade visual (vale para todos os prompts)

| Item | Decisão |
|---|---|
| Cor principal | Pelo em `#727cf5` (a cor da weeklly). Barriga e focinho num tom mais claro, `#c9cdfb`. Sombras e contorno, quando houver, em `#4b53c7`. |
| Olhos | Grandes, escuros (`#0a0a0a`), com um brilho branco. É o que dá simpatia. |
| Cauda | Grande, cheia, curvada sobre as costas, a assinatura da silhueta atual. |
| Traço | Vetor plano (flat), formas arredondadas, duas ou três camadas de tom, sem gradiente, sem textura, sem pelo realista. |
| Fundo | Transparente. A mascote aparece sobre `#0a0a0a` no tema escuro e sobre `#fafafa` no claro; tem que ler bem nos dois, por isso o contorno escuro é opcional mas o tom claro da barriga é obrigatório. |
| Texto | Nenhum texto na imagem, nunca. |
| Proporção | Cabeça grande em relação ao corpo (cerca de 1:2), como personagens de app. |

Regra de ouro: a mascote é um personagem, não um enfeite. Cada pose tem uma intenção clara.

## Como usar os prompts

1. Gere primeiro a **folha de personagem** (prompt 1). Escolha a versão que mais gosta. Ela vira a referência de todas as outras.
2. Gere as poses (prompts 2 a 5) **anexando a folha de personagem** como imagem de referência, quando o agente aceitar. Repita o bloco "Character" em todos os prompts, sem editar, para manter o mesmo esquilo.
3. Peça sempre **PNG com fundo transparente, 2048 por 2048**. Se o agente exportar SVG, melhor ainda: é o formato ideal para a página.
4. Cole também o **negative prompt** quando o agente tiver esse campo.

Os prompts estão em inglês porque os modelos de imagem respondem melhor nesse idioma.

### Bloco comum (cole em todos)

```
Character: a friendly cartoon squirrel mascot for a weekly planner app called weeklly. Periwinkle blue fur (#727cf5), lighter periwinkle belly and muzzle (#c9cdfb), darker periwinkle shading (#4b53c7). Big dark eyes with a white highlight, small rounded ears, a large fluffy tail curling over the back. Flat vector illustration, clean rounded shapes, two or three flat shading tones, no gradients, no texture, no outline or a thin dark periwinkle outline. Big head, small body, app-mascot proportions. Transparent background, no text, no logo, no watermark.
```

### Negative prompt (cole em todos)

```
text, letters, watermark, logo, background, scenery, realistic fur, photorealistic, 3D render, gradients, glossy, shadow on the ground, extra limbs, multiple characters, brown fur, orange fur
```

### 1. Folha de personagem

```
[Bloco comum]
Character sheet, front view, three-quarter view and side view of the same squirrel standing, neutral happy expression, arms relaxed. All three views aligned on one row with the same height, evenly spaced, generous margins. Consistent colors and proportions across all views. 2048 x 2048, transparent background.
```

### 2. Abertura da landing (a principal)

Onde entra: à direita da headline "Planejador semanal online, grátis e sem cadastro." O esquilo apresenta o produto.

```
[Bloco comum]
Hero pose: the squirrel stands in three-quarter view facing slightly left, holding up a small blank card with one paw as if presenting it, the other paw giving a thumbs-up. The card is a simple rounded rectangle in light periwinkle with seven thin vertical bars, no text. Confident, welcoming smile, eyes toward the viewer. Full body, feet visible, tail curling up behind. Centered, with margins around the character. 2048 x 2048, transparent background.
```

Variação sem objeto, caso o cartão não fique bom:

```
[Bloco comum]
Hero pose: the squirrel stands in three-quarter view facing slightly left, waving with one paw, the other paw on the hip, big welcoming smile. Full body, tail curling up behind. Centered with margins. 2048 x 2048, transparent background.
```

### 3. Tarefa feita (para a seção "como funciona" ou o estado de semana concluída)

```
[Bloco comum]
The squirrel holds a big pencil with both paws and has just drawn a large check mark floating beside it, in periwinkle blue. Satisfied expression, one eye slightly squinted, small smile. Full body, three-quarter view. 2048 x 2048, transparent background.
```

### 4. Planejando (o esquilo guarda para a semana)

```
[Bloco comum]
The squirrel sits and stacks three acorns in a neat row in front of it, as if organizing them for later, looking down at them with focused, content expression. Side view, full body, tail up. 2048 x 2048, transparent background.
```

### 5. Descansando (páginas "em breve", vazias ou de erro)

```
[Bloco comum]
The squirrel is curled up asleep on its own fluffy tail, eyes closed, tiny "z" shapes are NOT allowed, no text. Peaceful, cute. Compact composition, seen from the side. 2048 x 2048, transparent background.
```

## Depois de gerar

1. Salve os PNGs em `design/brand/mascot/` com estes nomes: `sheet.png`, `hero.png`, `done.png`, `planning.png`, `sleeping.png`. Se houver SVG, salve com o mesmo nome e `.svg`.
2. Para a landing, a imagem entra em `web/static/mascot-hero.png` (ou `.svg`), reduzida para no máximo 800 px de largura e comprimida (WebP com transparência serve, `mascot-hero.webp`).
3. Em `web/templates/pages/landing.html`, o bloco `<div class="hero-art">{{template "mascot"}}</div>` troca a silhueta por
   `<img class="mascot" src="{{asset "mascot-hero.webp"}}" width="800" height="800" alt="" decoding="async">`.
   O `alt` fica vazio de propósito: a mascote é decorativa, o sentido está na headline ao lado.
4. O CSS de `.hero-art .mascot` já dimensiona a imagem nos três tamanhos de tela; nada a ajustar.
5. O ícone do site (`favicon.svg`) continua sendo a silhueta: ícone precisa de forma simples em 16 px.

## Pendências

- Nome da mascote (se tiver).
- Se a cor `#727cf5` do pelo fica boa em cartoon ou se o esquilo deve ser marrom com um detalhe na cor da marca. Os prompts assumem pelo periwinkle, como o ícone; é a escolha mais forte de marca (a coruja da Duolingo é verde, não marrom).
- Animação da mascote na abertura: sem biblioteca de terceiros (regra do projeto), a saída é SVG animado por CSS ou uma sequência de PNGs em `steps()`. Decidir depois de ver o cartoon parado.

# Prompts de imagens para a weeklly

Base visual: [guia de estilo](ESTILO-WEEKLLY.md). O teste completo está em [PROMPT-TESTE-WEEKLLY.txt](PROMPT-TESTE-WEEKLLY.txt).

## Prompt reutilizável

Copie o bloco abaixo e preencha os campos entre colchetes. Para preservar a identidade, forneça a captura [landing-desktop-escuro.png](referencias/landing-desktop-escuro.png) como referência visual junto do pedido. Ela contém a silhueta realmente aplicada no site.

```text
Crie uma única imagem para a marca weeklly.
Uso: [post explicativo / apresentação de recurso / capa].
Formato: [proporção e dimensões desejadas].
Objetivo: [uma ideia que a pessoa precisa entender].
Conteúdo de produto confirmado: [informações verificadas].
Texto exato: [título, apoio curto, rótulos e uma chamada; não acrescentar texto].

IDENTIDADE FIXA
A weeklly é um planejador semanal online com sete dias de segunda a domingo,
sem datas fixas, tarefas com horário opcional e salvamento automático.
Design digital 2D limpo, organizado, calmo e acolhedor.
Fundo #0a0a0a com pontos minúsculos brancos a 9% de opacidade numa malha
regular discreta. Cartões #171717, superfícies secundárias #222222,
bordas finas #2e2e2e, títulos #ededed, texto secundário #a1a1a1.
Azul-pervinca #727cf5 é o único acento cromático: esquilo, chamada principal
e poucos sinais ativos. Cores planas e sombras suaves. Halo localizado
atrás do esquilo é permitido, como na referência, com intensidade baixa.
Títulos semelhantes à Nunito ExtraBold 800, sans-serif arredondada,
entrelinha compacta; corpo sans-serif de sistema, nítido e legível.
Texto alinhado à esquerda, respiro generoso, bordas discretas,
cartões arredondados e uma única chamada em cápsula.
Se usar o esquilo, preservar a silhueta plana da referência: sentado,
perfil para a esquerda, cauda grande curvada sobre as costas, preenchimento
pervinca uniforme. Sem rosto cartoon, olhos, pelo ou renderização 3D.
Grafia exata da assinatura: weeklly, com dois “l”. Não criar outro logotipo.

COMPOSIÇÃO VARIÁVEL
[Descrever uma cena que explica o conteúdo. Usar poucos elementos.]
Se mostrar uma semana inteira, desenhar exatamente sete cartões iguais
em uma única faixa: seg, ter, qua, qui, sex, sáb, dom.
Tarefas podem ser barras abstratas com checks; não inventar microtexto.
Não inserir datas ou reduzir o fim de semana.
A captura anexada define o estilo, não uma página inteira a ser copiada.
Em formato vertical, reorganizar os elementos e ampliar os textos.

RESTRIÇÕES
Não usar fotos, 3D, neon, vidro, grão, textura de papel, arco-íris,
gradientes decorativos, mascote marrom, coruja ou estética da Duolingo.
Não adicionar funcionalidades, métricas, preços, domínio, QR code,
marcas-d'água ou afirmações que não foram fornecidas.
Português brasileiro com acentos corretos. Sem texto cortado.
Hierarquia: título, explicação, cena, chamada. Nenhum excesso decorativo.
Entregar somente a arte final.
```

Para uma versão clara, trocar exclusivamente os neutros pelos valores do tema claro no guia; manter a marca `#727cf5`. Não misturar os dois temas sem um objetivo explícito de comparação.

## Teste: o que é a weeklly

O [prompt completo](PROMPT-TESTE-WEEKLLY.txt) já está preenchido e pode ser reutilizado integralmente. Ele solicita uma peça 4:5 com título “Sua semana, em um só lugar.”, definição do produto, uma faixa de sete dias e os benefícios: grátis e sem cadastro, ausência de datas fixas e salvamento automático.

A cena segue a ilustração simplificada da landing; não pretende reproduzir a geometria dos grandes cartões roláveis do aplicativo. A assinatura textual também não substitui o SVG oficial da marca.

Resultado salvo em [social/weeklly-o-que-e.png](social/weeklly-o-que-e.png), gerado pela ferramenta integrada de imagens a partir do prompt completo e da captura da landing escura. Não foi usado o fluxo CLI/API. A revisão visual confirmou os textos, os sete dias em ordem, a ausência de datas e a silhueta coerente com a referência. O gerador acrescentou leves variações tonais no azul; a peça mantém a direção visual, mas não deve ser usada para extrair os tokens exatos da marca.

Texto alternativo sugerido: “Apresentação da weeklly em fundo escuro pontilhado, com esquilo azul e sete cartões de segunda a domingo. Um planejador semanal online, grátis e sem cadastro, sem datas fixas e com tarefas salvas automaticamente.”

Dimensões efetivamente entregues pelo gerador: **1122 × 1402 px**, aproximadamente 4:5, PNG. O pedido de 1080 × 1350 px no prompt era o tamanho ideal; o original foi preservado sem redimensionamento.

## Revisão e correções

Confira a arte em tamanho integral e em largura próxima à de um celular. Conte os sete dias, confira “weeklly” e “sáb”, leia todas as frases e observe se a cor da marca está contida.

Se precisar corrigir, use uma instrução focada, por exemplo:

```text
Edite somente [elemento incorreto] para [correção exata]. Preserve toda
a composição, as cores, o texto restante, a silhueta do esquilo e os sete
cartões. Não acrescente elementos nem reformule o estilo.
```

A geração é raster e pode aproximar fontes, cores e marcas. Para reprodução exata do logotipo, aplicar o SVG original em uma etapa de composição. Não assumir que a imagem gerada é um novo arquivo oficial da identidade.

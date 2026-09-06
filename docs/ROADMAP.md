# Weekly Planner — Roadmap

> Estado: Fase 0 construída e núcleo da Fase 1 entregue (tarefas por dia com salvamento no ato), com hub, seletor, ações de semana e login com o Google adiantados das Fases 2 e 3. Monocromático até a cor de destaque ser definida. Falta CI verde no GitHub e aprovação visual. Atualizado em 2026-09-06.
> Decisões em [DECISIONS.md](DECISIONS.md). Stack fechada: Go. Acompanhamento em [BACKLOG.md](BACKLOG.md).

## 0. Em uma frase

Um planejador semanal pessoal: uma semana com nome, sete dias lado a lado num quadro fixo, texto livre em cada dia, salvo sozinho, aberto em menos de um segundo. Site primeiro, app depois.

## 1. Tese e princípios

**A dor.** O planejamento semanal de uma pessoa costuma ser o mesmo por várias semanas. Calendários obrigam a repetir cada tarefa em cada data, e isso vira trabalho. Aqui a semana é um espaço só: o usuário monta uma vez e consulta.

**Tese.** A maioria dos planejadores falha por excesso: listas, tags, projetos, recorrências. Este produto aposta no oposto. Uma semana, sete blocos de texto, zero cerimônia. O valor está na velocidade de consulta e na ausência de atrito para escrever.

**Princípios, em ordem de prioridade**

1. **Consulta instantânea.** Abrir o app é ver a semana atual. Sem cliques, sem loading perceptível, com o dia de hoje destacado.
2. **Escrever é gratuito.** Clicar num dia e digitar. Nada de modal, botão de salvar ou confirmação. O texto nunca se perde.
3. **Finito por desenho.** Sete dias, um quadro de tamanho fixo. Não é canvas infinito. A restrição é o produto.
4. **Uma semana por vez.** Só a semana selecionada está na tela. As outras vivem no hub.
5. **Os dados são do usuário.** Exportáveis a qualquer momento, em formato aberto.

**O que este produto não é**

- Não é o Miro. O Miro é referência estética (quadro, cartões, grid de pontos, toolbar flutuante), não funcional.
- Não é gestor de tarefas. Sem status, prioridade, projeto, responsável.
- Não é colaborativo. Um dono por semana.
- Não é calendário. Sem horários, sem eventos, sem integração com agenda.

## 2. Modelo do produto

### Conceitos

| Conceito | Definição |
|---|---|
| **Semana** | Um espaço de planejamento com nome, dado pelo usuário: "Semana padrão", "Semana de provas". Sete dias, de segunda a domingo, sem datas. Monta-se uma vez e consulta-se sempre. |
| **Dia** | Uma posição na semana (segunda a domingo) e a lista de tarefas daquele dia, na ordem que a pessoa deu. Tem seu próprio cartão no quadro. |
| **Tarefa** | Um item do plano de um dia: título, horário opcional e se já foi feita. |
| **Quadro** | O espaço aberto onde os sete cartões vivem, lado a lado, do mesmo tamanho. Grid de pontos na página inteira. Desliza só na horizontal; sem zoom. |
| **Hub** | A lista das semanas do usuário, as usadas mais recentemente primeiro. Cria semanas novas e troca a semana visível. |

### Invariantes (o sistema nunca quebra)

1. Uma semana tem exatamente sete dias, de segunda a domingo. Toda tarefa pertence a um deles.
2. Uma semana é identificada por um id opaco, gerado pelo sistema. O nome é livre, de 1 a 60 caracteres, e pode mudar.
3. Todo dia pertence a exatamente uma semana.
4. Toda edição é persistida no ato: adicionar, editar, concluir ou excluir uma tarefa já salva. Não existe botão "salvar".
5. Apenas uma semana é visível por vez.
6. O quadro é finito: sete cartões iguais numa faixa horizontal, sem zoom e sem rolagem vertical. O usuário desliza para o lado para ver os dias. O conteúdo de um dia rola dentro do seu cartão, nunca estoura o quadro.
7. "Hoje" é o dia da semana atual no fuso horário do usuário. É a única coisa que o calendário decide.

### Modelo de dados (conceitual, independente de stack)

```
Semana
  id              → opaco, 16 caracteres, gerado pelo sistema
  dono            → usuário
  nome            → texto livre, 1 a 60 caracteres
  tarefas[]
    id
    diaDaSemana   → 0 (segunda) a 6 (domingo)
    posição       → ordem manual dentro do dia
    título        → 1 a 200 caracteres
    horário       → HH:MM ou nenhum
    feita
    atualizadaEm  → por tarefa, para sincronização
  criadaEm
  atualizadaEm    → muda quando qualquer tarefa muda; ordena o hub

Usuário
  ordemDasSemanas → recentes ou A–Z
```

Por que `atualizadaEm` por tarefa e não só por semana: a sincronização entre dispositivos resolve conflitos no nível da tarefa (último a escrever vence). O domínio de conflito fica minúsculo e previsível.

### Fluxos principais

| Fluxo | Passos | Meta |
|---|---|---|
| Consultar | Abrir o app → a última semana usada aparece, com o dia de hoje no centro | < 1 s frio, < 100 ms com cache |
| Planejar | Escrever a tarefa no pé do cartão, horário se quiser, Enter → ela aparece e "Salvo" | persistida em < 1 s |
| Editar | Clicar no título ou no horário → mudar → Enter. Marcar a caixa conclui. | no ato |
| Criar semana | Hub → "Nova semana" → dar um nome → quadro abre vazio | 2 cliques |
| Trocar semana | Hub → clicar na semana | < 100 ms |
| Sair | Fechar a aba a qualquer momento | zero perda de texto |

### Endereçamento

Cada semana tem uma URL própria pelo id, no formato `/semana/k7m2p9xq4w3n8r5t`. Isso dá deep link, histórico do navegador e um contrato simples para o app reutilizar. O id é opaco de propósito: não revela quantas semanas existem nem permite adivinhar outras.

### Esboço do quadro

Faixa horizontal de sete cartões iguais (D11). A tela mostra dois ou três por vez; o resto está ao lado, fora da borda, e o usuário desliza. Abaixo da faixa, o paginador com os sete dias abreviados mostra quais estão visíveis e leva ao dia clicado.

```
 ·   ·   ·   ·   ·   ·   ·   ·   ·   ·   ·   ·   ·   ·   ·   ·   ·   ·
 Semana padrão                                        ( Ir para hoje )
 Hoje é sábado.
 ·   ·   ·   ·   ·   ·   ·   ·   ·   ·   ·   ·   ·   ·   ·   ·   ·   ·
      ┐   ┌────────────────┐   ┌────────────────┐   ┌────────────
      │   │ Sexta          │   │ Sábado   hoje  │   │ Domingo
      │   ├────────────────┤   ├────────────────┤   ├────────────
      │   │ Revisão da     │   │ Feira          │   │ Sem plano
      │   │ semana         │   │ Caminhada      │   │ ainda
      │   │                │   │                │   │
      ┘   └────────────────┘   └────────────────┘   └────────────
 ·   ·   ·   ·   ·   ·   ·   ·   ·   ·   ·   ·   ·   ·   ·   ·   ·   ·
                  seg   ter   qua   qui   sex   sáb   dom
```

## 3. Fases

Sem datas: projeto pessoal, o ritmo é seu. Tamanho indica esforço relativo (S, M, L). Cada fase tem um critério de saída objetivo. Só avança quando ele é atendido.

### Fase 0 — Fundação (próxima etapa) · M

**Objetivo.** Decidir a stack e a direção de design, e criar o repositório no padrão Higher Mind.

**Entregas**
- Decisão de stack documentada com razão para cada camada: interface, persistência local, persistência remota, autenticação, hospedagem. Registrada em `docs/DECISIONS.md`.
- Repositório iniciado via `/hm-init`: lint, typecheck, testes e CI rodando no primeiro commit.
- Direção de design: tokens (cor, tipografia, espaçamento, sombra, grid de pontos), gramática do quadro Miro traduzida para dark-first, layout desktop decidido.
- Modelo de dados escrito e revisado contra as invariantes.
- Uma tela do quadro renderizada com dados fictícios, aprovada visualmente.

**Critério de saída.** Repo com CI verde, uma tela do quadro que já não parece template, nenhuma decisão técnica sem razão escrita.

### Fase 1 — Núcleo: uma semana · L

**Objetivo.** Usar o produto de verdade na sua própria semana, ainda sem hub e sem conta.

**Entregas**
- Uma semana real, criada com nome, com os sete cartões. Hoje destacado pelo fuso do navegador, não mais pelo fuso configurado no servidor.
- Editor de texto livre por dia, com o cartão rolando internamente.
- Autosave em três gatilhos: pausa na digitação, perda de foco, saída da página. Indicador discreto de estado (salvo, salvando).
- Persistência local. Não é descartável: vira a camada de cache da Fase 3.
- Atalhos básicos: navegar entre dias pelo teclado, foco visível.
- O arrastar da faixa ignora o editor: arrastar dentro do texto seleciona, arrastar fora navega.
- Testes do dia da semana por fuso, do autosave e da persistência.

**Critério de saída.** Você usa por sete dias seguidos sem perder uma tecla e sem pensar no app.

### Fase 2 — Hub e múltiplas semanas · M

**Objetivo.** Várias semanas com navegação que não parece carregar.

**Entregas**
- Hub: lista das semanas do usuário, as usadas mais recentemente primeiro. Semana aberta marcada. Semanas com conteúdo distinguíveis das vazias.
- Criar semana: pedir só o nome e abrir o quadro vazio. Renomear e excluir a partir do hub.
- Trocar semana em menos de 100 ms, com as semanas vizinhas pré-carregadas.
- Abrir o app volta para a última semana usada.
- URL por semana, histórico do navegador funcionando.
- Excluir semana com confirmação e desfazer.

**Critério de saída.** Navegar entre vinte ou mais semanas sem perceber carregamento.

### Fase 3 — Conta e sincronização · L

**Objetivo.** Os mesmos dados em qualquer dispositivo, com isolamento e segurança.

**Entregas**
- Autenticação sem senha (link mágico ou passkey). Sessão segura, expiração, logout em todos os dispositivos.
- Isolamento por usuário em todas as consultas. Nenhum caminho lê dados de outro dono.
- Sincronização local ↔ remoto: escrita otimista, fila offline, último a escrever vence por dia com relógio do servidor.
- Estados visíveis: salvo, sincronizando, offline, conflito resolvido.
- Migração dos dados locais das Fases 1 e 2 para a conta, sem perda.
- Exportação completa em JSON e Markdown. Backups diários automatizados com restauração testada.
- Revisão de segurança com `/hm-engineer`.

**Critério de saída.** Editar no notebook e abrir no celular mostra a mesma coisa. Desligar a rede, editar, religar: sincroniza sem perda nem duplicação.

### Fase 4 — Encantamento · M

**Objetivo.** Chegar na barra de design: Miro como referência, Higher Mind como padrão.

**Entregas**
- Motion com propósito: transição entre semanas, entrada dos cartões, indicador de salvamento. Respeita reduced motion.
- Mobile web de verdade: um dia por vez com navegação por dias, sem encolher o desktop.
- Estados vazios com voz própria. Onboarding de uma tela.
- Tema claro como opção.
- Acessibilidade: teclado completo, contraste AA, leitores de tela nos cartões e no hub.
- Revisão com `/hm-designer`.

**Critério de saída.** Aprovado em `/hm-designer` sem ressalvas.

### Fase 5 — Lançamento web · S

**Objetivo.** Público, monitorado, seguro.

**Entregas**
- Domínio, HTTPS, cabeçalhos de segurança, rate limiting.
- PWA instalável: ícone, tela inicial, leitura e escrita offline.
- Orçamentos de performance verificados em CI (seção 4).
- Monitoramento de erros e uptime. Política de privacidade.
- Revisão final: `/hm-qa` e `/hm-deploy`.

**Critério de saída.** URL pública, Lighthouse 95 ou mais em todas as categorias, sete dias sem erro crítico.

### Fase 6 — Ponte para o app · M

**Objetivo.** Preparar o app sem reescrever nada.

**Entregas**
- Núcleo compartilhado isolado como pacote: modelo, dia da semana por fuso, validação, lógica de sync.
- Contrato de API documentado e versionado.
- Tokens de design exportáveis para o app.
- Decisão da tecnologia do app (nativo, multiplataforma ou PWA empacotado) com razão escrita.

**Critério de saída.** Um segundo cliente lê e escreve semanas usando apenas o contrato.

### Depois: o app

Fora deste roadmap. Entra quando a Fase 6 fechar.

## 4. Requisitos não funcionais

Valem desde a Fase 1. Não são fase de otimização. São restrição de design.

**Performance**

| Métrica | Orçamento |
|---|---|
| Abrir o app até ver a semana (cache quente) | < 100 ms |
| Abrir o app até ver a semana (frio, 4G) | < 1 s |
| Trocar de semana | < 100 ms |
| Latência de digitação | 0 (estado local, sem ida ao servidor) |
| Persistência após parar de digitar | < 1 s |
| JavaScript inicial | o mínimo que a stack escolhida permitir, medido em CI |

**Confiabilidade.** Nunca perder uma tecla: rascunho local sempre à frente do servidor, flush ao sair da página, fila offline. Leitura e escrita funcionam sem rede.

**Segurança.** Isolamento por usuário em cada consulta. Sessão segura. HTTPS obrigatório. Rate limiting. Sem rastreadores de terceiros. Backups com restauração testada. Afazeres são dados sensíveis: coletar apenas o e-mail.

**Acessibilidade.** Tudo operável por teclado, foco visível, contraste AA, reduced motion respeitado, semântica correta nos cartões.

**Qualidade.** Testes de matemática de datas e de autosave desde a Fase 1. Testes de sync na Fase 3. CI bloqueia merge com falha.

## 5. Decisões em aberto

Cada uma tem recomendação. Confirme ou mude antes da fase indicada. Sem resposta, a recomendação vale.

| # | Decisão | Recomendação | Por quê | Decidir até |
|---|---|---|---|---|
| 1 | Um usuário só ou contas para várias pessoas? | **Decidido (2026-09-05): contas.** Modelo de dados com dono desde a Fase 1. | O site é público, o app virá e o produto pode expandir. Ver D2. | — |
| 2 | Semana começa em segunda ou domingo? | **Decidido (2026-09-05): segunda.** A posição 0 é segunda e a 6 é domingo, no banco e no código. | Convenção brasileira e ISO. Ver D9. | — |
| 3 | Semana atual é criada ao abrir ou explicitamente? | Implícita para a semana atual (persistida na primeira tecla); explícita no hub para as demais | Zera o atrito da consulta diária, que é o fluxo principal. | Fase 1 |
| 4 | Texto livre ou lista com checkbox? | Texto livre no v1 | É o que foi pedido. Linhas iniciadas com `- ` podem virar checkbox depois sem migração. | Fase 1 |
| 5 | Tema: Miro é claro; padrão Higher Mind é dark-first | **Decidido (2026-09-05): dark-first** com a gramática do Miro; tema claro na Fase 4. | Mantém a referência sem virar cópia e respeita o padrão. Ver D1. | — |
| 6 | Layout do quadro | **Decidido (2026-09-05): faixa horizontal de sete cartões iguais**, deslizando para o lado. Substitui o 5 + 2. | Todos os dias são o mesmo espaço; a tela é fixa mas o ambiente é aberto. Ver D11. | — |
| 9 | A semana é ancorada em datas do calendário ou é um espaço sem datas? | **Decidido (2026-09-05): espaço sem datas**, com nome dado pelo usuário; "hoje" destaca o dia da semana atual. | É a dor descrita: não repetir o plano data a data. Ver D12. | — |
| 7 | Criar semanas no passado? | Sim, sem restrição | É um registro. O hub prioriza presente e futuro. | Fase 2 |
| 8 | Hub: barra lateral ou sobreposição? | Decidir no protótipo | Depende do layout do quadro. | Fase 2 |

## 6. Fora de escopo no v1

Colaboração e compartilhamento. Canvas infinito, pan e zoom. Arrastar texto entre dias. Lembretes e notificações. Recorrência. Tags, cores e prioridades. Anexos. Integração com calendários. Qualquer recurso de IA. Mais de uma semana na tela.

Nada disso está proibido para sempre. Está fora até o v1 provar a tese.

## 7. Riscos

| Risco | Mitigação |
|---|---|
| "Hoje" errado por fuso horário | Dia da semana calculado no fuso do usuário, em um único ponto do código, com teste. |
| Perda de texto no autosave (aba fechada, rede caiu) | Rascunho local sempre à frente, três gatilhos de salvamento, fila offline. |
| Conflito entre dispositivos | Último a escrever vence por dia com relógio do servidor. Domínio pequeno, resultado previsível. |
| Escopo crescer até virar gestor de tarefas | Invariantes e fora-de-escopo funcionam como contrato. Mudar exige mudar este documento. |
| Design cair em template | Direção definida na Fase 0, protótipo aprovado antes de codar, `/hm-designer` antes de shippar. |
| Retrabalho para o app | Modelo, datas e sync isolados desde a Fase 1; contrato formalizado na Fase 6. |

## 8. Próximo passo

Fase 0 construída (D1 a D16). Da Fase 2 já existem hub, seletor, criação de semana, URL por semana e "abrir onde parou"; da Fase 3, sessões e login com o Google. Para fechar a Fase 0: primeiro push com CI verde e aprovação visual. Em seguida, o núcleo da Fase 1: editor por dia com autosave. Acompanhamento em `docs/BACKLOG.md`.

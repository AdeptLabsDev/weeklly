# Weekly Planner — Roadmap

> Estado: planejamento (etapa 1). Atualizado em 2026-09-05.
> Próxima etapa: definição da stack (Fase 0).

## 0. Em uma frase

Um planejador semanal pessoal: sete dias em um quadro fixo, texto livre em cada dia, salvo sozinho, aberto em menos de um segundo. Site primeiro, app depois.

## 1. Tese e princípios

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
| **Semana** | Sete dias consecutivos, de segunda a domingo, ancorados em datas reais do calendário. Identificada pela data da segunda-feira. |
| **Dia** | Uma data dentro da semana e um bloco de texto livre. Tem seu próprio cartão (div) no quadro. |
| **Quadro** | A superfície fixa e finita onde os sete cartões vivem. Fundo estático com grid de pontos. Sem pan, sem zoom. |
| **Hub** | O navegador de semanas. Lista as semanas existentes, marca a atual, cria novas e troca a semana visível. |

### Invariantes (o sistema nunca quebra)

1. Uma semana tem exatamente sete dias, consecutivos, datados.
2. Não existem duas semanas para o mesmo intervalo de datas por usuário. A data de início é a identidade da semana.
3. Todo dia pertence a exatamente uma semana.
4. Toda edição é persistida sem ação explícita. Não existe botão "salvar".
5. Apenas uma semana é visível por vez.
6. O quadro tem tamanho fixo. O conteúdo de um dia rola dentro do seu cartão, nunca estoura o quadro.
7. "Hoje" e "semana atual" são calculados no fuso horário do usuário.

### Modelo de dados (conceitual, independente de stack)

```
Semana
  id
  dono            → usuário
  inicio          → data da segunda-feira (única por dono)
  dias[7]
    data
    texto
    atualizadoEm  → por dia, para sincronização
  criadaEm
  atualizadaEm
```

Por que `atualizadoEm` por dia e não só por semana: a sincronização entre dispositivos resolve conflitos no nível do dia (último a escrever vence). O domínio de conflito fica minúsculo e previsível.

### Fluxos principais

| Fluxo | Passos | Meta |
|---|---|---|
| Consultar | Abrir o app → semana atual aparece, hoje destacado | < 1 s frio, < 100 ms com cache |
| Editar | Clicar no dia → digitar → indicador "Salvo" | latência de digitação zero; persistido em < 1 s após parar |
| Criar semana | Hub → "Nova semana" → confirma a semana sugerida (a próxima sem plano) ou escolhe outra → quadro abre | 2 cliques |
| Trocar semana | Hub → clicar na semana, ou atalhos ← →, ou "Ir para hoje" | < 100 ms |
| Sair | Fechar a aba a qualquer momento | zero perda de texto |

### Endereçamento

Cada semana tem uma URL própria pela data de início, no formato `/semana/2026-09-07`. Isso dá deep link, histórico do navegador e um contrato simples para o app reutilizar.

### Esboço do quadro no desktop

Esboço da recomendação de layout (decisão 6), não o design. Fim de semana empilhado na sexta coluna.

```
┌──────────────────────────────────────────────────────────────┐
│  7 – 13 set 2026                                [Hub] [Hoje] │
│ ·  ·  ·  ·  ·  ·  ·  ·  ·  ·  ·  ·  ·  ·  ·  ·  ·  ·  ·  ·  │
│ ┌──────┐ ┌──────┐ ┌──────┐ ┌──────┐ ┌──────┐ ┌──────┐        │
│ │ SEG  │ │ TER  │ │ QUA  │ │ QUI  │ │ SEX  │ │ SÁB  │        │
│ │  7   │ │  8   │ │  9   │ │  10  │ │  11  │ │  12  │        │
│ │      │ │      │ │      │ │      │ │      │ ├──────┤        │
│ │      │ │      │ │      │ │      │ │      │ │ DOM  │        │
│ │      │ │      │ │      │ │      │ │      │ │  13  │        │
│ └──────┘ └──────┘ └──────┘ └──────┘ └──────┘ └──────┘        │
│ ·  ·  ·  ·  ·  ·  ·  ·  ·  ·  ·  ·  ·  ·  ·  ·  ·  ·  ·  ·  │
└──────────────────────────────────────────────────────────────┘
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
- Quadro fixo com sete cartões datados. Semana atual detectada pelo fuso do usuário. Hoje destacado.
- Editor de texto livre por dia, com o cartão rolando internamente.
- Autosave em três gatilhos: pausa na digitação, perda de foco, saída da página. Indicador discreto de estado (salvo, salvando).
- Persistência local. Não é descartável: vira a camada de cache da Fase 3.
- Atalhos básicos: navegar entre dias pelo teclado, foco visível.
- Testes da matemática de datas: virada de ano, semana 53, fusos, horário de verão.

**Critério de saída.** Você usa por sete dias seguidos sem perder uma tecla e sem pensar no app.

### Fase 2 — Hub e múltiplas semanas · M

**Objetivo.** Várias semanas com navegação que não parece carregar.

**Entregas**
- Hub: lista de semanas agrupada por mês e ano. Semana atual marcada. Semanas com conteúdo distinguíveis das vazias.
- Criar semana: sugere a próxima semana sem plano; permite escolher qualquer data, passado incluído.
- Trocar semana em menos de 100 ms, com as semanas vizinhas pré-carregadas.
- Atalhos ← → entre semanas e "Ir para hoje".
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
- Núcleo compartilhado isolado como pacote: modelo, matemática de datas, validação, lógica de sync.
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
| 1 | Um usuário só ou contas para várias pessoas? | Contas desde o início; modelo de dados com dono desde a Fase 1 | O site é público e o app virá. Adicionar dono depois é migração; ter desde o início é uma coluna. | Fase 0 |
| 2 | Semana começa em segunda ou domingo? | Segunda (ISO 8601, convenção BR) | Alinha com a matemática de semanas padrão. Configurável depois sem migração. | Fase 0 |
| 3 | Semana atual é criada ao abrir ou explicitamente? | Implícita para a semana atual (persistida na primeira tecla); explícita no hub para as demais | Zera o atrito da consulta diária, que é o fluxo principal. | Fase 1 |
| 4 | Texto livre ou lista com checkbox? | Texto livre no v1 | É o que foi pedido. Linhas iniciadas com `- ` podem virar checkbox depois sem migração. | Fase 1 |
| 5 | Tema: Miro é claro; padrão Higher Mind é dark-first | Dark-first com a gramática do Miro (grid de pontos, cartões, toolbar flutuante); tema claro na Fase 4 | Mantém a referência sem virar cópia e respeita o padrão. | Fase 0 |
| 6 | Layout desktop: 7 colunas ou 5 + 2 (fim de semana empilhado)? | 5 + 2 | Sete colunas ficam estreitas em notebook; fim de semana costuma ter menos texto. Validar no protótipo da Fase 0. | Fase 0 |
| 7 | Criar semanas no passado? | Sim, sem restrição | É um registro. O hub prioriza presente e futuro. | Fase 2 |
| 8 | Hub: barra lateral ou sobreposição? | Decidir no protótipo | Depende do layout do quadro. | Fase 2 |

## 6. Fora de escopo no v1

Colaboração e compartilhamento. Canvas infinito, pan e zoom. Arrastar texto entre dias. Lembretes e notificações. Recorrência. Tags, cores e prioridades. Anexos. Integração com calendários. Qualquer recurso de IA. Mais de uma semana na tela.

Nada disso está proibido para sempre. Está fora até o v1 provar a tese.

## 7. Riscos

| Risco | Mitigação |
|---|---|
| Matemática de datas errada (fuso, virada de ano, semana 53) | Testes desde a Fase 1, cálculo sempre no fuso do usuário, data de início como identidade. |
| Perda de texto no autosave (aba fechada, rede caiu) | Rascunho local sempre à frente, três gatilhos de salvamento, fila offline. |
| Conflito entre dispositivos | Último a escrever vence por dia com relógio do servidor. Domínio pequeno, resultado previsível. |
| Escopo crescer até virar gestor de tarefas | Invariantes e fora-de-escopo funcionam como contrato. Mudar exige mudar este documento. |
| Design cair em template | Direção definida na Fase 0, protótipo aprovado antes de codar, `/hm-designer` antes de shippar. |
| Retrabalho para o app | Modelo, datas e sync isolados desde a Fase 1; contrato formalizado na Fase 6. |

## 8. Próximo passo

Fase 0. Definir a stack com `/hm-init`, registrar cada escolha com razão em `docs/DECISIONS.md`, e fechar as decisões 1, 2, 5 e 6 acima.

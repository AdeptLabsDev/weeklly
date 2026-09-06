# SEO da Weeklly

Atualizado em 2026-09-06. Fontes externas acessadas em 2026-09-06.

## Produto, público e objetivo

A Weeklly é um planejador semanal pessoal: semanas com nome, sete dias de segunda a domingo, tarefas com horário opcional, conclusão e salvamento no ato. A semana não depende de datas; o usuário monta sua rotina e volta ao quadro para consultá-la. A entrada `/` deve continuar abrindo a última semana usada.

Contexto consultado: [CLAUDE.md](../CLAUDE.md), [ROADMAP.md](ROADMAP.md), [BACKLOG.md](BACKLOG.md), [ARCHITECTURE.md](../ARCHITECTURE.md) e handlers/templates atuais. O backlog distingue recursos entregues de intenções do roadmap; quando houver divergência, conferir o código antes de publicar uma promessa.

O público inicial proposto são pessoas organizando a própria rotina de trabalho, estudo e compromissos pessoais. O objetivo orgânico é ajudar essas pessoas a entender a Weeklly e começar uma semana. Trata-se de hipótese de posicionamento, ainda sem pesquisa de clientes ou dados de aquisição.

A comunicação pode mostrar sete dias, semanas reutilizáveis, tarefas, horários opcionais, conclusão e duplicação. Preço, exportação, funcionamento offline, PWA, sincronização offline, desempenho medido e depoimentos exigem comprovação própria antes de entrar em copy ou dados estruturados. O roadmap não é comprovação de disponibilidade.

## Skills pesquisadas e aplicação

A pesquisa partiu do catálogo skills.sh e conferiu o conteúdo original no GitHub. Foram selecionadas duas skills de `coreyhaines31/marketingskills`, de Corey Haines, para leitura e aplicação nesta rodada; nenhuma instalação permanente foi executada.

| Skill | Evidência consultada | Uso na Weeklly |
|---|---|---|
| `seo-audit` | [Catálogo](https://www.skills.sh/coreyhaines31/marketingskills/seo-audit), aproximadamente 201 mil instalações; [SKILL.md original](https://github.com/coreyhaines31/marketingskills/blob/main/skills/seo-audit/SKILL.md) | Priorizar rastreamento, indexação, páginas públicas, idiomas, metadados e evidências de validação. |
| `schema` | [Catálogo](https://www.skills.sh/coreyhaines31/marketingskills/schema), aproximadamente 56 mil instalações; [SKILL.md original](https://github.com/coreyhaines31/marketingskills/blob/main/skills/schema/SKILL.md) | Modelar JSON-LD que descreva o produto real e conferir sua correspondência com o HTML visível. |

O [repositório original](https://github.com/coreyhaines31/marketingskills) exibia aproximadamente 47 mil estrelas e licença MIT. Esses números são sinais de adoção, sujeitos a atualização; as skills são comunitárias e suas recomendações técnicas foram confrontadas com documentação oficial. `copywriting` e `frontend-design` já estão disponíveis localmente para texto e apresentação, sem necessidade de reinstalação.

Instalação opcional específica, conforme os comandos publicados nas páginas do catálogo:

```text
npx skills add https://github.com/coreyhaines31/marketingskills --skill seo-audit
npx skills add https://github.com/coreyhaines31/marketingskills --skill schema
```

Não é dependência do aplicativo nem etapa necessária do build. Uma futura instalação deve revisar a versão atual do arquivo; o pacote inteiro não é necessário para esta implementação.

## Auditoria inicial e escopo desta rodada

A fundação abaixo foi implementada e validada localmente nesta rodada. `task check` passou (vet, lint, testes, vulnerabilidades e build). A suíte Playwright completa terminou com 10 testes aprovados e 2 testes de interação exclusiva de desktop ignorados no mobile. Isso inclui quatro verificações de SEO, navegação sem JavaScript, criação de semana em inglês, largura de 320px e regressões dos fluxos existentes. As capturas PT/EN e os PNGs sociais foram inspecionados. Esta tabela não representa uma auditoria de domínio publicado.

| Prioridade | Evidência inicial no repositório | Entrega desta rodada |
|---|---|---|
| P0 | `/` atende o hub ou redireciona à última semana em `internal/server/handlers.go`. | Manter o fluxo; criar páginas públicas específicas que expliquem o produto. |
| P0 | `internal/server/server.go` não registra landing pages, robots ou sitemap. | Registrar duas páginas públicas, `/robots.txt` e `/sitemap.xml`. |
| P0 | Layout e middleware sem política explícita de `noindex`. | Aplicar `noindex` às rotas do aplicativo e desabilitar indexação pública por padrão. |
| P1 | `web/templates/layout.html` contém título, sem descrição, canonical, alternates ou metadados sociais. | Fornecer metadados completos nas páginas públicas, com origem configurada. |
| P1 | Idioma do aplicativo é uma preferência do usuário. | Fixar idioma pela URL pública, com conteúdo integral em PT-BR e inglês. |
| P1 | Sem apresentação pública do produto em HTML nem JSON-LD. | Renderizar conteúdo útil no servidor e grafo `WebPage` + `WebApplication`. |
| P1 | Sem imagem de compartilhamento para as páginas públicas. | Incluir duas imagens PNG localizadas, com URLs absolutas nos metadados. |
| P2 | Nenhum baseline de Search Console, Bing ou aquisição foi fornecido. | Preparar procedimento de lançamento e rotina de medição; coleta permanece pendente. |

## Mapa de páginas e intenção

As expressões abaixo são hipóteses de linguagem e intenção, escolhidas pela aderência ao produto. Não foram medidos volume, dificuldade, posição ou potencial de tráfego.

| URL pública | Idioma | Expressão principal proposta | Intenção e conteúdo |
|---|---|---|---|
| `/planejador-semanal` | `pt-BR` | planejador semanal online | Entender e experimentar uma ferramenta pessoal; explicar semana sem datas, sete dias, tarefas e consulta da rotina. |
| `/en/weekly-planner` | `en` | online weekly planner | A mesma intenção em inglês, com tradução completa e exemplos compreensíveis nesse idioma. |
| `/perguntas-frequentes` | `pt-BR` | perguntas sobre o planejador | Tirar dúvidas antes de começar: grátis, conta, datas, onde fica salvo, celular, idiomas. `FAQPage` no JSON-LD. |
| `/en/faq` | `en` | weekly planner questions | A mesma página em inglês. |

A landing tem título e descrição próprios, H1 com a expressão principal e os pontos fortes concretos (grátis, sem cadastro), "Como funciona" como três passos e uma cena que muda com o passo (rádios e CSS; com o script, a posição do bloco na tela durante a rolagem escolhe o passo, sem prender a página), seis pontos fortes e o rodapé institucional. Não há demonstração com tarefas de exemplo: decisão do Miguel, porque ela refazia o produto e passava outra impressão; a cena usa barras, não tarefas. As perguntas frequentes têm página própria, ligada no rodapé (D26). Links HTML para a versão equivalente (PT/EN na barra e no rodapé) e para o aplicativo. A barra tem o tema (cookie, formulário sem script) e o idioma (painel com links). O `app.js` é carregado como melhoria; nenhum conteúdo depende dele. As páginas públicas são uma lista em `seo.go` (`publicPages`): rota, hreflang, seletor de idioma e sitemap saem dela.

Criar outra URL só quando ela resolver uma necessidade diferente com conteúdo próprio. Variações como “planner semanal” e “organizador semanal” podem aparecer naturalmente na mesma página; não justificam cópias quase iguais.

## Contratos técnicos

1. **HTML público previsível:** as duas páginas devem responder `200` com conteúdo e metadados no HTML inicial, sem login, consulta a dados pessoais, criação de sessão ou dependência de JavaScript. Cookies e `Accept-Language` não podem trocar o idioma da URL. O único cookie lido é o do tema, que muda `data-theme` e o botão sol/lua, nunca o conteúdo.
2. **Origem única:** construir canonical, alternates, imagens e sitemap a partir de `WEEKLLY_BASE_URL`, jamais de um `Host` enviado pelo visitante. Cada idioma referencia sua própria URL como canonical. [Orientação Google sobre canonical](https://developers.google.com/search/docs/crawling-indexing/consolidate-duplicate-urls).
3. **Idiomas equivalentes:** incluir `pt-BR` e `en` em ambas as páginas, inclusive autorreferência, com URLs absolutas e links recíprocos. O fallback `x-default` deve apontar à página pública padrão definida pelo produto. [Orientação Google sobre versões localizadas](https://developers.google.com/search/docs/specialty/international/localized-versions).
4. **Aplicativo fora dos resultados:** hub, semanas, tarefas, autenticação e rotas operacionais recebem política explícita de `noindex`. O controle de acesso continua protegendo dados; `noindex` só orienta mecanismos de busca.
5. **Robots compatível com noindex:** permitir rastreamento para o robô conseguir ler a diretiva. Bloquear uma URL no robots pode impedir a leitura de `noindex` e ainda permitir que sua URL apareça nos resultados. [Orientação Google sobre bloqueio de indexação](https://developers.google.com/search/docs/crawling-indexing/block-indexing).
6. **Sitemap pequeno e correto:** listar somente as duas páginas canônicas públicas quando a indexação estiver habilitada. Não listar hub, IDs de semanas, endpoints ou URLs de preview. Referenciar o sitemap no robots. [Construção de sitemaps no Google](https://developers.google.com/search/docs/crawling-indexing/sitemaps/build-sitemap).
7. **Atualização verificável:** não inventar `lastmod` a cada requisição. Só acrescentá-lo quando houver uma data real de alteração relevante; omitir `priority` e `changefreq`, ignorados pelo Google. [Regras de sitemap XML](https://developers.google.com/search/docs/crawling-indexing/sitemaps/build-sitemap).
8. **JSON-LD fiel:** serializar `WebPage` e `WebApplication` com dados do catálogo e URLs configuradas. Nunca incluir `Offer`, preço, avaliações, estrelas ou perfis sociais sem informação pública comprovada. [Políticas de dados estruturados](https://developers.google.com/search/docs/appearance/structured-data/sd-policies).

O grafo semântico não promete um rich result de software. Esse recurso do Google exige, entre outros campos, oferta com preço e avaliação/review; esses dados não estão comprovados na Weeklly. Avisos sobre elegibilidade devem ser documentados, sem preenchimento fictício. [Requisitos de SoftwareApplication e WebApplication](https://developers.google.com/search/docs/appearance/structured-data/software-app).

## Ativação no domínio oficial

O domínio oficial ainda precisa ser confirmado. A configuração inicial é `WEEKLLY_SEARCH_INDEXING=false`. Desenvolvimentos locais e previews continuam nesse estado.

Para habilitar no ambiente público correto, configurar os três valores em conjunto:

```text
WEEKLLY_ENV=production
WEEKLLY_BASE_URL=https://dominio-oficial
WEEKLLY_SEARCH_INDEXING=true
```

`https://dominio-oficial` é um marcador documental: substituir pela origem HTTPS oficial confirmada, sem caminho, query ou fragmento. Confirmar também a variante de hostname e os redirecionamentos HTTP/HTTPS na hospedagem antes de enviar URLs aos buscadores.

O bootstrap termina quando o domínio real serve o build validado e os checks abaixo passam. Deploy, DNS, acesso às contas e submissões não foram realizados por este documento.

## Critérios de validação

Os testes de contrato estão em `internal/server/seo_test.go` e `internal/config/config_test.go`; os fluxos e capturas em `e2e/tests/seo.spec.ts`. Para repetir só o navegador: `node node_modules/@playwright/test/cli.js test --grep @seo`, dentro de `e2e/`, após `task css`. Os PNGs são regeneráveis com `node e2e/scripts/generate-social.mjs` na raiz, usando os catálogos, a marca e os tokens atuais.

- Rodar `task check` no código integrado, além dos testes de navegador adequados à mudança; registrar resultado real na entrega.
- Conferir HTTP e HTML das duas URLs com e sem cookies, com idiomas de navegador conflitantes e JavaScript desabilitado.
- Verificar título, descrição, H1, idioma, canonical, alternates recíprocos, JSON-LD válido e links internos funcionais.
- Conferir que o app ainda volta à última semana e que seus endpoints continuam com `noindex`, inclusive quando a indexação pública está ativa.
- Testar os estados com indexação desabilitada e habilitada; apenas as duas páginas públicas podem entrar no sitemap ativo.
- Garantir que URLs absolutas não mudam por `Host` arbitrário e que configuração inválida de produção falha de modo explícito.
- Verificar PNGs de compartilhamento, tipo de conteúdo, URL acessível, idioma e legibilidade; revisar a página em viewport móvel e desktop.
- No domínio publicado, validar o grafo no [Schema.org Validator](https://validator.schema.org/) e conferir elegibilidade no [Rich Results Test](https://search.google.com/test/rich-results), registrando as limitações de software acima.
- Medir desempenho público em [PageSpeed Insights](https://pagespeed.web.dev/) e acompanhar dados de campo quando disponíveis; distinguir teste de laboratório de uso real. [Core Web Vitals no Google](https://developers.google.com/search/docs/appearance/core-web-vitals).

## Próximas etapas priorizadas

| Ordem | Entrega | Critério para avançar |
|---|---|---|
| 1 | Publicar a fundação no domínio oficial e confirmar política de indexação. | Páginas e recursos acessíveis, HTTPS e metadados corretos, testes locais concluídos. |
| 2 | Verificar propriedade no Search Console, enviar `/sitemap.xml` e inspecionar as duas URLs. | Registrar estado de rastreamento, canonical escolhido e eventuais motivos de exclusão. |
| 3 | Adicionar/verificar o site no Bing Webmaster Tools e enviar o sitemap. | Registrar processamento e erros; não tratar envio como confirmação de indexação. |
| 4 | Observar consultas reais, conversar com primeiros usuários e revisar a intenção das páginas. | Decidir copy e pauta com evidência, mantendo histórico de alterações. |
| 5 | Publicar um guia original: como montar uma semana reutilizável sem datas. | Exemplo completo, autoria real, passos executáveis na Weeklly e link para experimentar. |
| 6 | Avaliar guias de rotina de estudo e de trabalho pessoal. | Necessidades distintas confirmadas; exemplos próprios, sem repetição mecânica entre páginas. |
| 7 | Conectar guias à página de produto e entre si quando houver relação útil. | Âncoras descritivas, nenhuma página órfã, tradução revisada antes de ampliar idiomas. |

Os passos de Search Console e Bing permanecem pendentes de domínio publicado e acesso autorizado às contas. Fontes: [envio ao Google](https://developers.google.com/search/docs/crawling-indexing/sitemaps/build-sitemap), [verificação no Bing](https://www.bing.com/webmasters/help/add-and-verify-site-12184f8b) e [sitemaps no Bing](https://www2.bing.com/webmasters/help/sitemaps-3b5cf6ed).

## Medição e rotina operacional

Baseline em 2026-09-06: indisponível. Não foram consultadas propriedades de Search Console/Bing, dados de conversão, posições, links externos ou métricas de campo da Weeklly.

| Cadência | Observar | Decisão apoiada |
|---|---|---|
| Após cada deploy relevante | Status HTTP, robots, sitemap, canonical, idiomas e páginas privadas. | Detectar regressões antes que se propaguem. |
| Semanal, após início da coleta | Páginas descobertas/indexadas, erros de rastreamento e canonical escolhido. | Corrigir impedimentos concretos de descoberta e indexação. |
| A cada 28 dias, quando houver dados | Impressões, cliques e CTR por página, consulta, país e dispositivo; separar marca de buscas genéricas. | Refinar texto e intenção; interpretar amostras pequenas com cuidado. |
| Mensal | Conteúdo entregue, links úteis, desempenho e dúvidas recebidas. | Escolher o próximo guia e atualizar instruções que mudaram. |

A métrica de produto proposta é a proporção de visitantes orgânicos que começam uma semana e adicionam a primeira tarefa. Sua instrumentação ainda está pendente; definir atribuição e agregação com o time, respeitando a regra de não carregar JavaScript de terceiros e sem coletar textos de tarefas. Métricas de busca e ativação têm denominadores diferentes e não devem ser confundidas.

## Busca com IA

Manter explicações claras, exemplos próprios, informações verificáveis e conteúdo rastreável em HTML. O Google informa que as boas práticas de SEO existentes também se aplicam a AI Overviews e AI Mode, sem otimização especial obrigatória. Não há motivo demonstrado nesta rodada para criar `llms.txt`, inventar estatísticas ou produzir páginas em massa. [AI features and your website](https://developers.google.com/search/docs/appearance/ai-features).

Novas iniciativas serão priorizadas pelos dados de uso e busca após o lançamento. Indexação, posições, tráfego e citações por IA permanecem resultados a observar; não são entregáveis garantidos por metadados, skills ou submissão de sitemap.

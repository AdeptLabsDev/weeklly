# Backlog

Só tópicos. `[x]` feito, `[ ]` não feito. Detalhes em [ROADMAP.md](ROADMAP.md) e [DECISIONS.md](DECISIONS.md).

## Planejamento

- [x] Roadmap do projeto
- [x] Decisão: tema dark-first
- [x] Decisão: contas multiusuário
- [x] Definição da stack: Go, Tailwind v4, JS nativo, SQLite
- [x] Backlog
- [x] Guia visual em design, prompt reutilizável e imagem social explicando a weeklly

## Fase 0 — Fundação

- [x] Repositório iniciado (`/hm-init`)
- [x] Go, Task, air, golangci-lint, govulncheck e Tailwind instalados
- [x] CI: vet, lint, testes, vulnerabilidades, build, imagem Docker
- [x] Tokens de design
- [x] Decisão: início da semana (segunda ou domingo)
- [x] Decisão: layout do quadro no desktop
- [x] Modelo de dados escrito
- [x] Tela do quadro com dados fictícios
- [x] Quadro refeito: faixa horizontal de cartões iguais
- [x] Decisão: semana com datas ou espaço sem datas
- [x] Modelo de dados refeito para semanas com nome
- [x] Skills de design instaladas e aplicadas
- [x] Faixa: roda, arrasto com inércia, teclado, paginador
- [x] Revisão visual: paleta, tipografia, copy
- [x] Barra de navegação: logo, seletor de semana, hub, nova semana, entrar
- [x] Playwright ponta a ponta em `e2e/`
- [x] Esquema de cor monocromático (Vercel) em dois temas
- [x] Botão sol/lua na barra
- [x] Troca de tema varrendo da esquerda para a direita, ícone girando
- [x] Recentes/A–Z com pílula deslizando e lista animada
- [x] Logo leva ao hub
- [x] Menu de configurações na barra
- [x] Idioma: português e inglês, catálogo com teste
- [x] Desfazer e refazer nas ações de tarefa (botões e atalhos)
- [x] Idioma em painel próprio ao lado do menu (pronto para mais idiomas)
- [x] Cursor próprio: seta e clique, opcional em "Tipo de mouse"
- [x] Cor de destaque à escolha da pessoa (cursor e textos principais acompanham)
- [ ] Outras configurações no menu
- [x] Nunito nos papéis de exibição, fonte do sistema no resto
- [x] Decisão: cor de destaque (cada pessoa escolhe a sua; sem cor por padrão)
- [ ] Primeiro push com CI verde
- [ ] Aprovação visual da tela

## Fase 1 — Núcleo: uma semana

- [x] Criar a primeira semana com nome
- [ ] Hoje pelo fuso do navegador
- [x] Tarefas por dia: adicionar, editar no lugar, concluir, excluir
- [x] Horário opcional por tarefa
- [x] Arrasto da faixa ignora campos e botões
- [x] Reordenar tarefas arrastando pela alça (e Alt+setas)
- [x] Mover tarefa para outro dia arrastando
- [x] Clicar no espaço vazio do dia começa uma tarefa
- [x] Campo de horário em texto com seletor leve
- [x] Concluir tarefa com animação (check e linha da esquerda para a direita)
- [x] Duplicar tarefa (ícone à direita do x, com desfazer)
- [ ] Ordenar tarefas do dia por horário (opção)
- [ ] Limpar tarefas concluídas do dia
- [ ] Reordenar por toque longo fora da alça
- [x] Salvamento no ato, com indicador "Salvo"
- [ ] Fila offline (persistência local)
- [x] Testes de tarefas: unidade, HTTP e navegador

## Fase 2 — Hub e múltiplas semanas

- [x] Hub: lista por uso recente
- [x] Criar semana pelo nome
- [x] Seletor de semanas na barra
- [x] Renomear semana
- [x] Duplicar semana com as tarefas
- [x] Ordem da lista: recentes ou A–Z, lembrada por usuário
- [ ] Trocar semana com pré-carregamento
- [x] Abrir o app na última semana usada
- [x] URL por semana
- [x] Excluir semana com confirmação
- [ ] Desfazer exclusão
- [x] Decisão: hub como página e seletor como menu na barra

## Fase 3 — Conta e sincronização

- [x] Sessões
- [x] Login com o Google (OpenID Connect)
- [ ] Credenciais do Google no Cloud Console (Miguel)
- [ ] Outras formas de entrar
- [x] Isolamento por usuário
- [ ] Limpeza de sessões vencidas e usuários anônimos abandonados
- [ ] Sincronização local e remoto
- [ ] Fila offline
- [ ] Estados: salvo, sincronizando, offline
- [ ] Migração dos dados locais para a conta
- [ ] Exportação JSON e Markdown
- [ ] Backup com Litestream
- [ ] Restauração de backup testada
- [ ] Passkeys
- [ ] Revisão `/hm-engineer`

## Fase 4 — Encantamento

- [ ] Motion: troca de semana, cartões, salvamento (tema e ordem já feitos)
- [ ] Mobile web
- [ ] Estados vazios
- [x] Página 404 com voz própria
- [ ] Onboarding
- [x] Tema claro
- [ ] Acessibilidade
- [ ] Revisão `/hm-designer`

## Fase 5 — Lançamento web

- [x] SEO: apresentação pública PT-BR e inglês em URLs estáveis
- [x] SEO: canonical, hreflang, metadados sociais, PNGs e dados estruturados
- [x] SEO: robots, sitemap público e noindex nas rotas do aplicativo
- [x] SEO: testes de idioma, indexação, isolamento e fluxo sem JavaScript
- [x] SEO: skills pesquisadas e plano de conteúdo/medição em docs/SEO.md
- [x] Landing revisada: barra com tema e idioma, headline concreta, passos, pontos fortes, FAQ por último, rodapé
- [x] Landing: idioma como botão com o atual e painel com todos os idiomas
- [x] Referência de landing: Duolingo seção a seção em docs/LANDING-REFERENCE.md
- [x] Landing definitiva, parte 1: barra sem atalho, "como funciona" como cena interativa (rádios + CSS, avanço automático), FAQ em página própria (`/perguntas-frequentes`, `/en/faq`)
- [x] Landing definitiva, parte 2: abertura em duas colunas (texto que convence, mascote à direita), botão principal na cor da marca com tremor no hover
- [x] Como funciona avança pela rolagem (posição do bloco na tela, sem prender a página; trilho pela posição); tremor do botão no modelo CSS-Tricks; FAQ centralizada com resposta descendo suave e título na cor da marca
- [x] Faixa dos aplicativos na landing (fundo na tonalidade da marca, selo do Google Play em cartoon, "em breve" até a URL existir)
- [ ] Publicar o app e preencher `playStoreURL` em `seo.go`
- [ ] Mascote em cartoon: gerar com os prompts de `design/brand/MASCOTE.md` e trocar a silhueta da abertura
- [ ] Landing: um botão só na barra e seções restantes (decisão do Miguel pendente)
- [x] Mascote: esquilo como ícone do site (SVG, PNG 32, iPhone 180)
- [x] Cor principal da marca `#727cf5` (ícone e token `--color-brand`)
- [x] Mascote na abertura da landing (silhueta; cartoon pendente)
- [ ] Mascote no rodapé
- [ ] Páginas do rodapé: Sobre a Adept Labs, Contato, Termos de uso, Política de privacidade
- [ ] Redes sociais: perfis e links no rodapé
- [ ] SEO: confirmar domínio e habilitar indexação na produção oficial
- [ ] SEO: verificar Search Console/Bing e enviar sitemap após publicação
- [ ] SEO: medir Core Web Vitals e registrar baseline de busca
- [ ] SEO: primeiro guia original de semana reutilizável
- [ ] Domínio e HTTPS
- [ ] Cabeçalhos de segurança e rate limiting
- [ ] PWA instalável
- [ ] Orçamentos de performance em CI
- [ ] Monitoramento de erros e uptime
- [ ] Política de privacidade
- [ ] Decisão: hospedagem final
- [ ] Revisão `/hm-qa` e `/hm-deploy`
- [ ] Lançamento

## Fase 6 — Ponte para o app

- [ ] Núcleo compartilhado como pacote
- [ ] API `/api/v1` documentada e versionada
- [ ] Tokens de design exportáveis
- [ ] Decisão: tecnologia do app

## Depois

- [ ] App mobile

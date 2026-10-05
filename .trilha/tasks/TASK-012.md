---
id: TASK-012
title: guia do agente na página + env/docs
status: done
spec: 002-dashboard-por-agente-agente-matarazzo-concierge-cartesia
milestone: m3
covers:
  - M5
depends_on:
  - TASK-011
acceptance:
  - prompt, primeira mensagem e payloads das duas tools copiáveis em /matarazzo
  - .env.example com MATARAZZO_AGENT_ID
checks:
  - trilha check
created: "2026-10-05T14:47:58Z"
updated: "2026-10-05T17:58:37Z"
---

### Objetivo
Guia de configuração do agente Matarazzo + env/docs.

### Arquivos
- `.env.example`: acrescentar
  `MATARAZZO_AGENT_ID=` (Agent ID do Guia Matarazzo na Cartesia).
  `.env` local idem (preenchido — ver TASK-008).
- NOVO `app/matarazzo/textos.go`: `PROMPT_MATARAZZO` (concierge turístico
  PT-BR: identidade "Gui, concierge da Cidade Matarazzo"; frases curtas;
  apresenta hotel/restaurantes/lojas/eventos COM a ferramenta de guia;
  reservas só via `criar_reserva_matarazzo`; Mata Città sem reserva
  online — explique a lista de espera; nunca inventa preço/horário;
  despedida oferecendo mais ajuda), `PRIMEIRA_MATARAZZO` (saudação
  convidando a perguntar sobre o complexo ou reservar) e os dois
  payloads JSON das client tools (`ESQUEMA_MT_GUIA`,
  `ESQUEMA_MT_RESERVA` — envelope Cartesia igual ao da Wave: type client,
  pre_tool_speech auto, execution_mode immediate, expects_response true,
  timeout 20; parameters idênticos às rotas da TASK-002).
- `/matarazzo` ganha card "Configurar o agente (já feito nesta POC)" com
  os blocos copiáveis (reusar `copiavel`/`.wh-copy`) e o agent_id real em
  `ui.Kbd` para consulta.
- `README.md`: seção "Agentes" — tabela agente × provedor × env × página;
  nota de que o painel separa por agente.

### Verificação
`trilha check`; `curl -s localhost:3210/matarazzo | grep -ci matarazzo` ≥ 5.

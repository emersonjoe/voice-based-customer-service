---
id: TASK-013
title: e2e concierge + regressão WaveHub
status: idea
spec: 002-dashboard-por-agente-agente-matarazzo-concierge-cartesia
milestone: m3
covers:
  - M6
depends_on:
  - TASK-009
  - TASK-010
  - TASK-011
acceptance:
  - três pedidos do roteiro via voz geram guia/reserva com protocolo
  - painel ?agente=matarazzo isolado e com badge
checks:
  - trilha check && trilha vendor --check
created: "2026-10-05T14:47:58Z"
---

### Objetivo
E2E do concierge + regressão WaveHub.

### Roteiro
1. `./dev.sh` (env com `CARTESIA_API_KEY` e `MATARAZZO_AGENT_ID`).
2. `/matarazzo` → Iniciar (Cartesia) → saudação do Gui →
   "o que tem no complexo?" (tool consultar_guia) →
   "quero reservar o Taraz sexta às 20h para 2, no nome de Marina" (tool
   criar_reserva → protocolo RS-) → "e o Mata Città?" (sem_reserva_online
   explicado) → Encerrar.
3. `/painel?agente=matarazzo`: só reservas, badge dourado; chips por
   agente corretos; detalhe do atendimento mostra local/data/horas.
4. Regressão: `/` com ElevenLabs e com Cartesia (agente WaveHub) — fluxos
   de sempre, painel geral com os dois agentes separados por badge.
5. `trilha check && trilha vendor --check`.

### Evidência a registrar
Agent ID do Guia Matarazzo (criado e publicado na Cartesia durante esta
spec): conferir/colar de `.env` (`MATARAZZO_AGENT_ID`). Token 200 em
`/api/voz/cartesia?agente=$MATARAZZO_AGENT_ID` antes do teste de voz.

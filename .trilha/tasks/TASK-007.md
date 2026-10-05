---
id: TASK-007
title: e2e dos dois provedores com roteiro de comparação
status: idea
spec: 001-provedor-de-voz-selecionavel-elevenlabs-cartesia
milestone: m3
covers:
  - R7
depends_on:
  - TASK-005
  - TASK-006
acceptance:
  - três fluxos nos dois provedores com as mesmas rotas
  - nenhuma chave no navegador (só access token curto)
checks:
  - trilha check && trilha vendor --check
created: "2026-10-05T13:01:58Z"
---

### Objetivo

Validar os DOIS provedores ponta a ponta e registrar o roteiro de comparação para a apresentação.

### Roteiro (executar e conferir)

Preparar: `./dev.sh` com `ELEVENLABS_API_KEY`, `CARTESIA_API_KEY` e os dois Agent IDs no `.env`; navegador em `http://localhost:3210`.

1. **Regressão ElevenLabs**: provedor ElevenLabs → Iniciar → saudação → "minha internet caiu e quero abrir um chamado" → feed com protocolo → `/painel` linha nova (origem `voz`) em ≤ 6 s → Encerrar.
2. **Troca de provedor**: selecionar Cartesia → campo de texto desabilita com aviso → Agent ID da Cartesia → Iniciar → status "Conectando — sessão assinada (Cartesia)…" → "Conectado — pode falar."
3. **Fluxos na Cartesia**:
   - chamado técnico (fala) → feed + painel;
   - segunda via com o CPF `168.995.350-09` → assistente fala valor/vencimento do JSON;
   - religue com o CPF `111.444.777-35` SEM confirmar pagamento → assistente explica a pendência (status `comprovacao_pendente`); confirmando → protocolo RG-…;
   - barge-in: interromper a fala → áudio para na hora.
4. **Sem chave**: comentar `CARTESIA_API_KEY` no `.env`, reiniciar `/api` → widget mostra a mensagem de indisponibilidade (não trava).
5. **Rate limit**: `for i in $(seq 1 35); do curl -s -o /dev/null -w "%{http_code} " localhost:3210/api/voz/cartesia; done` → últimos `429`.

### Registros

- Colar no relatório da tarefa (`trilha-spec evidence` ou corpo): protocolos gerados em cada provedor e o tempo até aparecer no painel.

### Verificação final

- `trilha check` verde; `trilha vendor --check` verde (nada novo).

### Critérios de aceite

- Os dois provedores completam os três fluxos chamando as MESMAS rotas.
- Nenhuma chave exposta no navegador (DevTools → Network: só access tokens de curta vida).
- Painel mostra ambos com origem `voz`.

---
id: TASK-002
title: "CSP: wss da Cartesia no connect-src"
status: done
spec: 001-provedor-de-voz-selecionavel-elevenlabs-cartesia
milestone: m1
covers:
  - R2
depends_on: []
acceptance:
  - "header CSP contém wss://api.cartesia.ai e nenhum host novo além dele"
checks:
  - "go build ./... && curl -sI localhost:3210/ | grep -i content-security-policy | grep -q 'wss://api.cartesia.ai'"
created: "2026-10-05T13:01:46Z"
updated: "2026-10-05T14:08:41Z"
---

### Objetivo

Permitir na CSP apenas o host do WebSocket da Cartesia.

### Arquivo

- `app/setup.go` (existe)

### Passos

1. Em `Setup`, no mapa `cfg.Security.CSPExtra`, acrescentar ao array de `"connect-src"` a entrada `"wss://api.cartesia.ai"` (mantendo as quatro entradas ElevenLabs existentes e a ordem).
2. Atualizar o comentário acima do mapa para citar a Cartesia.

Só `wss://`: o browser apenas ABRE WebSocket (`wss://api.cartesia.ai/v1/agents/websocket/...`). O POST que cunha o access token acontece no SERVIDOR (não passa da CSP).

### Verificação

1. `go build ./...`
2. Subir `./dev.sh` e conferir o header:
   `curl -sI http://localhost:3210/ | grep -i content-security-policy | grep -c "wss://api.cartesia.ai"` → `1`
3. `trilha check`

### Critérios de aceite

- Header contém `wss://api.cartesia.ai` em connect-src e NENHUM host novo além dele.
- `trilha check` verde.

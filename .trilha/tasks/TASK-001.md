---
id: TASK-001
title: env e docs do provedor Cartesia
status: done
spec: 001-provedor-de-voz-selecionavel-elevenlabs-cartesia
milestone: m1
covers:
  - R1
depends_on: []
acceptance:
  - grep -q CARTESIA_API_KEY .env.example
  - README explica a troca de provedor
checks:
  - grep -q CARTESIA_API_KEY .env.example && grep -qi cartesia README.md
created: "2026-10-05T13:01:46Z"
updated: "2026-10-05T14:08:41Z"
---

### Objetivo

Preparar variáveis de ambiente e documentação para o provedor Cartesia, sem tocar em código.

### Arquivos

- `.env.example` (existe) — acrescentar bloco
- `.env` (local, fora do git) — acrescentar chaves vazias
- `README.md` — tabela de variáveis + seção "Provedores de voz"

### Passos

1. Em `.env.example`, DEPOIS do bloco `ELEVENLABS_API_KEY`, acrescentar:

```
# Cartesia (provedor alternativo de voz). A chave fica SÓ no servidor:
# é ela que cunha os access tokens de curta vida que o widget usa.
CARTESIA_API_KEY=
# Agent ID do Managed Agent criado em play.cartesia.ai/agents
CARTESIA_AGENT_ID=
```

2. Em `.env` local, acrescentar as mesmas duas variáveis (vazias — o dono preenche).

3. No `README.md`:
   - Na tabela "Variáveis de ambiente", acrescentar as duas linhas (mesmo estilo das linhas ELEVENLABS).
   - Nova seção `## Provedores de voz` depois de "Configurar o agente de voz": explicar o seletor do widget (`Provedor: ElevenLabs | Cartesia`), que cada provedor guarda o próprio Agent ID no navegador, que a troca é só recarregar e clicar, e a limitação: entrada por texto só existe no ElevenLabs (na Cartesia, apenas voz).

### Critérios de aceite

- `grep -q "CARTESIA_API_KEY" .env.example` passa.
- README menciona Cartesia, as duas variáveis e como trocar de provedor.
- Nada de segredo novo no código.

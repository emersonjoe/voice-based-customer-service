---
id: 001-provedor-de-voz-selecionavel-elevenlabs-cartesia
title: "Provedor de voz selecionável: ElevenLabs + Cartesia"
status: approved
assets:
  - public/voice.js
  - app/page.go
  - app/setup.go
  - app/api/voz
  - app/agente/page.go
  - app/agente/textos.go
  - .env.example
  - README.md
trust_boundaries:
  - Nenhuma mudança nas rotas de ferramenta (app/api/tools) nem no comportamento ElevenLabs
  - Sem nova dependência JavaScript vendored e sem CDN
  - "Sem segredo no browser: chaves só no servidor via env"
controls:
  - csp-connect-src-so-apenas-wss-api-cartesia
  - access-token-expires-curto-sem-log
  - rate-limit-na-rota-nova
evidence:
  - trilha check
requirements:
  - id: R1
    text: variáveis de ambiente e documentação do provedor Cartesia (.env.example, .env, README)
  - id: R2
    text: "CSP permite apenas wss://api.cartesia.ai no connect-src"
  - id: R3
    text: rota GET /api/voz/cartesia cunha access token de curta vida com a chave do servidor
  - id: R4
    text: seletor de provedor no widget com Agent ID por provedor e regressão zero na ElevenLabs
  - id: R5
    text: provedor Cartesia no widget por WebSocket bruto (mic PCM 44,1 kHz, playback, barge-in, client tools)
  - id: R6
    text: guia Cartesia na página /agente com passo a passo do console e payloads copiáveis
  - id: R7
    text: validação e2e dos dois provedores com roteiro de comparação registrado
---

# Provedor de voz selecionável: ElevenLabs + Cartesia

## Contexto

Hoje o widget da página inicial fala com um agente da ElevenLabs (ConvAI):
o servidor assina a sessão (`GET /api/voz/assinada`), o SDK vendored conecta
por WebRTC e as client tools chamam as três rotas de `app/api/tools/`.
Queremos **a mesma experiência com a Cartesia** e um **seletor de provedor**
no widget, para comparar os dois lado a lado na apresentação ao cliente.

A Cartesia oferece **Managed Agents** (play.cartesia.ai/agents): Ink-2 (STT)
+ LLM à escolha + Sonic-3.6 (TTS), com turn-taking no servidor deles. O
browser conecta por **WebSocket** (`wss://api.cartesia.ai/v1/agents/websocket/{agent_id}`)
autenticado por **access token de curta vida** cunhado no servidor
(`POST https://api.cartesia.ai/access-token`, grant `agent`). Ferramentas
executam no cliente pela mesma mensagem `client_tool_call` /
`client_tool_result` — os mesmos nomes e esquemas das ferramentas atuais
funcionam lá sem mudança no backend.

## Objetivo

1. Seletor "Provedor: ElevenLabs | Cartesia" no widget, persistido no
   navegador, com Agent ID por provedor.
2. Rota servidor `GET /api/voz/cartesia` que cunha o access token com a
   `CARTESIA_API_KEY` (chave nunca vai ao browser).
3. Provedor Cartesia no widget: WebSocket bruto (sem SDK novo, sem vendor
   novo), captura de microfone em PCM 44,1 kHz base64, playback com fila e
   barge-in, client tools chamando as mesmas rotas de sempre.
4. Guia `/agente` com o passo a passo do console Cartesia e os payloads
   JSON das três client tools prontos para copiar.
5. Regressão zero no caminho ElevenLabs.

## Fora de escopo

- Trocar o provedor ElevenLabs, o SDK vendored ou as rotas de ferramenta.
- STT/TTS próprios, gravação de áudio, nova base de dados.
- Entrada por texto no provedor Cartesia (o WS de agentes desta versão não
   documenta mensagem de texto — o campo avisa e fica desabilitado).
- Hospedar CDN nova: o provedor Cartesia usa o protocolo WS documentado,
   sem dependência JavaScript adicional.

## Decisões

- **WebSocket bruto na Cartesia** em vez de SDK: o protocolo é JSON
   documentado (mensagens ≤ 32 KiB, áudio base64 em pedaços de 20–100 ms),
   o que evita vendor novo e mantém a CSP apertada. O worklet de captura é
   um blob (já coberto por `script-src blob:`).
- **Token no servidor**: mesmo desenho da ElevenLabs — o widget pede a
   sessão pronta (`disponivel` + credencial curta) e o segredo fica no
   servidor. Sem chave → 503 e aviso no widget; upstream recusado → 502.
- **Agent ID por provedor** no localStorage (`wh_agent_id_elevenlabs`,
   `wh_agent_id_cartesia`; migração da chave antiga `wh_agent_id`), e o
   provedor em `wh_provedor` (padrão `elevenlabs`).
- **Mensagens de ferramenta idênticas**: o Cartesia recebe o mesmo JSON que
   a ElevenLabs (`cliente`, `protocolo`, `mensagem`, …) e o `call_id` da
   sessão vira o `conversation_id` enviado às rotas.
- **Estados e UI preservados**: mesmos ids (`wh-voice`, `wh-orb`,
   `wh-status`, `wh-toggle`, `wh-feed`, `wh-text`, `wh-send`), mesma
   máquina de estados (`data-state` idle/conectando/ativo) e mesmo feed.

## Segurança

- `CARTESIA_API_KEY` só no servidor (env); access token com
  `expires_in` curto e sem log.
- CSP: acrescentar apenas `wss://api.cartesia.ai` ao `connect-src`.
- Rota nova rate-limited pelo `limite.Limitador` existente.

## Evidência

`trilha check` verde, `trilha vendor --check` verde (nada novo pinado) e
roteiro e2e dos dois provedores descrito na última tarefa.

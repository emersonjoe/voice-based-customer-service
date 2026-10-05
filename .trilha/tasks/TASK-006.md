---
id: TASK-006
title: guia Cartesia na página /agente
status: idea
spec: 001-provedor-de-voz-selecionavel-elevenlabs-cartesia
milestone: m3
covers:
  - R6
depends_on:
  - TASK-003
acceptance:
  - payloads das três client tools copiáveis com envelope Cartesia
  - passo a passo do console completo
checks:
  - "trilha check && curl -s localhost:3210/agente | grep -ci cartesia | grep -v '^0$'"
created: "2026-10-05T13:01:58Z"
---

### Objetivo

Seção "Provedor Cartesia" na página `/agente`: passo a passo do console da Cartesia + payloads JSON das três client tools prontos para copiar (reusar a maquinaria `.wh-copy`/`data-copy` existente).

### Arquivos

- `app/agente/textos.go` (existe) — novas constantes
- `app/agente/page.go` (existe) — nova seção

### Passo 1 — textos.go: payloads das client tools Cartesia

Criar `ESQUEMA_CARTESIA_CHAMADO`, `ESQUEMA_CARTESIA_SEGUNDA_VIA`, `ESQUEMA_CARTESIA_RELIGUE`. Mesmos `parameters` dos esquemas ElevenLabs existentes; o envelope muda para o formato Cartesia (`POST /v1/agents/tools`). Modelo (repetir o padrão para as outras duas, trocando name/description/parameters):

```go
const ESQUEMA_CARTESIA_CHAMADO = `{
  "type": "client",
  "name": "abrir_chamado_tecnico",
  "description": "Abre um chamado técnico para o cliente. Chame quando tiver o CPF e a descrição do problema — confirme o resumo com o cliente antes. O nome é opcional: com o CPF no cadastro, a ferramenta identifica o cliente sozinha.",
  "pre_tool_speech": "auto",
  "execution_mode": "immediate",
  "expects_response": true,
  "response_timeout_secs": 20,
  "parameters": {
    "type": "object",
    "properties": {
      "nome":       { "type": "string", "description": "Nome completo do cliente, apenas se ele informar" },
      "cpf":        { "type": "string", "description": "CPF do cliente, com ou sem pontuação" },
      "problema":   { "type": "string", "enum": ["sem_internet", "sem_sinal", "lentidao", "wifi", "tv", "outro"] },
      "descricao":  { "type": "string", "description": "Descrição do problema em uma ou duas frases" },
      "prioridade": { "type": "string", "enum": ["baixa", "media", "alta"] },
      "telefone":   { "type": "string" },
      "email":      { "type": "string" }
    },
    "required": ["cpf", "descricao"]
  }
}`
```

Segunda via: `name enviar_segunda_via_boleto`, description igual à ElevenLabs, parameters cpf (required) + email + canal enum ["email","whatsapp"]; mesmo envelope.
Religue: `name solicitar_religue_confirmacao`, parameters cpf + pagamento_confirmado boolean + forma_pagamento enum ["pix","cartao","boleto","dinheiro"] + data_pagamento + comprovante; required ["cpf","pagamento_confirmado"]; envelope com `expects_response: true`.

### Passo 2 — textos.go: bloco do prompt

Constante `GUIA_CARTESIA` (texto corrido, será renderizado em um `h.Pre` copiável) com o passo a passo do console:

1. Conta em play.cartesia.ai → API Keys → criar chave (`sk_car_...`) → colar em `CARTESIA_API_KEY` no `.env` (só no servidor).
2. Playground → Agents → Create agent — escolher o tipo **Standard Agent** ("Recommended": a Cartesia roda STT+LLM+TTS e o loop de conversa; o Client-Managed é outro desenho, fora de escopo). Configurar: nome "Wave"; Instructions = o mesmo prompt de sistema da ElevenLabs (página acima — os nomes das ferramentas são idênticos); Initial message = a mesma saudação; Language `pt`; Timezone `America/Sao_Paulo`; LLM padrão sugerido pelo console; Voice = catálogo filtrado em Portuguese (pt-BR) — ouvir e escolher; Audio input: noise suppression `auto`, keyterms `["WaveHub", "WaveTV", "religue", "Pix", "CPF"]`.
3. Tools → criar as três client tools com os payloads abaixo → anexar as três ao agente.
4. Copiar o `agent_id` do agente → colar no widget (provedor Cartesia) ou em `CARTESIA_AGENT_ID` no `.env`.
5. Testar: o Playground exercita a conversa; as client tools só disparam no widget (executam no seu navegador).
6. Versões: mudar a config cria nova versão; conversas em curso mantêm a antiga. Rotule com version_description.

### Passo 3 — page.go: seção

Entre `webhookAlternativo()` e `checklist()` (ou depois de `agentePrivado()`), chamar `guiaCartesia()` com:

- `ui.Card` "Provedor Cartesia — configuração" com `ui.CardDescription("Paralelo à ElevenLabs: mesmo prompt, mesmas ferramentas")`
- `copiavel("guia-cartesia", GUIA_CARTESIA)` + botão copiar (`.wh-copy` + `data-copy-target="guia-cartesia"`, como os outros)
- Três cards (grid) com cada `ESQUEMA_CARTESIA_*` copiável, ids `esquema-cartesia-chamado`, `esquema-cartesia-segunda-via`, `esquema-cartesia-religue`, cada um com botão "Copiar JSON da tool"
- Nota curta: "A chave (`sk_car_...`) fica só no servidor: o widget recebe um access token de 2 minutos cunhado por `/api/voz/cartesia`."

### Verificação

1. `trilha check`
2. `curl -s localhost:3210/agente | grep -ci cartesia` → ≥ 3
3. Browser em `/agente`: botões "Copiar" das tools Cartesia funcionam (mesmo mecanismo dos demais).

### Critérios de aceite

- Payloads idênticos em nomes/parameters aos da ElevenLabs (só o envelope Cartesia muda).
- Nenhuma chave de exemplo que pareça real (usar `sk_car_...` literal).

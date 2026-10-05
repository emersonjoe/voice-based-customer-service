---
name: voice-based-customer-service
description: POC de atendimento por voz com IA (WaveHub) em Trilha/Go, com agente ElevenLabs, três ferramentas de chamado/segunda via/religue e painel em tempo real.
default_agent: coder
verify:
  - trilha check
milestones:
  - id: m1
    title: "Servidor: env, CSP e rota de token"
  - id: m2
    title: Widget multi-provedor
  - id: m3
    title: Guia Cartesia e validação dupla
---

# voice-based-customer-service

Prova de conceito de atendimento por voz com IA para a WaveHub, feita com o
framework Go [Trilha](https://github.com/emersonjoe/trilha) (zero dependências
fora da biblioteca padrão) e identidade visual da WaveHub (roxo→fúcsia sobre
quase preto).

## O que existe hoje

- **Agente de voz ElevenLabs** (ConvAI, client tools no browser + sessão
  assinada no servidor para agente privado) — `public/voice.js` +
  `app/api/voz/assinada/route.go`.
- **Três ferramentas** chamadas pelo agente, em `app/api/tools/`:
  `abrir_chamado_tecnico`, `enviar_segunda_via_boleto`,
  `solicitar_religue_confirmacao`. Validam CPF (com reparo para voz),
  registram atendimentos no store JSON (`internal/atendimento`) e respondem
  JSON pensado para ser falado.
- **Painel** `/painel` (ui.Shell + DataTable + Poll 6s) e dossiê por atendimento.
- **Guia do agente** `/agente` com prompt e esquemas copiáveis.
- SDK `@elevenlabs/client` vendored em `public/vendor` (sha256 no
  `vendor.lock`); CSP mínima com hosts da ElevenLabs; microfone `(self)`.

## Comandos

- build: `go build ./...`
- test: `trilha check` (gen, gofmt, vet, test, audit)
- dev: `./dev.sh` (carrega `.env` e sobe na :3210)
- vendor: `trilha vendor --check`

## Onde ficam as coisas

- `app/` — rotas por arquivos (trilha). `page.go` é `/`, `agente/` é o guia,
  `painel/` o painel, `api/` as rotas JSON (tools, chamados, voz).
- `internal/atendimento` — modelo, store JSON em disco, regras de CPF.
- `internal/limite` — rate limit por IP (janela fixa, memória).
- `internal/apiutil` — corpo JSON tolerante, rate limit, CPF da ferramenta.
- `public/` — `style.css` (tema WaveHub), `voice.js`, `copy.js`, `vendor/`.

## Segredo

Nada de segredo no código: `.env` (fora do git) carrega `TRILHA_SECRET`,
`ELEVENLABS_API_KEY`, `ELEVENLABS_AGENT_ID`. Chaves nunca vão ao navegador.

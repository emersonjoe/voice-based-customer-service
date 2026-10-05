---
id: 002-dashboard-por-agente-agente-matarazzo-concierge-cartesia
title: Dashboard por agente + Agente Matarazzo (concierge Cartesia)
status: approved
assets:
  - internal/atendimento/atendimento.go
  - internal/matarazzo
  - app/api/tools/matarazzo
  - app/painel/page.go
  - app/matarazzo
  - app/widget.go
  - public/voice.js
  - public/style.css
  - .env.example
  - README.md
trust_boundaries:
  - Regressão zero no agente WaveHub e nos fluxos existentes
  - "Reservas são simuladas: nenhuma integração real com o Rosewood"
  - Matarazzo roda apenas no provedor Cartesia
controls:
  - validacao-e-rate-limit-nas-rotas-novas
  - paleta-matarazzo-sem-imagens-externas
evidence:
  - trilha check
requirements:
  - id: M1
    text: campo agente no atendimento, tipo reserva e dataset do guia Matarazzo
  - id: M2
    text: rotas /api/tools/matarazzo/consultar_guia e criar_reserva com validação e rate limit
  - id: M3
    text: painel separando atendimentos por agente (filtro, chips, coluna, badge)
  - id: M4
    text: página /matarazzo com identidade Matarazzo e widget de voz Cartesia fixo
  - id: M5
    text: guia de configuração do agente na página e env/docs (MATARAZZO_AGENT_ID)
  - id: M6
    text: e2e do concierge + regressão WaveHub com evidências
---

# Dashboard por agente + Agente Matarazzo (concierge Cartesia)

## Contexto

A POC hoje tem um agente (WaveHub, provedores ElevenLabs e Cartesia) e um
painel único. Queremos: (1) o painel separando atendimentos POR AGENTE, e
(2) um segundo agente — o **Guia Matarazzo**, concierge de voz no provedor
Cartesia que apresenta o Complexo Cidade Matarazzo (hotel Rosewood São
Paulo, restaurantes, lojas, eventos) a turistas e registra reservas.

Fatos do complexo (pesquisa out/2026, base do guia):
- Complexo na Bela Vista (Alameda Rio Claro, 260), antigo Hospital
  Matarazzo (1904), ~30.000 m² de edifícios tombados restaurados.
- Hotel **Rosewood São Paulo**: 1º Rosewood da América do Sul (jan/2022),
  ~160 suítes (duplex e cobertura de 4 quartos), torre Mata Atlântica de
  Jean Nouvel, interiores de Philippe Starck, ~450 obras de arte brasileira.
- Restaurantes do hotel: **Le Jardin** (grand café 24h, opções kosher e
  jardim), **Blaise** (brasserie francesa com ingredientes brasileiros,
  chef Fernando Bouzan), **Taraz** (sul-americano, Guia Michelin),
  **Rabo di Galo** (bar de jazz), Bela Vista Bar, In Room/Private Dining.
- **Mata Città**: espaço italiano de 1.600 m², 7 conceitos, ~500 lugares,
  estética do cinema italiano anos 60/70; NÃO aceita reserva online
  (lista de espera no local).
- Lojas: butiques de grife em edifícios históricos. Eventos: espaços para
  casamentos e corporativo; Festival Movimento Cidade (cultura).

Identidade (extraída do site cidadematarazzo.com.br, sem brandbook
público): dourado #CEB678, creme #E8E1D7, preto #222222, branco e
tipografia serifada (Georgia nos estilos do site).

## Objetivo

1. Atendimento ganha campo `agente` (wavehub|matarazzo); painel filtra,
   conta e rotula por agente.
2. Página `/matarazzo` com a identidade acima e widget de voz próprio
   (Cartesia, `MATARAZZO_AGENT_ID`).
3. Ferramentas do concierge chamando rotas novas deste app:
   `consultar_guia_matarazzo` (tópicos do guia) e `criar_reserva_matarazzo`
   (hotel Rosewood ou restaurantes; protocolo RS-AAAA-NNNN).
4. Agente configurado e publicado no console Cartesia (FEITO nesta spec,
   ver evidência) — a implementação consome o agent_id do .env.
5. Guia de configuração e roteiro de demo na própria página.

## Fora de escopo

- Pagamento/cancelamento de reservas, integração real com o Rosewood,
  e-mails de confirmação.
- Mudança nos fluxos WaveLab (regressão zero no agente WaveHub).
- Segundo provedor para o Matarazzo (Cartesia apenas).

## Decisões

- **Agente como campo simples** (`string`, default `wavehub`) no
  atendimento — sem tabela de agentes; a lista válida é
  `atendimento.Agentes` (wavehub|matarazzo).
- **Reserva é um atendimento** de tipo novo (`reserva`, prefixo RS),
  prioridade baixa, com detalhes (tipo, data, horário, pessoas, nome).
- **Página /matarazzo reaproveita o widget**: mesmo markup/ids e mesmo
  voice.js, com `data-provedor-fixado="cartesia"` escondendo o seletor.
- **Guia como mapa em código** (`internal/matarazzo/guia.go`): conteúdo
  curado e versionado no repo; sem banco.

## Segurança

- Mesmo desenho das rotas atuais: validação de entrada, rate limit,
  sem PII em log, dados fictícios de reservas.
- `MATARAZZO_AGENT_ID` é identificador público (igual aos outros).

## Evidência

`trilha check`; agente Cartesia publicado (agent id registrado no
corpo da TASK-008) e token cunhado com 200 via /api/voz/cartesia.

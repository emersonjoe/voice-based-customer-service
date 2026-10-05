---
id: TASK-008
title: modelo multi-agente + guia Matarazzo em código
status: done
spec: 002-dashboard-por-agente-agente-matarazzo-concierge-cartesia
milestone: m1
covers:
  - M1
depends_on: []
acceptance:
  - Criar com agente matarazzo grava o campo; sem agente fica wavehub
  - protocolo de reserva usa prefixo RS
checks:
  - go test ./... && trilha check
created: "2026-10-05T14:47:58Z"
updated: "2026-10-05T17:44:15Z"
---

### Objetivo
Campo `agente` no atendimento + tipo `reserva` + dataset do guia Matarazzo.

### Arquivos
- `internal/atendimento/atendimento.go` (existe): acrescentar
  - `Agente string `json:"agente"`` no struct Atendimento
  - constantes `AgenteWavehub = "wavehub"`, `AgenteMatarazzo = "matarazzo"`;
    `var Agentes = []string{AgenteWavehub, AgenteMatarazzo}`; `AgentesValido(a)`.
  - `TipoReserva Tipo = "reserva"` com Prefixo `RS` e Rotulo `Reserva`; incluir em `Validos`.
  - `Criar(...)` ganha parâmetro `agente string` (vazio → AgenteWavehub). Atualizar TODOS os chamadores (rotas tools, chamados, seeds, testes).
- NOVO `internal/matarazzo/guia.go`: pacote `matarazzo` com
  `type Topico string` e `var Guia = map[Topico]Entrada{...}` onde
  `Entrada{Titulo, Texto string; Dicas []string}`. Tópicos (texto curto,
  factual, com os dados da spec): `complexo`, `hotel_rosewood`,
  `restaurantes`, `mata_citta`, `lojas`, `eventos`, `como_chegar`.
  - hotel_rosewood: ~160 suítes, torre Jean Nouvel + interiores Starck,
    ~450 obras; reservas pelo concierge (nossa tool) ou site oficial.
  - restaurantes: Le Jardin (24h, kosher, jardim), Blaise (francesa,
    chef Fernando Bouzan), Taraz (sul-americano, Michelin), Rabo di Galo
    (jazz). Mata_citta à parte (sem reserva online — lista de espera).
  - como_chegar: Alameda Rio Claro 260, Bela Vista, próximo à Av. Paulista.
  - `var RestaurantesReservaveis = []string{"hotel_rosewood","le_jardin","blaise","taraz","rabo_di_galo"}` com `Rotulo(id)` para nomes bonitos.

### Testes
Estender `atendimento_test.go`: protocolo de reserva usa prefixo RS;
`Criar` com agente matarazzo grava o campo; sem agente → wavehub.

### Verificação
`go test ./... && trilha check`

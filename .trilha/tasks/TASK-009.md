---
id: TASK-009
title: rotas consultar_guia e criar_reserva
status: done
spec: 002-dashboard-por-agente-agente-matarazzo-concierge-cartesia
milestone: m1
covers:
  - M2
depends_on:
  - TASK-008
acceptance:
  - consultar_guia devolve conteúdo por tópico com tópicos_disponiveis no erro
  - criar_reserva valida tipo/data/horário/pessoas e devolve RS-
checks:
  - trilha gen && trilha check
created: "2026-10-05T14:47:58Z"
updated: "2026-10-05T17:44:29Z"
---

### Objetivo
As duas ferramentas do concierge, seguindo o padrão das rotas existentes
(ler antes `app/api/tools/enviar_segunda_via_boleto/route.go`).

### Arquivos novos
- `app/api/tools/matarazzo/consultar_guia/route.go` — pacote `consultar_guia`, `POST`:
  corpo `{topico}` (tolerante via `apiutil.Corpo`); valida contra os
  tópicos de `matarazzo.Guia` (campo `TopicoValido(t)`); 200 com
  `{status:"sucesso", topico, titulo, texto, dicas[], mensagem}` — a
  mensagem é o resumo falável. Tópico inválido → 200 com
  `{status:"topico_invalido", mensagem, topicos_disponiveis[]}`.
  Rate limit balde `guia_matarazzo`. Origem igual às outras (voz/webhook).
- `app/api/tools/matarazzo/criar_reserva/route.go` — pacote `criar_reserva`, `POST`:
  corpo `{tipo, data (YYYY-MM-DD), horario (HH:MM), pessoas (1..12),
  nome_hospede (3..120), telefone?, observacoes? (<=300)}`.
  - `tipo` em `matarazzo.RestaurantesReservaveis`; data não no passado
    (compara com `time.Now()` no fuso local, aceita até 18 meses à frente);
    horario 06:00–23:59.
  - mata_citta NÃO é reservável: se `tipo=mata_citta`, 200 com
    `{status:"sem_reserva_online", mensagem: "O Mata Città não aceita reserva online — o guia explica: chegar cedo ou deixar o nome na lista de espera no local."}`.
  - Sucesso: `store.Criar(atendimento.TipoReserva, "", nome, resumo
    "Reserva: <Nome do local> — DD/MM HH:MM (N pessoas)", descricao com
    observacoes, PrioridadeBaixa, origem, conversation_id, detalhes{tipo,
    local, data, horario, pessoas, telefone, observacoes}, AgenteMatarazzo)`
    → 200 `{status:"sucesso", protocolo RS-…, mensagem}` com o nome do
    local e o protocolo.
  - Falhas de validação → 200 com status explicativo (`tipo_invalido`,
    `data_invalida`, `horario_invalido`, `pessoas_invalido`,
    `nome_invalido`), mensagem pronta para falar.

### Verificação
`trilha gen && trilha check`; com `./dev.sh`:
`curl -s localhost:3210/api/tools/matarazzo/consultar_guia -d '{"topico":"hotel_rosewood"}' -H 'Content-Type: application/json'`
→ sucesso; e criar_reserva válido devolve RS-…

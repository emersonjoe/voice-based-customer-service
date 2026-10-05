---
id: TASK-010
title: painel separando por agente
status: done
spec: 002-dashboard-por-agente-agente-matarazzo-concierge-cartesia
milestone: m2
covers:
  - M3
depends_on:
  - TASK-008
acceptance:
  - filtro ?agente=, chips por agente e coluna com badge
  - stats respeitam o filtro
checks:
  - trilha check
created: "2026-10-05T14:47:58Z"
updated: "2026-10-05T17:44:29Z"
---

### Objetivo
Painel separando por agente (dashboard multi-agente).

### Arquivos
- `internal/atendimento/atendimento.go`: `Filtro` ganha `Agente string`;
  `Listar` filtra por ele quando válido; `ContagemPorStatus(agente)` e
  `ResolvidosHoje(agora, agente)` ganham o parâmetro (vazio = todos);
  `Ultimos(n)` mantém. Acrescentar `PorAgente() map[string]int`.
- `app/painel/page.go`:
  - struct `filtro` ganha `Agente string `form:"agente"``;
  - faceta nova: select `agente` (Todos os agentes / WaveHub / Matarazzo)
    ao lado das existentes;
  - `resumo(...)` passa a receber o agente do filtro e desenha, além dos
    4 stats atuais, uma linha de chips por agente
    (`ui.Badge` + contagem: "WaveHub · N" com classe `wh-agente-wavehub`,
    "Matarazzo · N" com `wh-agente-matarazzo`), clicável (link com
    `?agente=...`);
  - coluna nova na tabela: `{Key:"agente", Label:"Agente"}` com badge por
    agente (rotulado "WaveHub"/"Matarazzo"; vazio mostra "WaveHub").
- `public/style.css`: `.wh-agente-wavehub` (gradiente roxo→fúcsia atual)
  e `.wh-agente-matarazzo` (dourado #CEB678 sobre #222) nos badges.
- Sidebar (`app/painel/layout.go`): item novo no grupo Atendimento —
  `{Href:"/matarazzo", Label:"Matarazzo", Icon:"house"}` (ou ícone
  existente do kit: `settings`/`calendar` — conferir `trilha ui icons`).

### Verificação
`trilha check`; browser: /painel mostra chips, coluna e filtro; ao filtrar
`agente=matarazzo` com reservas criadas, só elas aparecem; sem filtro,
tudo junto com o badge certo em cada linha.

---
id: TASK-011
title: página /matarazzo com brandbook e widget Cartesia
status: done
spec: 002-dashboard-por-agente-agente-matarazzo-concierge-cartesia
milestone: m2
covers:
  - M4
depends_on:
  - TASK-008
acceptance:
  - tema dourado/creme/preto serifado sem imagens externas
  - widget com provedor fixo cartesia e MATARAZZO_AGENT_ID; / continua igual
checks:
  - trilha gen && trilha check
created: "2026-10-05T14:47:58Z"
updated: "2026-10-05T17:58:37Z"
---

### Objetivo
Página `/matarazzo` com identidade Matarazzo (paleta do site: dourado
#CEB678, creme #E8E1D7, preto #222222, serifas) e widget de voz próprio.

### Arquivos
- Extrair o widget para NOVO `app/widget.go` (package app): `func
  WidgetVoz(ids WidgetIDs) h.Node` parametrizado — campos
  `RaizID`, `ProvedorFixo string` (vazio = seletor visível),
  `AgentePadrao`, `AgenteCartesiaPadrao`, `Titulo`, `Descricao`.
  `app/page.go` passa a chamar `WidgetVoz(...)` (markup igual ao atual);
  a página Matarazzo chama com `ProvedorFixo:"cartesia"`.
- NOVO `app/matarazzo/page.go` (GET /matarazzo): dentro do layout raiz,
  seção própria `h.Section(h.Class("mt-page"))`:
  - hero: chip "Prova de conceito · Concierge por voz", título serifado
    "O concierge da **Cidade Matarazzo**", lead curto (guia + reservas),
    CTAs "Falar com o guia" (âncora #wh-voice) e "Ver o painel".
  - grid: coluna esquerda com cards "O que dá para pedir" (hotel
    Rosewood, restaurantes, Mata Città, lojas, eventos, como chegar —
    texto de `matarazzo.Guia`, títulos via `Entrada.Titulo`) e o widget
    de voz à direita (Cartesia fixo, `MATARAZZO_AGENT_ID` do env).
  - card "Roteiro da demonstração": peça "quero jantar no Taraz sexta às
    20h para 2 pessoas" e "o que tem no complexo?" — anote o protocolo RS.
- `public/style.css`: bloco `.mt-*` — `.mt-page` fundo #222 com detalhes
  creme; `.mt-hero .ui-h1` e `.mt-titulo` em Georgia/serif, cor creme;
  `.mt-chip` borda dourada; `.mt-card` fundo #2a2a28, borda #CEB6783d;
  botão primário da página com fundo #CEB678 e texto #222; badges
  dourados. Sem imagens externas (CSP img-src 'self' data:).
- Menu do layout raiz: `ui.NavLink("/matarazzo", "Matarazzo", path==...)`
  no header público.
- voice.js: se `root.dataset.provedorFixado` existir, esconder o campo do
  select (`.wh-field-wh-provedor` → `hidden`) e travar `provedorNome` no
  valor fixado; manter todo o resto.

### Verificação
`trilha gen && trilha check`; browser: /matarazzo com tema preto/dourado,
seletor de provedor oculto, widget iniciando sessão Cartesia com o
MATARAZZO_AGENT_ID; / continua idêntico (regressão).

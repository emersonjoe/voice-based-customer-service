---
id: TASK-004
title: seletor de provedor no widget + refatoração em provedores
status: ready
spec: 001-provedor-de-voz-selecionavel-elevenlabs-cartesia
milestone: m2
covers:
  - R4
depends_on: []
acceptance:
  - todos os ids de DOM preservados; ElevenLabs idêntico ao comportamento atual
  - provedor e Agent ID persistem no localStorage com migração da chave antiga
checks:
  - "trilha check && curl -s localhost:3210/ | grep -q wh-provedor"
created: "2026-10-05T13:01:58Z"
updated: "2026-10-05T13:03:17Z"
---

### Objetivo

Seletor de provedor no widget + refatoração de `voice.js` em provedores, SEM mudar o comportamento ElevenLabs (regressão zero). Preservar TODOS os ids de DOM existentes (o CSS e o JS dependem deles).

### Arquivos

- `app/page.go` (existe) — acrescentar o select
- `public/voice.js` (existe) — refatorar
- `public/provedores/elevenlabs.js` (NOVO) — o que hoje está em voice.js
- `public/provedores/cartesia.js` (NOVO, nesta tarefa só o esqueleto; corpo na tarefa seguinte)
- `public/style.css` (existe) — pequeno ajuste se necessário

### Passo 1 — markup (`app/page.go`)

Dentro do cartão do widget, ANTES do campo "Agent ID", acrescentar um campo com o select (usar `ui.Field` + `ui.Select`, mesmo estilo das facetas do painel):

- `ui.Field("wh-provedor", "Provedor de voz", <select>)`
- select: `h.ID("wh-provedor")`, `h.Name("provedor")`, `h.Aria("label", "Provedor de voz")`
- opções: `h.Option(h.Value("elevenlabs"), h.Text("ElevenLabs"))` e `h.Option(h.Value("cartesia"), h.Text("Cartesia"))`
- `ui.Help("Os dois ficam prontos para a demonstração — troque e recarregue para comparar.")`

### Passo 2 — contrato de provedor

Cada provedor é um módulo `public/provedores/<nome>.js` que exporta `criar<X>(ctx)` devolvendo `{ iniciar, encerrar, enviarTexto }`:

```js
ctx = {
  agenteId,                       // string
  cena(estado, texto),            // idle | conectando | ativo + frase do status
  anotar(texto),                  // feed de ferramentas (li no #wh-feed)
  volume(nivel0a1),               // animação das barras do orbe
  onEncerrado(),                  // chamado quando a sessão termina sozinha
}
// iniciar(): Promise<void> — resolve conectado; lança Error com .message legível
// encerrar(): Promise<void> — encerra sessão, devolve microfone
// enviarTexto(texto): Promise<{ok:boolean, motivo?:string}>
```

### Passo 3 — extrair ElevenLabs para `public/provedores/elevenlabs.js`

Mover de `voice.js` para o novo módulo, quase sem mudança:
- o `import "/vendor/elevenlabs-client.js"` (fica NO TOPO deste módulo, não em voice.js)
- `iniciar()`: pedido de sessão assinada em `GET /api/voz/assinada?agente=...` (503 → modo público silencioso; outro erro → aviso), `Conversation.startSession` com `clientTools` e callbacks (`onConnect/onDisconnect/onError/onModeChange`) chamando `ctx.cena`
- `ferramentas()`: o mapa `nome → POST /api/tools/<nome>` que hoje existe (manter o `anotar` com `ctx.anotar`)
- `encerrar()` e `enviarTexto()` como hoje
- EXPORTAR `criarElevenLabs(ctx)`; nada de estado global além do `convo` interno

### Passo 4 — `public/provedores/cartesia.js` (esqueleto nesta tarefa)

```js
export function criarCartesia(ctx) {
  return {
    async iniciar() { ctx.cena("conectando", "Provedor Cartesia ainda não implementado."); throw new Error("Cartesia: implementação na próxima tarefa."); },
    async encerrar() {},
    async enviarTexto() { return { ok: false, motivo: "Cartesia: apenas voz." }; },
  };
}
```

### Passo 5 — `voice.js` vira o maestro

- Remove o import do SDK; importa `criarElevenLabs` e `criarCartesia` (`import { criarElevenLabs } from "/provedores/elevenlabs.js"`) — caminhos absolutos a partir da raiz do site funcionam por a pasta `public/` ser servida em `/`.
- Estado: `provedorAtual` (`"elevenlabs" | "cartesia"`), lido de `localStorage.wh_provedor` (padrão `elevenlabs`).
- Agent ID por provedor: `wh_agent_id_elevenlabs` e `wh_agent_id_cartesia`. MIGRAÇÃO: se `localStorage.wh_agent_id` existe e `wh_agent_id_elevenlabs` não, copiar para `wh_agent_id_elevenlabs` e remover a antiga.
- Troca de provedor (`#wh-provedor` change): salva em `wh_provedor`, troca o `value` do `#wh-agent-id` pelo do provedor escolhido, troca o placeholder (`"cole aqui o Agent ID da ElevenLabs"` / `"cole aqui o Agent ID do agente Cartesia"`) e o `data-provedor` da raiz `#wh-voice`.
- Se sessão ativa ao trocar: encerrar antes.
- `#wh-toggle` click: cria o provedor via fábrica `PROVEDORES[provedorAtual].criar(ctx)` e chama `iniciar()`/`encerrar()`; demais fluxos (orbe, status, feed, barras por `ctx.volume`) continuam em voice.js — o provedor só usa o `ctx`.
- Campo de texto: quando provedor = cartesia, `#wh-text` ganha `disabled` e placeholder `"Entrada por texto disponível apenas na ElevenLabs"`; `#wh-send` disabled. `enviarTexto` que responder `{ok:false, motivo}` mostra o motivo no status por 4s.

### Verificação

1. `trilha check`
2. `curl -s localhost:3210/ | grep -c wh-provedor` → ≥ 1
3. Browser: abrir `/`, alternar o select — placeholder e valor do campo trocam; recarregar mantém o escolhido; com provedor ElevenLabs a conversa conecta como antes (mesma experiência de hoje), feed e orbe funcionam; campo de texto desabilitado em Cartesia.

### Critérios de aceite

- Todos os ids preservados; CSS existente intocado (exceto ajuste pontual).
- Comportamento ElevenLabs idêntico ao de hoje (mesmas mensagens de status).
- `trilha check` verde.

---
id: TASK-005
title: provedor Cartesia por WebSocket no widget
status: idea
spec: 001-provedor-de-voz-selecionavel-elevenlabs-cartesia
milestone: m2
covers:
  - R5
depends_on:
  - TASK-003
  - TASK-004
acceptance:
  - "sessão completa: saudação audível, feed com protocolo, painel atualiza em <=6s"
  - barge-in para o áudio na hora
  - sem dependência nova; vendor --check inalterado
checks:
  - trilha check && trilha vendor --check
created: "2026-10-05T13:01:58Z"
---

### Objetivo

Provedor Cartesia completo em `public/provedores/cartesia.js`: WebSocket bruto do Managed Agent, microfone PCM 44,1 kHz base64, playback com fila e barge-in, client tools chamando as mesmas rotas. Sem SDK novo, sem vendor novo.

### Documentação de referência (protocolo)

- Conexão: `wss://api.cartesia.ai/v1/agents/websocket/{agent_id}?cartesia_version=2026-08-14&access_token=<token>` (token vem de `GET /api/voz/cartesia?agente=...`; sem `disponivel:true` → lançar Error com a `mensagem` do servidor).
- Após abrir: enviar em até 10s `{"type":"session_create","audio":{"input_format":"pcm_44100","output_delivery":"speaking_pace"}}`.
- Receber: `session_ready` (guardar `msg.call_id`), `audio_output` (base64 PCM 44,1 kHz mono 16-bit), `audio_output_clear` (barge-in), `turn_started`, `turn_output_text_delta`, `turn_ended`, `client_tool_call`.
- Enviar áudio do mic: `{"type":"audio_input","audio":"<base64>"}` em pedaços de 20–100 ms (~2048 samples a 44,1 kHz ≈ 46 ms).
- Ferramenta: `client_tool_call {tool_call_id, tool_name, parameters, expects_response}` → responder `{"type":"client_tool_result","tool_call_id":...,"result":"<JSON string ≤4096 bytes>","is_error":bool}` (só quando `expects_response`).
- Limites: mensagens ≤ 32 KiB; inatividade 120 s (o envio contínuo do mic já reseta); closes: 1000 normal, 1008 protocolo, 1009 grande, 1011 erro deles.

### Estrutura do módulo

```js
export function criarCartesia(ctx) {
  // estado interno: ws, audioCtx, micStream, workletNode, analyserMic,
  // fontesAtivas[], proximoInstante, callId, encerrando
}
```

### iniciar()

1. `fetch("/api/voz/cartesia?agente=" + encodeURIComponent(ctx.agenteId))` → JSON; se `!disponivel` → `throw new Error(dados.mensagem)`. Guardar `access_token`.
2. `ctx.cena("conectando","Conectando — sessão assinada (Cartesia)…")`.
3. Abrir `ws` na URL do protocolo. `onopen`: enviar `session_create` (stringify). `onerror`: `throw` traduzido (guardar promessa de conexão e rejeitá-la).
4. `onmessage`: `JSON.parse` e despachar (switch abaixo). Mensagens fora do switch: ignorar.
5. Se `ws` fechar antes de `session_ready` → rejeitar com código mapeado:
   `{1008:"erro de protocolo",1009:"mensagem grande demais",1011:"falha na Cartesia"}` + `(" código " + e.code)`; outros: `"conexão fechada (" + e.code + ")"`.
6. Depois de `session_ready`: ligar microfone e playback (abaixo); `ctx.cena("ativo","Conectado — pode falar.")`; iniciar loop de volume (AnalyserNode do mic e do nó de saída, `requestAnimationFrame` chamando `ctx.volume(max(mic,out))` — parar em `encerrar`).

### Mensagens recebidas (switch)

- `session_ready`: `callId = msg.call_id`.
- `turn_started`: `ctx.cena("ativo","Ouvindo você…")`.
- `turn_output_text_delta`: primeira de um turno → `ctx.cena("ativo","A assistente está falando…")` (não acumular texto: o status só nomece o estado).
- `turn_ended`: `ctx.cena("ativo","Ouvindo você…")`.
- `audio_output`: base64 → Int16 → `AudioBuffer` (1 canal, 44100) → agendar (playback abaixo).
- `audio_output_clear`: barge-in — `fontesAtivas.forEach(s=>{try{s.stop()}catch{}})`, esvaziar `fontesAtivas`, `proximoInstante = audioCtx.currentTime`.
- `client_tool_call`: `despacharFerramenta(msg)` (abaixo).

### Microfone

```js
micStream = await navigator.mediaDevices.getUserMedia({
  audio: { echoCancellation: true, noiseSuppression: true, autoGainControl: true },
});
audioCtx = new AudioContext({ sampleRate: 44100 }); // se o construtor lançar:
// throw new Error("Seu navegador não abriu áudio a 44,1 kHz para a Cartesia.");
const fonte = audioCtx.createMediaStreamSource(micStream);
```

Worklet por blob (CSP já permite `blob:` em script-src/worker-src — `URL.createObjectURL`):

```js
const CODIGO_WORKLET = `
class Capturador extends AudioWorkletProcessor {
  process(entradas) {
    const canal = entradas[0][0];
    if (canal) this.port.postMessage(canal.slice(0));
    return true;
  }
}
registerProcessor("capturador", Capturador);
`;
const urlWorklet = URL.createObjectURL(new Blob([CODIGO_WORKLET], { type: "application/javascript" }));
await audioCtx.audioWorklet.addModule(urlWorklet);
workletNode = new AudioWorkletNode(audioCtx, "capturador");
workletNode.port.onmessage = (ev) => enviarChunk(ev.data); // Float32Array de 2048
fonte.connect(workletNode);
// Analyser para o medidor: fonte.connect(analyserMic)
```

`enviarChunk(float32)`:
1. Int16 com clamp: `v = Math.max(-1, Math.min(1, x)); int16[i] = v < 0 ? v * 0x8000 : v * 0x7fff`.
2. Bytes little-endian em `Uint8Array` do `ArrayBuffer` do Int16.
3. Base64 SEM spread (limite de argumentos): laço acumulando `String.fromCharCode(...)` em fatias de 0x8000.
4. Se `ws.readyState === 1`: `ws.send(JSON.stringify({ type: "audio_input", audio: b64 }))`.

### Playback

```js
function tocar(b64) {
  const bin = atob(b64);
  const bytes = new Uint8Array(bin.length);
  for (let i = 0; i < bin.length; i++) bytes[i] = bin.charCodeAt(i);
  const vista = new Int16Array(bytes.buffer);
  const buffer = audioCtx.createBuffer(1, vista.length, 44100);
  const canal = buffer.getChannelData(0);
  for (let i = 0; i < vista.length; i++) canal[i] = vista[i] / 32768;
  const fonte = audioCtx.createBufferSource();
  fonte.buffer = buffer;
  fonte.connect(audioCtx.destination);
  proximoInstante = Math.max(proximoInstante, audioCtx.currentTime + 0.05);
  fonte.start(proximoInstante);
  proximoInstante += buffer.duration;
  fontesAtivas.push(fonte);
  fonte.onended = () => { fontesAtivas = fontesAtivas.filter(f => f !== fonte); };
}
```

### Ferramentas (despacharFerramenta)

Mapa (idêntico ao da ElevenLabs):

```js
const ROTAS = {
  abrir_chamado_tecnico: "/api/tools/abrir_chamado_tecnico",
  enviar_segunda_via_boleto: "/api/tools/enviar_segunda_via_boleto",
  solicitar_religue_confirmacao: "/api/tools/solicitar_religue_confirmacao",
};
```

Fluxo: se `!msg.expects_response` → `ctx.anotar("• " + msg.tool_name + " (sem resposta esperada)")` e return.
Senão: `fetch(rota, {method:"POST", headers:{"Content-Type":"application/json"}, body: JSON.stringify({...msg.parameters, conversation_id: callId})})`; em sucesso HTTP, `dados = await resp.json()`; `ok = dados.status === "sucesso" || dados.status === "solicitacao_registrada"`;
`ctx.anotar(dados.protocolo ? "✓ " + msg.tool_name + " — " + dados.protocolo : "• " + msg.tool_name + " (" + (dados.status ?? "sem resposta") + ")")`;
enviar `client_tool_result` com `result: JSON.stringify(dados)` e `is_error: !ok`. Erro de rede: `is_error: true`, result `"{\"status\":\"erro\",\"mensagem\":\"Falha de rede ao registrar o atendimento. Peça para tentar de novo.\"}"`.

### encerrar()

Flag `encerrando = true` (para ignorar o close handler); `try { ws.close(1000) } catch {}`; parar tracks do mic; `workletNode?.disconnect()`; `fontesAtivas.forEach(s=>{try{s.stop()}catch{}})`; `audioCtx?.close()`; cancelar o rAF do volume.

### enviarTexto()

`return { ok: false, motivo: "A entrada por texto não existe no WebSocket de agentes da Cartesia nesta versão — use o microfone." }`.

### Verificação

1. `trilha check`
2. `./dev.sh` e abrir `/`; selecionar Cartesia, colar o Agent ID, "Iniciar conversa": status vai a "Conectando — sessão assinada (Cartesia)…" → "Conectado — pode falar." (a saudação `initial_message` toca no alto-falante e o orbe reage).
3. Falar "minha internet caiu e quero abrir um chamado": feed registra `✓ abrir_chamado_tecnico — CH-…` e `/painel` ganha a linha (origem `voz`) em ≤ 6 s.
4. Interromper a assistente falando (barge-in): áudio para na hora, sem sobreposição.
5. Recarregar com provedor Cartesia: continua selecionado; campo de texto desabilitado com o motivo.

### Critérios de aceite

- Sem dependência nova; `trilha vendor --check` inalterado.
- Token nunca impresso no console nem em log do browser.
- Close codes mapeados para mensagens legíveis; erro de rede vira frase no status.

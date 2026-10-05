/* Provedor Cartesia: Managed Agent falando por WebSocket bruto
   (wss://api.cartesia.ai/v1/agents/websocket/{agent_id}). O servidor cunha
   o access token (GET /api/voz/cartesia) — a chave nunca chega aqui.
   Microfone em PCM 44,1 kHz base64; playback com fila e barge-in;
   client tools chamando as mesmas rotas da ElevenLabs. */

const VERSAO = "2026-08-14";

const ROTAS = {
  abrir_chamado_tecnico: "/api/tools/abrir_chamado_tecnico",
  enviar_segunda_via_boleto: "/api/tools/enviar_segunda_via_boleto",
  solicitar_religue_confirmacao: "/api/tools/solicitar_religue_confirmacao",
};

const CLOSES = {
  1006: "conexão recusada — verifique os créditos da conta Cartesia (Agents usage limit) e a chave do servidor",
  1008: "erro de protocolo",
  1009: "mensagem grande demais",
  1011: "falha na Cartesia",
};

export function criarCartesia(ctx) {
  let ws = null;
  let audioCtx = null;
  let micStream = null;
  let workletNode = null;
  let analyserMic = null;
  let analyserSaida = null;
  let fontesAtivas = [];
  let proximoInstante = 0;
  let callId = "";
  let encerrando = false;
  let quadro = 0;

  async function iniciar(agenteId) {
    ctx.cena("conectando", "Conectando — pedindo sessão assinada (Cartesia)…");

    const resp = await fetch("/api/voz/cartesia?agente=" + encodeURIComponent(agenteId));
    const sessao = await resp.json().catch(() => null);
    if (!resp.ok || !sessao?.disponivel || !sessao.access_token) {
      throw new Error(sessao?.mensagem || "A Cartesia não está disponível neste servidor.");
    }

    const url = new URL("wss://api.cartesia.ai/v1/agents/websocket/" + encodeURIComponent(sessao.agent_id));
    url.searchParams.set("cartesia_version", VERSAO);
    url.searchParams.set("access_token", sessao.access_token);

    await new Promise((resolve, rejeita) => {
      let aberto = false;
      ws = new WebSocket(url.toString());
      ws.onopen = () => {
        aberto = true;
        ws.send(JSON.stringify({
          type: "session_create",
          audio: { input_format: "pcm_44100", output_delivery: "speaking_pace" },
        }));
      };
      ws.onmessage = (ev) => {
        let msg;
        try { msg = JSON.parse(ev.data); } catch { return; }
        const pronta = processar(msg);
        if (pronta && !resolveChamado) { resolveChamado = true; resolve(); }
      };
      ws.onclose = (e) => {
        if (encerrando) return;
        const motivo = CLOSES[e.code] || "conexão fechada (" + e.code + ")";
        if (!resolveChamado) { resolveChamado = true; rejeita(new Error(motivo)); }
        else ctx.onEncerrado();
      };
      ws.onerror = () => {
        if (!aberto && !resolveChamado) {
          resolveChamado = true;
          rejeita(new Error("não consegui abrir o WebSocket com a Cartesia"));
        }
      };
    });

    ctx.cena("conectando", "Conectado — abrindo o microfone…");
    await ligarMicrofone();
    ctx.cena("ativo", "Conectado — pode falar.");
    animar();
  }

  let resolveChamado = false;

  // Devolve true quando a mensagem completa a conexão (session_ready).
  function processar(msg) {
    switch (msg.type) {
      case "session_ready":
        callId = msg.call_id || "";
        return true;
      case "turn_started":
        ctx.cena("ativo", "Ouvindo você…");
        return false;
      case "turn_output_text_delta":
        ctx.cena("ativo", "A assistente está falando…");
        return false;
      case "turn_ended":
        ctx.cena("ativo", "Ouvindo você…");
        return false;
      case "audio_output":
        if (msg.audio) tocar(msg.audio);
        return false;
      case "audio_output_clear":
        // barge-in: o cliente falou em cima — solta o que estava na fila
        for (const fonte of fontesAtivas) { try { fonte.stop(); } catch { /* já parou */ } }
        fontesAtivas = [];
        proximoInstante = audioCtx ? audioCtx.currentTime : 0;
        return false;
      case "client_tool_call":
        despacharFerramenta(msg);
        return false;
    }
    return false;
  }

  async function ligarMicrofone() {
    micStream = await navigator.mediaDevices.getUserMedia({
      audio: { echoCancellation: true, noiseSuppression: true, autoGainControl: true },
    }).catch((e) => {
      throw new Error("Preciso da permissão do microfone para conversar (" + (e.name || "bloqueado") + ").");
    });

    try {
      audioCtx = new AudioContext({ sampleRate: 44100 });
    } catch {
      throw new Error("Seu navegador não abriu áudio a 44,1 kHz para a Cartesia.");
    }

    // Captura: worklet por blob (a CSP já permite blob: em script-src) que
    // manda fatias de Float32 — ~2048 amostras ≈ 46 ms — para o WebSocket.
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
    workletNode.port.onmessage = (ev) => enviarChunk(ev.data);

    const fonte = audioCtx.createMediaStreamSource(micStream);
    fonte.connect(workletNode);

    analyserMic = audioCtx.createAnalyser();
    analyserMic.fftSize = 512;
    fonte.connect(analyserMic);

    // Saída: as fontes de playback passam por um analyser antes do destino,
    // para o orbe reagir à voz da assistente também.
    analyserSaida = audioCtx.createAnalyser();
    analyserSaida.fftSize = 512;
    analyserSaida.connect(audioCtx.destination);
  }

  function enviarChunk(float32) {
    if (!ws || ws.readyState !== 1) return;
    const int16 = new Int16Array(float32.length);
    for (let i = 0; i < float32.length; i++) {
      const v = Math.max(-1, Math.min(1, float32[i]));
      int16[i] = v < 0 ? v * 0x8000 : v * 0x7fff;
    }
    const bytes = new Uint8Array(int16.buffer);
    const b64 = base64(bytes);
    ws.send(JSON.stringify({ type: "audio_input", audio: b64 }));
  }

  // Base64 sem spread (limite de argumentos do fromCharCode).
  function base64(bytes) {
    let binario = "";
    const FATIA = 0x8000;
    for (let i = 0; i < bytes.length; i += FATIA) {
      binario += String.fromCharCode.apply(null, bytes.subarray(i, i + FATIA));
    }
    return btoa(binario);
  }

  function tocar(b64) {
    const binario = atob(b64);
    const bytes = new Uint8Array(binario.length);
    for (let i = 0; i < binario.length; i++) bytes[i] = binario.charCodeAt(i);
    const amostras = new Int16Array(bytes.buffer);
    const buffer = audioCtx.createBuffer(1, amostras.length, 44100);
    const canal = buffer.getChannelData(0);
    for (let i = 0; i < amostras.length; i++) canal[i] = amostras[i] / 32768;

    const fonte = audioCtx.createBufferSource();
    fonte.buffer = buffer;
    fonte.connect(analyserSaida);
    proximoInstante = Math.max(proximoInstante, audioCtx.currentTime + 0.05);
    fonte.start(proximoInstante);
    proximoInstante += buffer.duration;
    fontesAtivas.push(fonte);
    fonte.onended = () => {
      fontesAtivas = fontesAtivas.filter((f) => f !== fonte);
    };
  }

  // Mesmas rotas e mesmos nomes de ferramenta da ElevenLabs: o backend não
  // sabe de qual provedor veio o atendimento.
  async function despacharFerramenta(msg) {
    const rota = ROTAS[msg.tool_name];
    if (!rota) {
      if (msg.expects_response) responder(msg, { status: "erro", mensagem: "Ferramenta desconhecida." }, true);
      return;
    }
    if (!msg.expects_response) {
      ctx.anotar("• " + msg.tool_name + " (sem resposta esperada)");
      return;
    }
    try {
      const resp = await fetch(rota, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ ...(msg.parameters ?? {}), conversation_id: callId }),
      });
      const dados = await resp.json();
      const ok = dados.status === "sucesso" || dados.status === "solicitacao_registrada";
      if (dados.protocolo) ctx.anotar("✓ " + msg.tool_name + " — " + dados.protocolo);
      else ctx.anotar("• " + msg.tool_name + " (" + (dados.status ?? "sem resposta") + ")");
      responder(msg, dados, !ok);
    } catch {
      ctx.anotar("✕ " + msg.tool_name + " — falha de rede");
      responder(msg, { status: "erro", mensagem: "Falha de rede ao registrar o atendimento. Peça para tentar de novo." }, true);
    }
  }

  function responder(msg, dados, erro) {
    let resultado;
    try {
      resultado = JSON.stringify(dados).slice(0, 4096);
    } catch {
      resultado = '{"status":"erro","mensagem":"Resposta inválida da ferramenta."}';
    }
    ws.send(JSON.stringify({
      type: "client_tool_result",
      tool_call_id: msg.tool_call_id,
      result: resultado,
      is_error: erro,
    }));
  }

  function animar() {
    const dados = new Uint8Array(256);
    const passo = () => {
      let nivel = 0;
      for (const analisador of [analyserMic, analyserSaida]) {
        if (!analisador) continue;
        analisador.getByteTimeDomainData(dados);
        let pico = 0;
        for (let i = 0; i < dados.length; i++) {
          const desvio = Math.abs(dados[i] - 128) / 128;
          if (desvio > pico) pico = desvio;
        }
        if (pico > nivel) nivel = pico;
      }
      ctx.volume(Math.min(1, nivel * 3));
      quadro = requestAnimationFrame(passo);
    };
    cancelAnimationFrame(quadro);
    quadro = requestAnimationFrame(passo);
  }

  async function encerrar() {
    encerrando = true;
    try { if (ws && ws.readyState <= 1) ws.close(1000); } catch { /* já fechado */ }
    ws = null;
    if (micStream) for (const faixa of micStream.getTracks()) faixa.stop();
    micStream = null;
    if (workletNode) { try { workletNode.disconnect(); } catch { /* já solto */ } }
    workletNode = null;
    for (const fonte of fontesAtivas) { try { fonte.stop(); } catch { /* já parou */ } }
    fontesAtivas = [];
    cancelAnimationFrame(quadro);
    ctx.volume(0);
    if (audioCtx) { try { await audioCtx.close(); } catch { /* já fechado */ } }
    audioCtx = null;
    analyserMic = analyserSaida = null;
  }

  return {
    iniciar,
    encerrar,
    async enviarTexto() {
      return { ok: false, motivo: "A entrada por texto não existe no WebSocket de agentes da Cartesia nesta versão — use o microfone." };
    },
  };
}

/* Widget de voz da WaveHub — conversa com o agente da ElevenLabs e executa
   as client tools chamando a API deste mesmo app. Sem segredos: o Agent ID
   de agente público é identificador, não credencial. */

// O SDK vem do bundle local (public/vendor, sha fixado no vendor.lock).
let Conversation;
try {
  ({ Conversation } = await import("/vendor/elevenlabs-client.js"));
} catch (e) {
  const raiz = document.getElementById("wh-voice");
  if (raiz) {
    document.getElementById("wh-status").textContent =
      "SDK da ElevenLabs não carregou — verifique o deploy (public/vendor).";
  }
  throw e;
}

const root = document.getElementById("wh-voice");
if (root) init(root);

function init(root) {
  const orb = document.getElementById("wh-orb");
  const statusEl = document.getElementById("wh-status");
  const toggle = document.getElementById("wh-toggle");
  const input = document.getElementById("wh-text");
  const send = document.getElementById("wh-send");
  const feed = document.getElementById("wh-feed");
  const agentInput = document.getElementById("wh-agent-id");

  let convo = null;
  let quadro = 0;

  const CHAVE_AGENTE = "wh_agent_id";
  agentInput.value = agentInput.value.trim()
    || localStorage.getItem(CHAVE_AGENTE)
    || root.dataset.agentId
    || "";

  toggle.addEventListener("click", () => (convo ? encerrar(false) : iniciar()));
  send.addEventListener("click", enviarTexto);
  input.addEventListener("keydown", (e) => {
    if (e.key === "Enter") enviarTexto();
  });

  function estado(texto) {
    statusEl.textContent = texto;
  }

  function cena(state, texto) {
    root.dataset.state = state;
    if (texto !== undefined) estado(texto);
  }

  async function iniciar() {
    const agenteId = agentInput.value.trim();
    if (!agenteId) {
      cena("idle", "Cole o Agent ID do agente para começar — o passo a passo está em Agente IA.");
      agentInput.focus();
      return;
    }
    localStorage.setItem(CHAVE_AGENTE, agenteId);
    toggle.disabled = true;
    cena("conectando", "Conectando com a ElevenLabs…");

    // Agente privado: o servidor assina a sessão com a chave de API (que
    // nunca chega ao navegador). Sem chave configurada (503), o modo
    // público segue em silêncio; chave recusada vira aviso para o usuário.
    let sessao = { agentId: agenteId };
    let avisoAssinatura = "";
    try {
      const assinatura = await fetch("/api/voz/assinada?agente=" + encodeURIComponent(agenteId));
      if (assinatura.ok) {
        const dados = await assinatura.json();
        if (dados.signed_url) sessao = { signedUrl: dados.signed_url };
      } else if (assinatura.status !== 503) {
        const dados = await assinatura.json().catch(() => null);
        if (dados?.mensagem) avisoAssinatura = dados.mensagem + " ";
      }
    } catch {
      /* sem assinatura disponível: segue com o Agent ID */
    }

    // Transparência do modo: agente privado (sessão assinada) ou público.
    cena("conectando", sessao.signedUrl
      ? "Conectando — sessão assinada (agente privado)…"
      : "Conectando — modo público (Agent ID direto)…");

    try {
      convo = await Conversation.startSession({
        ...sessao,
        clientTools: ferramentas(),
        onConnect: () => {
          toggle.disabled = false;
          toggle.textContent = "Encerrar conversa";
          cena("ativo", "Conectado — pode falar.");
        },
        onDisconnect: () => encerrar(true),
        onError: (e) => estado("Erro na conversa: " + mensagem(e)),
        onModeChange: ({ mode }) => {
          if (mode === "speaking") cena("ativo", "A assistente está falando…");
          else cena("ativo", "Ouvindo você…");
        },
      });
    } catch (e) {
      convo = null;
      toggle.disabled = false;
      const detalhe = mensagem(e);
      const dica = /\bfetch\b/i.test(detalhe)
        ? avisoAssinatura || " — se o agente for privado, configure ELEVENLABS_API_KEY no servidor (veja /agente)."
        : "";
      cena("idle", "Não consegui conectar: " + detalhe + dica);
    }
  }

  async function encerrar(silencioso) {
    try {
      if (convo) await convo.endSession();
    } catch {
      /* a sessão já tinha morado */
    }
    convo = null;
    pararBarras();
    toggle.disabled = false;
    toggle.textContent = "Iniciar conversa";
    cena("idle", silencioso ? "Conversa encerrada." : "Conversa encerrada. Até a próxima!");
  }

  // As client tools conversam com a API do próprio app; o JSON devolvido é
  // o que a assistente lê em voz alta.
  function ferramentas() {
    const chamar = (rota, rotulo) => async (params) => {
      try {
        const resp = await fetch(rota, {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify(params ?? {}),
        });
        const dados = await resp.json();
        if (dados.protocolo) anotar("✓ " + rotulo + " — " + dados.protocolo);
        else anotar("• " + rotulo + " (" + (dados.status ?? "sem resposta") + ")");
        return dados;
      } catch {
        anotar("✕ " + rotulo + " — falha de rede");
        return { status: "erro", mensagem: "Falha de rede ao registrar o atendimento. Peça para tentar de novo." };
      }
    };
    return {
      abrir_chamado_tecnico: chamar("/api/tools/abrir_chamado_tecnico", "Chamado técnico aberto"),
      enviar_segunda_via_boleto: chamar("/api/tools/enviar_segunda_via_boleto", "Segunda via enviada"),
      solicitar_religue_confirmacao: chamar("/api/tools/solicitar_religue_confirmacao", "Religue solicitado"),
    };
  }

  function anotar(texto) {
    feed.querySelector(".wh-feed-empty")?.remove();
    const li = document.createElement("li");
    li.textContent = texto;
    feed.prepend(li);
  }

  // Barras do orbe seguem o volume real de entrada e saída, quando o SDK
  // expõe os medidores; sem eles, as barras ficam em repouso.
  function animarBarras() {
    const barras = orb.children;
    let t = 0;
    const passo = () => {
      t += 0.08;
      let entrada = 0;
      let saida = 0;
      try {
        entrada = Number(convo?.getInputVolume?.() ?? 0);
        saida = Number(convo?.getOutputVolume?.() ?? 0);
      } catch {
        /* SDK sem medidores: fica no repouso */
      }
      if (!Number.isFinite(entrada)) entrada = 0;
      if (!Number.isFinite(saida)) saida = 0;
      const nivel = Math.max(entrada, saida);
      for (let i = 0; i < barras.length; i++) {
        const forma = Math.abs(Math.sin(t + i * 0.7));
        const altura = 8 + nivel * 44 * (0.45 + 0.55 * forma);
        barras[i].style.height = altura.toFixed(1) + "px";
      }
      quadro = requestAnimationFrame(passo);
    };
    cancelAnimationFrame(quadro);
    quadro = requestAnimationFrame(passo);
  }

  function pararBarras() {
    cancelAnimationFrame(quadro);
    for (const barra of orb.children) barra.style.height = "";
  }

  async function enviarTexto() {
    const texto = input.value.trim();
    if (!texto) return;
    if (!convo) {
      estado("Inicie a conversa para escrever — sem sessão não há quem ouça.");
      return;
    }
    if (typeof convo.sendUserMessage !== "function") {
      estado("Esta versão do SDK não aceita texto; use o microfone.");
      return;
    }
    input.value = "";
    try {
      await convo.sendUserMessage(texto);
    } catch (e) {
      estado("Não consegui enviar: " + mensagem(e));
    }
  }

  const conectado = new MutationObserver(() => {
    if (root.dataset.state === "ativo") animarBarras();
    else pararBarras();
  });
  conectado.observe(root, { attributeFilter: ["data-state"] });

  function mensagem(e) {
    if (!e) return "erro desconhecido";
    if (typeof e === "string") return e;
    return e.message || e.statusText || String(e);
  }
}

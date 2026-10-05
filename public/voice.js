/* Widget de voz da WaveHub — maestro dos provedores (ElevenLabs e Cartesia).
   Guarda a máquina de estados, o seletor de provedor, o orbe e o feed; cada
   provedor vive em public/provedores/<nome>.js e só usa o ctx daqui.
   Sem segredos: Agent IDs de agente publicado são identificadores. */

// Os provedores são carregados dinamicamente para um problema de deploy
// (vendor faltando, por exemplo) virar frase no widget, não tela morta.
let criarElevenLabs;
let criarCartesia;
try {
  ({ criarElevenLabs } = await import("/provedores/elevenlabs.js"));
  ({ criarCartesia } = await import("/provedores/cartesia.js"));
} catch (e) {
  const raiz = document.getElementById("wh-voice");
  if (raiz) {
    document.getElementById("wh-status").textContent =
      "Os módulos de voz não carregaram — verifique o deploy (public/provedores e public/vendor).";
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
  const provedorSelect = document.getElementById("wh-provedor");

  const CHAVES = {
    elevenlabs: "wh_agent_id_elevenlabs",
    cartesia: "wh_agent_id_cartesia",
  };

  // Migração: a chave antiga única vira a da ElevenLabs.
  if (localStorage.getItem(CHAVES.elevenlabs) === null && localStorage.getItem("wh_agent_id")) {
    localStorage.setItem(CHAVES.elevenlabs, localStorage.getItem("wh_agent_id"));
    localStorage.removeItem("wh_agent_id");
  }

  let provedorNome = localStorage.getItem("wh_provedor") === "cartesia" ? "cartesia" : "elevenlabs";
  let provedor = null; // instância ativa (iniciar/encerrar/enviarTexto)
  let avisoTemporario = null;

  const FABRICAS = {
    elevenlabs: criarElevenLabs,
    cartesia: criarCartesia,
  };

  function padraoDoProvedor() {
    return provedorNome === "cartesia"
      ? root.dataset.agentIdCartesia || ""
      : root.dataset.agentId || "";
  }

  function agenteAtual() {
    return agentInput.value.trim()
      || localStorage.getItem(CHAVES[provedorNome])
      || padraoDoProvedor();
  }

  function aplicarProvedor() {
    provedorSelect.value = provedorNome;
    agentInput.value = localStorage.getItem(CHAVES[provedorNome]) || padraoDoProvedor();
    agentInput.placeholder = provedorNome === "cartesia"
      ? "cole aqui o Agent ID do agente Cartesia"
      : "cole aqui o Agent ID da ElevenLabs";
    root.dataset.provedor = provedorNome;
    const semTexto = provedorNome === "cartesia";
    input.disabled = semTexto;
    send.disabled = semTexto;
    input.placeholder = semTexto
      ? "Entrada por texto disponível apenas na ElevenLabs"
      : "Sem microfone? Escreva sua mensagem…";
  }

  function estado(texto) {
    statusEl.textContent = texto;
  }

  function cena(estadoNovo, texto) {
    root.dataset.state = estadoNovo;
    if (texto !== undefined) estado(texto);
  }

  function anotar(texto) {
    feed.querySelector(".wh-feed-empty")?.remove();
    const li = document.createElement("li");
    li.textContent = texto;
    feed.prepend(li);
  }

  // Barras do orbe: forma de onda desenhada do nível (0–1) que o provedor
  // mede de verdade no microfone e na saída.
  function volume(nivel, t) {
    const barras = orb.children;
    const formaBase = t === undefined ? 0 : t;
    for (let i = 0; i < barras.length; i++) {
      const forma = Math.abs(Math.sin(formaBase * 12 + i * 0.7));
      const altura = 8 + nivel * 44 * (0.45 + 0.55 * forma);
      barras[i].style.height = altura.toFixed(1) + "px";
    }
  }

  provedorSelect.addEventListener("change", async () => {
    if (provedor) await encerrar(true);
    provedorNome = provedorSelect.value;
    localStorage.setItem("wh_provedor", provedorNome);
    aplicarProvedor();
  });

  toggle.addEventListener("click", () => (provedor ? encerrar(false) : iniciar()));
  send.addEventListener("click", enviarTexto);
  input.addEventListener("keydown", (e) => {
    if (e.key === "Enter") enviarTexto();
  });

  async function iniciar() {
    const agenteId = agenteAtual();
    if (!agenteId) {
      cena("idle", "Cole o Agent ID do agente para começar — o passo a passo está em Agente IA.");
      agentInput.focus();
      return;
    }
    agentInput.value = agenteId;
    localStorage.setItem(CHAVES[provedorNome], agenteId);
    toggle.disabled = true;
    cena("conectando", "Conectando com " + (provedorNome === "cartesia" ? "a Cartesia" : "a ElevenLabs") + "…");

    provedor = FABRICAS[provedorNome](ctx(agenteId));
    try {
      await provedor.iniciar(agenteId);
      toggle.disabled = false;
      toggle.textContent = "Encerrar conversa";
    } catch (e) {
      provedor = null;
      toggle.disabled = false;
      volume(0);
      cena("idle", "Não consegui conectar: " + mensagem(e));
    }
  }

  async function encerrar(silencioso) {
    if (provedor) await provedor.encerrar();
    provedor = null;
    toggle.disabled = false;
    toggle.textContent = "Iniciar conversa";
    volume(0);
    cena("idle", silencioso ? "Conversa encerrada." : "Conversa encerrada. Até a próxima!");
  }

  function ctx(agenteId) {
    return {
      agenteId,
      cena,
      anotar,
      volume,
      onEncerrado: () => encerrar(true),
    };
  }

  async function enviarTexto() {
    const texto = input.value.trim();
    if (!texto) return;
    if (!provedor) {
      estado("Inicie a conversa para escrever — sem sessão não há quem ouça.");
      return;
    }
    const resultado = await provedor.enviarTexto(texto);
    if (resultado.ok) {
      input.value = "";
      return;
    }
    if (resultado.motivo) {
      const antes = statusEl.textContent;
      estado(resultado.motivo);
      clearTimeout(avisoTemporario);
      avisoTemporario = setTimeout(() => estado(antes), 4000);
    }
  }

  aplicarProvedor();

  function mensagem(e) {
    if (!e) return "erro desconhecido";
    if (typeof e === "string") return e;
    return e.message || e.statusText || String(e);
  }
}

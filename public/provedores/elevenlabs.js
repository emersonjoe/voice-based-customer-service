/* Provedor ElevenLabs: conversa por WebRTC com o agente ConvAI e executa as
   client tools chamando a API deste app. A sessão de agente privado vem
   assinada do servidor (GET /api/voz/assinada); a chave nunca chega aqui. */

import { Conversation } from "/vendor/elevenlabs-client.js";

export function criarElevenLabs(ctx) {
  let convo = null;
  let quadro = 0;

  // Barras do orbe seguem o volume real de entrada e saída, quando o SDK
  // expõe os medidores; sem eles, o nível fica em zero.
  function animar() {
    let t = 0;
    const passo = () => {
      t += 0.08;
      let entrada = 0;
      let saida = 0;
      try {
        entrada = Number(convo?.getInputVolume?.() ?? 0);
        saida = Number(convo?.getOutputVolume?.() ?? 0);
      } catch {
        /* SDK sem medidores */
      }
      const nivel = Math.max(
        Number.isFinite(entrada) ? entrada : 0,
        Number.isFinite(saida) ? saida : 0
      );
      ctx.volume(nivel, t);
      quadro = requestAnimationFrame(passo);
    };
    cancelAnimationFrame(quadro);
    quadro = requestAnimationFrame(passo);
  }

  function parar() {
    cancelAnimationFrame(quadro);
    ctx.volume(0);
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
        if (dados.protocolo) ctx.anotar("✓ " + rotulo + " — " + dados.protocolo);
        else ctx.anotar("• " + rotulo + " (" + (dados.status ?? "sem resposta") + ")");
        return dados;
      } catch {
        ctx.anotar("✕ " + rotulo + " — falha de rede");
        return { status: "erro", mensagem: "Falha de rede ao registrar o atendimento. Peça para tentar de novo." };
      }
    };
    return {
      abrir_chamado_tecnico: chamar("/api/tools/abrir_chamado_tecnico", "Chamado técnico aberto"),
      enviar_segunda_via_boleto: chamar("/api/tools/enviar_segunda_via_boleto", "Segunda via enviada"),
      solicitar_religue_confirmacao: chamar("/api/tools/solicitar_religue_confirmacao", "Religue solicitado"),
    };
  }

  return {
    async iniciar(agenteId) {
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

      try {
        convo = await Conversation.startSession({
          ...sessao,
          clientTools: ferramentas(),
          onConnect: () => {
            ctx.cena("ativo", "Conectado — pode falar.");
            animar();
          },
          onDisconnect: () => ctx.onEncerrado(),
          onError: (e) => ctx.cena("ativo", "Erro na conversa: " + mensagem(e)),
          onModeChange: ({ mode }) => {
            if (mode === "speaking") ctx.cena("ativo", "A assistente está falando…");
            else ctx.cena("ativo", "Ouvindo você…");
          },
        });
      } catch (e) {
        convo = null;
        parar();
        const detalhe = mensagem(e);
        const dica = /\bfetch\b/i.test(detalhe)
          ? avisoAssinatura || " — se o agente for privado, configure ELEVENLABS_API_KEY no servidor (veja /agente)."
          : "";
        throw new Error(detalhe + dica);
      }
    },

    async encerrar() {
      try {
        if (convo) await convo.endSession();
      } catch {
        /* a sessão já tinha morado */
      }
      convo = null;
      parar();
    },

    async enviarTexto(texto) {
      if (!convo) return { ok: false, motivo: "Inicie a conversa para escrever — sem sessão não há quem ouça." };
      if (typeof convo.sendUserMessage !== "function") {
        return { ok: false, motivo: "Esta versão do SDK não aceita texto; use o microfone." };
      }
      await convo.sendUserMessage(texto);
      return { ok: true };
    },
  };

  function mensagem(e) {
    if (!e) return "erro desconhecido";
    if (typeof e === "string") return e;
    return e.message || e.statusText || String(e);
  }
}

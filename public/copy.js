/* Botões de copiar da página /agente: um único listener por delegação —
   nada de script inline (a CSP bloqueia) e nenhum segredo em jogo. */

document.addEventListener("click", async (e) => {
  const botao = e.target.closest(".wh-copy");
  if (!botao) return;

  const alvo = document.getElementById(botao.dataset.copyTarget);
  if (!alvo) return;

  const texto = alvo.dataset.copy ?? alvo.textContent;
  const original = botao.textContent;

  // Clipboard pode ficar preso esperando permissão: corrida com limite e
  // fallback de seleção para copiar à mão.
  const comLimite = Promise.race([
    navigator.clipboard.writeText(texto),
    new Promise((_, rejeita) => setTimeout(() => rejeita(new Error("tempo")), 600)),
  ]);

  try {
    await comLimite;
    botao.textContent = "Copiado!";
  } catch {
    const faixa = document.createRange();
    faixa.selectNodeContents(alvo);
    const selecao = getSelection();
    selecao.removeAllRanges();
    selecao.addRange(faixa);
    botao.textContent = "Selecione e copie";
  }

  setTimeout(() => {
    botao.textContent = original;
  }, 1600);
});

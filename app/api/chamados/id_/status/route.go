// A mudança de status é o POST dos formulários do painel — da listagem e da
// tela do atendimento. O Kind de página do ramo garante o CSRF.
package status

import (
	"github.com/emersonjoe/trilha"
	"github.com/emersonjoe/trilha/ui"

	"github.com/emersonjoe/voice-based-customer-service/internal/atendimento"
)

func POST(c *trilha.Ctx) error {
	id := c.Param("id")
	novo := atendimento.Status(c.Form("status"))

	// O caminho de volta vem do formulário e só pode ser caminho local:
	// nada de redirect aberto para fora.
	voltar := c.Form("voltar")
	if len(voltar) < 1 || voltar[0] != '/' || (len(voltar) > 1 && voltar[1] == '/') {
		voltar = "/painel"
	}

	if !novo.Valido() {
		c.Flash(ui.FlashError, "Status desconhecido.")
		return c.Redirect(voltar)
	}
	store := trilha.Use[*atendimento.Store](c)
	if !store.AtualizarStatus(id, novo) {
		c.Flash(ui.FlashError, "Atendimento não encontrado.")
		return c.Redirect(voltar)
	}
	c.Flash(ui.FlashSuccess, "Status atualizado para "+novo.Rotulo()+".")
	return c.Redirect(voltar)
}

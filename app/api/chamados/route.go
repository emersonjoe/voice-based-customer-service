// A listagem JSON dos atendimentos, para consulta de fora do painel —
// integrar um CRM, testar no Postman ou demonstrar a API na apresentação.
package chamados

import (
	"net/http"

	"github.com/emersonjoe/trilha"

	"github.com/emersonjoe/voice-based-customer-service/internal/apiutil"
	"github.com/emersonjoe/voice-based-customer-service/internal/atendimento"
	"github.com/emersonjoe/voice-based-customer-service/internal/limite"
)

func GET(c *trilha.Ctx) error {
	if !apiutil.Limita(c.Writer(), c.Request(), trilha.Use[*limite.Limitador](c), "listar") {
		return nil
	}

	var f struct {
		trilha.ListParams
		Tipo   string `form:"tipo"`
		Status string `form:"status"`
	}
	if err := c.Bind(&f); err != nil {
		return err
	}
	filtro := atendimento.Filtro{ListParams: f.ListParams}
	if atendimento.Tipo(f.Tipo).Valido() {
		filtro.Tipo = atendimento.Tipo(f.Tipo)
	}
	if atendimento.Status(f.Status).Valido() {
		filtro.Status = atendimento.Status(f.Status)
	}

	store := trilha.Use[*atendimento.Store](c)
	atts, total := store.Listar(filtro)
	return c.JSON(http.StatusOK, map[string]any{
		"total":        total,
		"pagina":       f.Page,
		"atendimentos": atts,
	})
}

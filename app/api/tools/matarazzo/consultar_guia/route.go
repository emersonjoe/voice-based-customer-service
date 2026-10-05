// A ferramenta consultar_guia_matarazzo devolve ao concierge o conteúdo
// curado do guia do complexo. Mesmo desenho das ferramentas da Wave:
// corpo tolerante, rate limit e resposta JSON pensada para ser falada.
package consultar_guia

import (
	"net/http"
	"strings"

	"github.com/emersonjoe/trilha"

	"github.com/emersonjoe/voice-based-customer-service/internal/apiutil"
	"github.com/emersonjoe/voice-based-customer-service/internal/limite"
	"github.com/emersonjoe/voice-based-customer-service/internal/matarazzo"
)

type resposta struct {
	Status             string   `json:"status"`
	Topico             string   `json:"topico,omitempty"`
	Titulo             string   `json:"titulo,omitempty"`
	Texto              string   `json:"texto,omitempty"`
	Menu               []string `json:"menu,omitempty"`
	Horario            string   `json:"horario,omitempty"`
	Dicas              []string `json:"dicas,omitempty"`
	TopicosDisponiveis []string `json:"topicos_disponiveis,omitempty"`
	Mensagem           string   `json:"mensagem"`
}

func POST(c *trilha.Ctx) error {
	w, r := c.Writer(), c.Request()
	if !apiutil.Limita(w, r, trilha.Use[*limite.Limitador](c), "guia_matarazzo") {
		return nil
	}
	corpo, err := apiutil.Corpo(r)
	if err != nil {
		return c.JSON(http.StatusBadRequest, resposta{
			Status:   "corpo_invalido",
			Mensagem: "Não entendi o corpo da requisição: esperava JSON com o campo topico.",
		})
	}

	topico := matarazzo.Topico(strings.ToLower(strings.TrimSpace(apiutil.Texto(corpo, "topico"))))
	if !matarazzo.TopicoValido(topico) {
		return c.JSON(http.StatusOK, resposta{
			Status:             "topico_invalido",
			TopicosDisponiveis: matarazzo.TopicosValidos(),
			Mensagem:           "Não tenho esse tópico no guia. Os tópicos disponíveis são: complexo, hotel_rosewood, restaurantes, mata_citta, lojas, eventos e como_chegar.",
		})
	}

	entrada := matarazzo.Guia[topico]
	c.Log().Info("ferramenta consultar_guia_matarazzo: consulta", "topico", string(topico))
	return c.JSON(http.StatusOK, resposta{
		Status:   "sucesso",
		Topico:   string(topico),
		Titulo:   entrada.Titulo,
		Texto:    entrada.Texto,
		Menu:     entrada.Menu,
		Horario:  entrada.Horario,
		Dicas:    entrada.Dicas,
		Mensagem: entrada.Titulo + ": " + entrada.Texto,
	})
}

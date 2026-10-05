// A ferramenta consultar_guia_matarazzo devolve ao concierge o conteúdo
// curado do guia do complexo — e o cardápio do Mata Città com busca por
// prato e listagem por categoria. Mesmo desenho das ferramentas da Wave:
// corpo tolerante, rate limit e resposta JSON pensada para ser falada.
package consultar_guia

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/emersonjoe/trilha"

	"github.com/emersonjoe/voice-based-customer-service/internal/apiutil"
	"github.com/emersonjoe/voice-based-customer-service/internal/limite"
	"github.com/emersonjoe/voice-based-customer-service/internal/matarazzo"
)

type resposta struct {
	Status             string                      `json:"status"`
	Topico             string                      `json:"topico,omitempty"`
	Titulo             string                      `json:"titulo,omitempty"`
	Texto              string                      `json:"texto,omitempty"`
	Menu               []string                    `json:"menu,omitempty"`
	Horario            string                      `json:"horario,omitempty"`
	Dicas              []string                    `json:"dicas,omitempty"`
	Resultados         []matarazzo.PratoEncontrado `json:"resultados,omitempty"`
	Categorias         []string                    `json:"categorias,omitempty"`
	Itens              []matarazzo.ItemPrato       `json:"itens,omitempty"`
	TopicosDisponiveis []string                    `json:"topicos_disponiveis,omitempty"`
	Mensagem           string                      `json:"mensagem"`
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
			Mensagem:           "Não tenho esse tópico no guia. Os tópicos disponíveis são: complexo, hotel_rosewood, restaurantes, le_jardin, blaise, taraz, rabo_di_galo, mata_citta, lojas, eventos e como_chegar.",
		})
	}

	// Cardápio do Mata Città: busca por prato ou listagem por categoria.
	if topico == matarazzo.TopicoMataCitta {
		if busca := strings.TrimSpace(apiutil.Texto(corpo, "busca")); busca != "" {
			return responderBusca(c, busca)
		}
		if categoria := strings.TrimSpace(apiutil.Texto(corpo, "categoria")); categoria != "" {
			return responderCategoria(c, categoria)
		}
		return responderMenuResumo(c)
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

// responderBusca devolve os pratos que casam com o termo, para o agente
// responder "vocês têm carbonara?" com precisão.
func responderBusca(c *trilha.Ctx, busca string) error {
	achados := matarazzo.BuscarNoCardapio(busca)
	if len(achados) == 0 {
		return c.JSON(http.StatusOK, resposta{
			Status:   "sem_resultados",
			Topico:   "mata_citta",
			Mensagem: "Não achei \"" + busca + "\" no cardápio do Mata Città. As categorias são: " + strings.Join(matarazzo.CategoriasCardapio(), ", ") + ".",
		})
	}
	nomes := make([]string, 0, len(achados))
	for _, a := range achados {
		nomes = append(nomes, a.Nome+" ("+a.Categoria+")")
	}
	c.Log().Info("ferramenta consultar_guia_matarazzo: busca no cardápio", "termo", busca, "achados", len(achados))
	return c.JSON(http.StatusOK, resposta{
		Status:     "sucesso",
		Topico:     "mata_citta",
		Resultados: achados,
		Categorias: matarazzo.CategoriasCardapio(),
		Mensagem:   "Achei " + plural(len(achados)) + " no cardápio do Mata Città: " + strings.Join(nomes, ", ") + ".",
	})
}

func responderCategoria(c *trilha.Ctx, categoria string) error {
	cat, ok := matarazzo.CategoriaPorNome(categoria)
	if !ok {
		return c.JSON(http.StatusOK, resposta{
			Status:     "categoria_invalida",
			Topico:     "mata_citta",
			Categorias: matarazzo.CategoriasCardapio(),
			Mensagem:   "Não tenho a categoria \"" + categoria + "\". As categorias do cardápio são: " + strings.Join(matarazzo.CategoriasCardapio(), ", ") + ".",
		})
	}
	nomes := make([]string, 0, len(cat.Itens))
	for _, item := range cat.Itens {
		nomes = append(nomes, item.Nome)
	}
	c.Log().Info("ferramenta consultar_guia_matarazzo: categoria", "categoria", cat.Nome)
	return c.JSON(http.StatusOK, resposta{
		Status:     "sucesso",
		Topico:     "mata_citta",
		Titulo:     "Mata Città — " + cat.Nome,
		Itens:      cat.Itens,
		Categorias: matarazzo.CategoriasCardapio(),
		Mensagem:   "Na categoria " + cat.Nome + " temos " + strconv.Itoa(len(cat.Itens)) + ": " + strings.Join(nomes, ", ") + ".",
	})
}

// responderMenuResumo devolve a visão geral: contagens por categoria, sem
// despejar os 54 itens na fala do agente.
func responderMenuResumo(c *trilha.Ctx) error {
	var resumo []string
	for _, cat := range matarazzo.CardapioMataCitta {
		resumo = append(resumo, cat.Nome+" ("+strconv.Itoa(len(cat.Itens))+")")
	}
	return c.JSON(http.StatusOK, resposta{
		Status:     "sucesso",
		Topico:     "mata_citta",
		Categorias: matarazzo.CategoriasCardapio(),
		Mensagem: "O cardápio do Mata Città tem oito categorias: " + strings.Join(resumo, ", ") +
			". Posso listar uma categoria ou buscar um prato pelo nome — pergunte o que o cliente quer.",
	})
}

func plural(n int) string {
	if n == 1 {
		return "1 prato"
	}
	return strconv.Itoa(n) + " pratos"
}

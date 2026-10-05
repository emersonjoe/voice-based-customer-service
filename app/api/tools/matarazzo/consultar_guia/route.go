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

	// Cardápios (Mata Città e LAVVA): busca por prato ou listagem por
	// categoria.
	var restauranteMenu string
	switch topico {
	case matarazzo.TopicoMataCitta:
		restauranteMenu = matarazzo.RestauranteMataCitta
	case matarazzo.TopicoLavva:
		restauranteMenu = matarazzo.RestauranteLavva
	}
	if restauranteMenu != "" {
		if busca := strings.TrimSpace(apiutil.Texto(corpo, "busca")); busca != "" {
			return responderBusca(c, restauranteMenu, busca)
		}
		if categoria := strings.TrimSpace(apiutil.Texto(corpo, "categoria")); categoria != "" {
			return responderCategoria(c, restauranteMenu, categoria)
		}
		return responderMenuResumo(c, restauranteMenu)
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
// responder "vocês têm carbonara?" e "vocês têm wagyu?" com precisão.
func responderBusca(c *trilha.Ctx, restaurante, busca string) error {
	achados := matarazzo.BuscarPratos(restaurante, busca)
	if len(achados) == 0 {
		return c.JSON(http.StatusOK, resposta{
			Status:     "sem_resultados",
			Topico:     restaurante,
			Categorias: matarazzo.CategoriasCardapio(restaurante),
			Mensagem:   "Não achei \"" + busca + "\" no cardápio do " + matarazzo.NomesDeExibicao[restaurante] + ". As categorias são: " + strings.Join(matarazzo.CategoriasCardapio(restaurante), ", ") + ".",
		})
	}
	nomes := make([]string, 0, len(achados))
	for _, a := range achados {
		nomes = append(nomes, a.Nome+" ("+a.Categoria+")")
	}
	c.Log().Info("ferramenta consultar_guia_matarazzo: busca no cardápio", "termo", busca, "achados", len(achados))
	return c.JSON(http.StatusOK, resposta{
		Status:     "sucesso",
		Topico:     restaurante,
		Resultados: achados,
		Categorias: matarazzo.CategoriasCardapio(restaurante),
		Mensagem:   "Achei " + plural(len(achados)) + " no cardápio do " + matarazzo.NomesDeExibicao[restaurante] + ": " + strings.Join(nomes, ", ") + ".",
	})
}

func responderCategoria(c *trilha.Ctx, restaurante, categoria string) error {
	cat, ok := matarazzo.CategoriaPorNome(restaurante, categoria)
	if !ok {
		return c.JSON(http.StatusOK, resposta{
			Status:     "categoria_invalida",
			Topico:     restaurante,
			Categorias: matarazzo.CategoriasCardapio(restaurante),
			Mensagem:   "Não tenho a categoria \"" + categoria + "\". As categorias do cardápio são: " + strings.Join(matarazzo.CategoriasCardapio(restaurante), ", ") + ".",
		})
	}
	nomes := make([]string, 0, len(cat.Itens))
	for _, item := range cat.Itens {
		nomes = append(nomes, item.Nome)
	}
	c.Log().Info("ferramenta consultar_guia_matarazzo: categoria", "restaurante", restaurante, "categoria", cat.Nome)
	return c.JSON(http.StatusOK, resposta{
		Status:     "sucesso",
		Topico:     restaurante,
		Titulo:     matarazzo.NomesDeExibicao[restaurante] + " — " + cat.Nome,
		Itens:      cat.Itens,
		Categorias: matarazzo.CategoriasCardapio(restaurante),
		Mensagem:   "Na categoria " + cat.Nome + " temos " + strconv.Itoa(len(cat.Itens)) + ": " + strings.Join(nomes, ", ") + ".",
	})
}

// responderMenuResumo devolve a visão geral: contagens por categoria, sem
// despejar os 54 itens na fala do agente.
func responderMenuResumo(c *trilha.Ctx, restaurante string) error {
	var resumo []string
	for _, cat := range matarazzo.Cardapios[restaurante] {
		resumo = append(resumo, cat.Nome+" ("+strconv.Itoa(len(cat.Itens))+")")
	}
	return c.JSON(http.StatusOK, resposta{
		Status:     "sucesso",
		Topico:     restaurante,
		Categorias: matarazzo.CategoriasCardapio(restaurante),
		Mensagem: "O cardápio do " + matarazzo.NomesDeExibicao[restaurante] + " tem estas categorias: " + strings.Join(resumo, ", ") +
			". Posso listar uma categoria ou buscar um prato pelo nome — pergunte o que o cliente quer.",
	})
}

func plural(n int) string {
	if n == 1 {
		return "1 prato"
	}
	return strconv.Itoa(n) + " pratos"
}

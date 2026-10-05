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
	Aproximado         bool                        `json:"aproximado,omitempty"`
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

	// O tópico também chega com fala imperfeita ("mata sita", "o
	// italiano"): tenta resolver antes de negar.
	topicoFalado := strings.ToLower(strings.TrimSpace(apiutil.Texto(corpo, "topico")))
	topico := matarazzo.Topico(topicoFalado)
	topicoAproximado := false
	if !matarazzo.TopicoValido(topico) {
		if resolvido, ok := matarazzo.ResolverTopico(topicoFalado); ok {
			topico, topicoAproximado = resolvido, true
		}
	}
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
			return responderBusca(c, restauranteMenu, busca, topicoAproximado)
		}
		if categoria := strings.TrimSpace(apiutil.Texto(corpo, "categoria")); categoria != "" {
			return responderCategoria(c, restauranteMenu, categoria, topicoAproximado)
		}
		return responderMenuResumo(c, restauranteMenu, topicoAproximado)
	}

	entrada := matarazzo.Guia[topico]
	c.Log().Info("ferramenta consultar_guia_matarazzo: consulta", "topico", string(topico))
	aviso := ""
	if topicoAproximado {
		aviso = "Interpretei como \"" + entrada.Titulo + "\" — confirme com o cliente. "
	}
	return c.JSON(http.StatusOK, resposta{
		Status:     "sucesso",
		Topico:     string(topico),
		Aproximado: topicoAproximado,
		Titulo:     entrada.Titulo,
		Texto:      entrada.Texto,
		Menu:       entrada.Menu,
		Horario:    entrada.Horario,
		Dicas:      entrada.Dicas,
		Mensagem:   aviso + entrada.Titulo + ": " + entrada.Texto,
	})
}

// responderBusca devolve os pratos que casam com o termo, para o agente
// responder "vocês têm carbonara?" e "vocês têm wagyu?" com precisão.
func responderBusca(c *trilha.Ctx, restaurante, busca string, topicoAproximado bool) error {
	// Ordem da busca: exato no restaurante → exato no outro → parecidos no
	// restaurante → parecidos no outro. Pedaços do nome e erros de fala
	// ("picana", "bife") ainda trazem os pratos mais parecidos.
	outro := ""
	if restaurante == matarazzo.RestauranteMataCitta {
		outro = matarazzo.RestauranteLavva
	} else if restaurante == matarazzo.RestauranteLavva {
		outro = matarazzo.RestauranteMataCitta
	}
	passos := []struct {
		lugar string
		aprox bool
	}{
		{restaurante, false},
		{outro, true},
		{restaurante, true},
		{outro, true},
	}
	for _, passo := range passos {
		if passo.lugar == "" {
			continue
		}
		achados := matarazzo.BuscarPratosTolerante(passo.lugar, busca)
		if len(achados) == 0 {
			continue
		}
		nomes := make([]string, 0, len(achados))
		for _, a := range achados {
			nomes = append(nomes, a.Nome+" ("+a.Categoria+")")
		}
		c.Log().Info("ferramenta consultar_guia_matarazzo: busca no cardápio", "termo", busca, "lugar", passo.lugar, "achados", len(achados))
		prefixo := ""
		if topicoAproximado {
			prefixo = "Interpretei como " + matarazzo.NomesDeExibicao[restaurante] + " — confirme com o cliente. "
		}
		var mensagem string
		switch {
		case passo.lugar == restaurante && !passo.aprox:
			mensagem = prefixo + "Achei " + plural(len(achados)) + " no cardápio do " + matarazzo.NomesDeExibicao[restaurante] +
				": " + strings.Join(nomes, ", ") + "."
		case passo.lugar != restaurante:
			mensagem = "Não achei \"" + busca + "\" no " + matarazzo.NomesDeExibicao[restaurante] +
				", mas o " + matarazzo.NomesDeExibicao[passo.lugar] + " tem: " + strings.Join(nomes, ", ") +
				". Confirme com o cliente se é isso."
		default:
			mensagem = "Os pratos mais parecidos com \"" + busca + "\" são " + strings.Join(nomes, ", ") +
				". Confirme com o cliente qual é."
		}
		return c.JSON(http.StatusOK, resposta{
			Status:     "sucesso",
			Topico:     passo.lugar,
			Aproximado: passo.aprox || topicoAproximado,
			Resultados: achados,
			Categorias: matarazzo.CategoriasCardapio(passo.lugar),
			Mensagem:   mensagem,
		})
	}
	return c.JSON(http.StatusOK, resposta{
		Status:     "sem_resultados",
		Topico:     restaurante,
		Categorias: matarazzo.CategoriasCardapio(restaurante),
		Mensagem:   "Não achei \"" + busca + "\" em nenhum cardápio. As categorias do " + matarazzo.NomesDeExibicao[restaurante] + " são: " + strings.Join(matarazzo.CategoriasCardapio(restaurante), ", ") + ".",
	})
}

func responderCategoria(c *trilha.Ctx, restaurante, categoria string, topicoAproximado bool) error {
	cat, ok := matarazzo.CategoriaMaisProxima(restaurante, categoria)
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
		Aproximado: topicoAproximado,
		Titulo:     matarazzo.NomesDeExibicao[restaurante] + " — " + cat.Nome,
		Itens:      cat.Itens,
		Categorias: matarazzo.CategoriasCardapio(restaurante),
		Mensagem:   prefixoTopico(topicoAproximado, restaurante) + "Na categoria " + cat.Nome + " temos " + strconv.Itoa(len(cat.Itens)) + ": " + strings.Join(nomes, ", ") + ".",
	})
}

// responderMenuResumo devolve a visão geral: contagens por categoria, sem
// despejar os 54 itens na fala do agente.
func responderMenuResumo(c *trilha.Ctx, restaurante string, topicoAproximado bool) error {
	var resumo []string
	for _, cat := range matarazzo.Cardapios[restaurante] {
		resumo = append(resumo, cat.Nome+" ("+strconv.Itoa(len(cat.Itens))+")")
	}
	return c.JSON(http.StatusOK, resposta{
		Status:     "sucesso",
		Topico:     restaurante,
		Aproximado: topicoAproximado,
		Categorias: matarazzo.CategoriasCardapio(restaurante),
		Mensagem: prefixoTopico(topicoAproximado, restaurante) + "O cardápio do " + matarazzo.NomesDeExibicao[restaurante] + " tem estas categorias: " + strings.Join(resumo, ", ") +
			". Posso listar uma categoria ou buscar um prato pelo nome — pergunte o que o cliente quer.",
	})
}

// prefixoTopico pede confirmação quando o tópico veio de uma fala imperfeita.
func prefixoTopico(aproximado bool, restaurante string) string {
	if aproximado {
		return "Interpretei como " + matarazzo.NomesDeExibicao[restaurante] + " — confirme com o cliente. "
	}
	return ""
}

func plural(n int) string {
	if n == 1 {
		return "1 prato"
	}
	return strconv.Itoa(n) + " pratos"
}

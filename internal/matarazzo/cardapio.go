// Cardápio do Mata Città, por categoria — o dado que a ferramenta
// consultar_guia usa para responder "vocês têm carbonara?" e listar
// categorias. Códigos são os do sistema do restaurante.
package matarazzo

import (
	"strings"
	"unicode"
)

// ItemPrato é um prato ou drink do cardápio.
type ItemPrato struct {
	Codigo int
	Nome   string
}

// CategoriaCardapio agrupa os itens do Mata Città.
type CategoriaCardapio struct {
	Nome  string
	Itens []ItemPrato
}

// CardapioMataCitta é o cardápio completo, por categoria.
var CardapioMataCitta = []CategoriaCardapio{
	{Nome: "Pizze", Itens: []ItemPrato{
		{1849, "Pizza 4 Formaggi"}, {1850, "Pizza Abobrinha"}, {1856, "Pizza Calabresa"},
		{1859, "Pizza Margherita"}, {1861, "Pizza Milho"}, {1870, "Pizzetta Burrata"},
		{8471, "Pizza Corniccioni"}, {8497, "Pizza Caprese della Matta"},
		{8498, "Calzone Mortadela e Pistache"}, {8499, "Pizza Pepperoni"}, {8500, "Pizza Carbonara"},
	}},
	{Nome: "Pane", Itens: []ItemPrato{
		{1990, "Focaccia Classica"}, {2867, "Ciabatta"},
	}},
	{Nome: "Pasta seca", Itens: []ItemPrato{
		{1351, "Lasagna de Ragù de Osso Buco"}, {1929, "Amatriciana"}, {1931, "Cacio e Pepe"},
		{1932, "Carbonara"}, {8478, "Linguini Camarão e Abobrinha"},
		{8479, "Spaghetti al Polvo Mediterrâneo"}, {8480, "Penne alla Caprese"},
		{8481, "Rigatoni alla Bolonhesa"},
	}},
	{Nome: "Pasta fresca", Itens: []ItemPrato{
		{1945, "Gnocchi al Tartufo"}, {1946, "Agnolotti dal Plin"}, {1948, "Lasagna della Casa"},
		{1950, "Ravioli Verde"}, {8482, "Rigatoni alla Arrabiata"},
		{8483, "Linguini Limone e Pistache"}, {8484, "Alla Genovese"},
	}},
	{Nome: "Risotti", Itens: []ItemPrato{
		{1952, "Risoto Frutti di Mare"}, {1954, "Risoto Funghi e Aspargos"},
	}},
	{Nome: "Secondi", Itens: []ItemPrato{
		{1959, "Bistecca"}, {1964, "Bife à Parmegiana"}, {1967, "Tagliata di Manzo"},
		{1969, "Polvo Grelhado"}, {8489, "Peixe Grelhado"},
		{8491, "Peito de Frango com Crosta de Ervas"}, {8494, "Ossobuco"},
	}},
	{Nome: "Sobremesa", Itens: []ItemPrato{
		{1993, "Tiramisu Classico"}, {1994, "Torta Caprese"},
		{1995, "Trio Cannoli Siciliani"}, {1996, "Semifreddo al Pistacchio"},
		{1997, "Panna Cotta alla Vaniglia"}, {1998, "Torta al Limone e Meringa"},
	}},
	{Nome: "Cocktail classic", Itens: []ItemPrato{
		{1021, "Aperol Spritz"}, {1037, "Caipiroska"}, {1114, "Fitzgerald"},
		{1766, "Mimosa"}, {1773, "Negroni"}, {2759, "Adonis"},
		{2760, "Drink Americano"}, {2761, "Bamboo"}, {2762, "Bees Knees"},
		{2763, "Bellini"}, {2764, "Between the Sheets"}, {2765, "Bloody Mary"},
	}},
}

// CategoriasCardapio devolve os nomes das categorias, na ordem do cardápio.
func CategoriasCardapio() []string {
	nomes := make([]string, 0, len(CardapioMataCitta))
	for _, c := range CardapioMataCitta {
		nomes = append(nomes, c.Nome)
	}
	return nomes
}

// CategoriaPorNome devolve a categoria pelo nome (sem diferenciar maiúsculas
// nem acentos) — "pasta seca", "Pizze" e "RISOTTI" funcionam.
func CategoriaPorNome(nome string) (CategoriaCardapio, bool) {
	alvo := Normaliza(nome)
	for _, c := range CardapioMataCitta {
		if Normaliza(c.Nome) == alvo {
			return c, true
		}
	}
	return CategoriaCardapio{}, false
}

// PratoEncontrado é um resultado de busca no cardápio.
type PratoEncontrado struct {
	Categoria string
	Codigo    int
	Nome      string
}

// BuscarNoCardapio procura pratos pelo nome (contém, sem diferenciar
// maiúsculas nem acentos): "carbonara" acha o da pasta e o da pizza.
func BuscarNoCardapio(termo string) []PratoEncontrado {
	alvo := Normaliza(termo)
	if alvo == "" {
		return nil
	}
	var achados []PratoEncontrado
	for _, c := range CardapioMataCitta {
		for _, item := range c.Itens {
			if strings.Contains(Normaliza(item.Nome), alvo) {
				achados = append(achados, PratoEncontrado{Categoria: c.Nome, Codigo: item.Codigo, Nome: item.Nome})
			}
		}
	}
	return achados
}

// Normaliza minúsculas e tira acentos, para a busca sobreviver à fala.
func Normaliza(s string) string {
	de, para := map[rune]rune{
		'á': 'a', 'à': 'a', 'ã': 'a', 'â': 'a', 'ä': 'a',
		'é': 'e', 'è': 'e', 'ê': 'e', 'ë': 'e',
		'í': 'i', 'ì': 'i', 'ï': 'i',
		'ó': 'o', 'ò': 'o', 'õ': 'o', 'ô': 'o', 'ö': 'o',
		'ú': 'u', 'ù': 'u', 'ü': 'u',
		'ç': 'c',
	}, []rune(strings.ToLower(strings.TrimSpace(s)))
	saida := make([]rune, 0, len(para))
	for _, r := range para {
		if substituido, trocou := de[r]; trocou {
			r = substituido
		}
		if unicode.IsLetter(r) || unicode.IsNumber(r) || r == ' ' {
			saida = append(saida, r)
		}
	}
	return string(saida)
}

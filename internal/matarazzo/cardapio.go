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

// Restaurantes que têm cardápio estruturado nesta POC.
const (
	RestauranteMataCitta = "mata_citta"
	RestauranteLavva     = "lavva"
)

// NomesDeExibicao liga o identificador ao nome de exibição.
var NomesDeExibicao = map[string]string{
	RestauranteMataCitta: "Mata Città",
	RestauranteLavva:     "LAVVA",
}

// Cardapios indexa os cardápios por restaurante.
var Cardapios = map[string][]CategoriaCardapio{}

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

// CategoriasCardapio devolve os nomes das categorias de um restaurante,
// na ordem do cardápio.
func CategoriasCardapio(restaurante string) []string {
	var nomes []string
	for _, c := range Cardapios[restaurante] {
		nomes = append(nomes, c.Nome)
	}
	return nomes
}

// CategoriaPorNome devolve a categoria pelo nome (sem diferenciar maiúsculas
// nem acentos) — "pasta seca", "Pizze" e "cortes de carne" funcionam.
func CategoriaPorNome(restaurante, nome string) (CategoriaCardapio, bool) {
	alvo := Normaliza(nome)
	for _, c := range Cardapios[restaurante] {
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

// BuscarPratos procura pratos pelo nome num restaurante (contém, sem
// diferenciar maiúsculas nem acentos): "carbonara" acha o da pasta e o da
// pizza; "wagyu" acha os cortes do LAVVA.
func BuscarPratos(restaurante, termo string) []PratoEncontrado {
	alvo := Normaliza(termo)
	if alvo == "" {
		return nil
	}
	var achados []PratoEncontrado
	for _, c := range Cardapios[restaurante] {
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

// CardapioLavva é o cardápio do LAVVA, o steakhouse coreano do complexo,
// por categoria.
var CardapioLavva = []CategoriaCardapio{
	{Nome: "Cortes de carne", Itens: []ItemPrato{
		{8143, "Bife de Ancho Angus"}, {8144, "Arroz de Brócolis"}, {8147, "Mil Folhas Batatas"},
		{8148, "Bibimbap"}, {8154, "Chorizo Wagyu"}, {8155, "Conjunto de Banchan (12 unidades, sendo 1 de cada)"},
		{8159, "Denver Steak Angus"}, {8160, "Farofa de Allium"}, {8163, "Filet Mignon Black Angus"},
		{8164, "Flat Iron Wagyu 300g"}, {8166, "Galbi (assado de tira marinado)"},
		{8172, "Japchae Macarrão de Batata Doce"}, {8174, "Pão de Alho"},
		{8178, "Picanha Angus"}, {8180, "Porterhouse Taurus Reserve"},
		{8395, "Chorizo Angus"}, {8396, "Bife de Ancho Wagyu"}, {8397, "Picanha Wagyu"},
		{8398, "Denver Steak Wagyu"}, {8399, "Rib Cap Wagyu"},
	}},
	{Nome: "Sobremesas", Itens: []ItemPrato{
		{8157, "Gergelim Negro"}, {8175, "Pavlova"}, {8176, "Milho Tostado"},
		{8177, "Mamuri"}, {8401, "Kokoneot Beluga"}, {8402, "Kokoneot Ossetra"},
		{8403, "Kokoneot Baeri"},
	}},
	{Nome: "Signature cocktails", Itens: []ItemPrato{
		{8025, "Amber Hearth"}, {8026, "Espresso Bori Cha"}, {8027, "Golden Sting"},
		{8028, "Goryeo Palace"}, {8029, "Honeydew"}, {8030, "Jeju Garden"},
		{8031, "Jirisan Collins"}, {8032, "K Pop"}, {8035, "Red Snapper"},
		{8036, "Margarita Gochujang"}, {8039, "Naju Haiboru"}, {8043, "Summer Bloom"},
	}},
	{Nome: "Classics", Itens: []ItemPrato{
		{1037, "Caipiroska"}, {1114, "Fitzgerald"}, {1773, "Negroni"},
		{2768, "Boulevardier"}, {2769, "Smoked Boulevardier"}, {2771, "Caipirinha"},
	}},
}

func init() {
	Cardapios[RestauranteMataCitta] = CardapioMataCitta
	Cardapios[RestauranteLavva] = CardapioLavva
}

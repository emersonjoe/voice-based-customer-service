// Package matarazzo guarda o guia curado do Complexo Cidade Matarazzo —
// hotel Rosewood São Paulo, restaurantes, lojas e eventos — que a
// ferramenta consultar_guia devolve ao concierge. Conteúdo factual
// (out/2026); atualize aqui quando o complexo mudar.
package matarazzo

import "slices"

// Topico identifica uma seção do guia.
type Topico string

// Tópicos válidos do guia.
const (
	TopicoComplexo     Topico = "complexo"
	TopicoHotel        Topico = "hotel_rosewood"
	TopicoRestaurantes Topico = "restaurantes"
	TopicoMataCitta    Topico = "mata_citta"
	TopicoLojas        Topico = "lojas"
	TopicoEventos      Topico = "eventos"
	TopicoComoChegar   Topico = "como_chegar"
)

// Entrada é uma seção do guia.
type Entrada struct {
	Titulo string
	Texto  string
	Dicas  []string
}

// Guia é o conteúdo falável do concierge.
var Guia = map[Topico]Entrada{
	TopicoComplexo: {
		Titulo: "Cidade Matarazzo",
		Texto: "A Cidade Matarazzo é um complexo de luxo na Bela Vista, colado na Avenida Paulista: cerca de trinta mil metros quadrados de edifícios " +
			"do antigo Hospital Matarazzo, de mil novecentos e quatro, restaurados e transformados em hotel, restaurantes, lojas e espaços de eventos. " +
			"É patrimônio histórico tombado, com jardins e arte por toda parte.",
		Dicas: []string{"Reserve meio dia para circular com calma", "A entrada principal fica pela Alameda Rio Claro"},
	},
	TopicoHotel: {
		Titulo: "Hotel Rosewood São Paulo",
		Texto: "O Rosewood São Paulo é o primeiro hotel da rede Rosewood na América do Sul, inaugurado em dois mil e vinte e dois. " +
			"São cerca de cento e sessenta suítes — inclusive duplex e uma cobertura de quatro quartos — numa torre assinada por Jean Nouvel, " +
			"com interiores de Philippe Starck e quase quinhentas obras de artistas brasileiros.",
		Dicas: []string{"Reservas pelo concierge (nossa ferramenta) ou pelo site oficial do Rosewood", "O café da manhã no Le Jardin é aberto a não hóspedes"},
	},
	TopicoRestaurantes: {
		Titulo: "Restaurantes do complexo",
		Texto: "Dentro do hotel: o Le Jardin, grand café aberto vinte e quatro horas com opções kosher e mesas no jardim; " +
			"o Blaise, brasserie francesa do chef Fernando Bouzan com ingredientes brasileiros; o Taraz, culinária sul-americana que aparece no Guia Michelin; " +
			"e o bar Rabo di Galo, com jazz ao vivo.",
		Dicas: []string{"Reserve com antecedência para fim de semana", "O Taraz costuma exigir reserva com alguns dias"},
	},
	TopicoMataCitta: {
		Titulo: "Mata Città",
		Texto: "O Mata Città é um espaço italiano de mil e seiscentos metros quadrados, com sete conceitos e cerca de quinhentos lugares, " +
			"inspirado no cinema italiano dos anos sessenta e setenta: restaurante, bar, sorveteria e brunch.",
		Dicas: []string{"Não aceita reserva online — chegue cedo ou deixe o nome na lista de espera no local", "Pratos a partir de valores acessíveis; bom para grupos"},
	},
	TopicoLojas: {
		Titulo: "Lojas",
		Texto:  "Butiques de grifes de luxo instaladas nos edifícios históricos restaurados do complexo, misturadas a espaços de decoração e arte.",
		Dicas:  []string{"Horários seguem o varejo de shopping; confira loja a loja", "Combine com um café no Le Jardin"},
	},
	TopicoEventos: {
		Titulo: "Eventos",
		Texto: "O complexo recebe casamentos, eventos corporativos e celebrações em espaços próprios, além de programação cultural aberta — " +
			"como o Festival Movimento Cidade, com música e artes nos jardins.",
		Dicas: []string{"Para eventos privados, o contato é pelo próprio Rosewood", "Acompanhe a agenda do complexo para eventos abertos"},
	},
	TopicoComoChegar: {
		Titulo: "Como chegar",
		Texto: "O endereço é Alameda Rio Claro, duzentos e sessenta, Bela Vista, São Paulo, a poucos minutos da Avenida Paulista. " +
			"De metrô, a estação Trianon-Masp fica a uma caminhada curta; de carro, há valet do hotel.",
		Dicas: []string{"Chegue com folga: o complexo é grande e as ruas internas são de pedestre", "Valet na entrada principal"},
	},
}

// RestaurantesReservaveis são os locais que a ferramenta criar_reserva
// aceita (o Mata Città fica de fora: não aceita reserva online).
var RestaurantesReservaveis = []string{
	"hotel_rosewood",
	"le_jardin",
	"blaise",
	"taraz",
	"rabo_di_galo",
}

// TopicoValido diz se o tópico existe no guia.
func TopicoValido(t Topico) bool {
	_, ok := Guia[t]
	return ok
}

// TopicosValidos lista os identificadores, para o erro da ferramenta.
func TopicosValidos() []string {
	fora := make([]string, 0, len(Guia))
	for t := range Guia {
		fora = append(fora, string(t))
	}
	return fora
}

// Rotulo devolve o nome de exibição de um local reservável.
func Rotulo(local string) string {
	rotulos := map[string]string{
		"hotel_rosewood": "Hotel Rosewood São Paulo",
		"le_jardin":      "Le Jardin",
		"blaise":         "Blaise",
		"taraz":          "Taraz",
		"rabo_di_galo":   "Rabo di Galo",
	}
	return rotulos[local]
}

// LocalReservavel diz se o local está na lista.
func LocalReservavel(local string) bool {
	return slices.Contains(RestaurantesReservaveis, local)
}

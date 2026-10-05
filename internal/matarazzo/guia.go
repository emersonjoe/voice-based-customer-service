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
	TopicoLavva        Topico = "lavva"
	TopicoLeJardin     Topico = "le_jardin"
	TopicoBlaise       Topico = "blaise"
	TopicoTaraz        Topico = "taraz"
	TopicoRaboDiGalo   Topico = "rabo_di_galo"
	TopicoMataCitta    Topico = "mata_citta"
	TopicoLojas        Topico = "lojas"
	TopicoEventos      Topico = "eventos"
	TopicoComoChegar   Topico = "como_chegar"
)

// Entrada é uma seção do guia.
type Entrada struct {
	Titulo  string
	Texto   string
	Dicas   []string
	Horario string
	Menu    []string // destaques do cardápio, quando houver
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
		Texto: "São três mundos: dentro do hotel Rosewood ficam o Le Jardin, o Blaise, o Taraz e o bar Rabo di Galo; " +
			"o Mata Città é o grande espaço italiano do complexo, com sete ambientes; e o LAVVA é o steakhouse coreano, " +
			"com cortes Angus e Wagyu na brasa. Pergunte qual deles o visitante quer conhecer em detalhe.",
		Dicas: []string{"Reserve com antecedência para fim de semana", "Posso detalhar qualquer um e já fazer a reserva"},
	},
	TopicoLavva: {
		Titulo:  "LAVVA — steakhouse coreano",
		Horario: "Almoço e jantar",
		Texto: "O LAVVA é o steakhouse coreano do complexo: cortes Angus e Wagyu na brasa (picanha, ancho, Denver, " +
			"Rib Cap, Galbi marinado), acompanhamentos da cozinha coreana como Bibimbap, Japchae e o conjunto de Banchan, " +
			"e coquetelaria autoral com toque coreano — da Margarita Gochujang ao Jeju Garden.",
		Menu: []string{
			"Cortes na brasa: Angus, Wagyu e Taurus Reserve",
			"Bibimbap, Japchae, Banchan e acompanhamentos",
			"Signature cocktails e clássicos",
			"Reservável pela nossa ferramenta",
		},
		Dicas: []string{"Bom para grupos: os cortes vêm para dividir no centro da mesa", "Reservável pela nossa ferramenta"},
	},
	TopicoLeJardin: {
		Titulo:  "Le Jardin — grand café do Rosewood (24 horas)",
		Horario: "Aberto 24 horas",
		Texto: "O grand café do hotel: do café da manhã ao lanche da madrugada, com opções kosher e mesas no jardim. " +
			"Culinária refinada em clima de conservatório — é a porta de entrada gastronômica do Rosewood.",
		Menu: []string{
			"Café da manhã e brunch servidos a qualquer hora",
			"Bife ancho e pratos de bistrô de hotel de luxo",
			"Polvo na grelha e sobremesas de confeitaria",
			"Opções kosher e mesa no jardim",
		},
		Dicas: []string{"Não precisa ser hóspede", "Reservável pela nossa ferramenta"},
	},
	TopicoBlaise: {
		Titulo:  "Blaise — brasserie francesa",
		Horario: "Jantar; menu degustação sazonal",
		Texto: "Brasserie de cozinha francesa assinada pelo chef Fernando Bouzan, com ingredientes brasileiros " +
			"no comando dos clássicos franceses. É o restaurante mais formal do Rosewood.",
		Menu: []string{
			"Atum selado com purê de pistache",
			"Vieiras salteadas com legumes",
			"Camarão com palmito pupunha",
			"Menu degustação sazonal do chef",
		},
		Dicas: []string{"Ideal para ocasiões especiais", "Reservável pela nossa ferramenta"},
	},
	TopicoTaraz: {
		Titulo:  "Taraz — sul-americano (Guia Michelin)",
		Horario: "Almoço e jantar; também atende no quarto",
		Texto: "Culinária sul-americana contemporânea que aparece no Guia Michelin, com cardápio sazonal. " +
			"É o único dos restaurantes que também serve o in-room dining do hotel.",
		Menu: []string{
			"Menu sazonal de raízes sul-americanas",
			"Peixes e cortes grelhados na brasa",
			"Coquetelaria com destilados regionais",
		},
		Dicas: []string{"Costuma exigir reserva com alguns dias", "Reservável pela nossa ferramenta"},
	},
	TopicoRaboDiGalo: {
		Titulo:  "Rabo di Galo — bar de jazz",
		Horario: "Fim de tarde até a noite, com pocket shows",
		Texto: "Bar intimista ao lado do Le Jardin, com poucas mesas e palco para pocket shows de jazz ao vivo. " +
			"Programa clássico: drink bem feito e música ao vivo.",
		Menu: []string{
			"Coquetelaria autoral e clássicos",
			"Petiscos de bar de hotel",
			"Programação de jazz e pocket shows",
		},
		Dicas: []string{"Mesas poucas: chegue cedo", "Reservável pela nossa ferramenta"},
	},
	TopicoMataCitta: {
		Titulo:  "Mata Città — Spettacolo Italiano",
		Horario: "Do café da manhã (8h) ao jantar, todos os dias",
		Texto: "Restaurante italiano de mil e seiscentos metros quadrados e cerca de quinhentos lugares, com sete ambientes " +
			"inspirados no cinema italiano dos anos sessenta e setenta — entre eles o Dolce Vita (pátio com café e brunch o dia todo), " +
			"o Capo (bar intimista), o Conde (homenagem a Francesco Matarazzo) e o Positano (costa italiana). " +
			"Comandado pelos chefs Felipe Rodrigues e Thiago Saldiva, a casa celebra a generosidade italiana: pratos fartos para dividir no centro da mesa.",
		Menu: []string{
			"Pizze, Pane, Pasta seca, Pasta fresca, Risotti, Secondi, Sobremesa e Cocktail classic",
			"Pratos a partir de R$ 35; cardápio completo em matacitta.help/cardapio",
			"Peça um prato pelo nome ('tem carbonara?') ou peça uma categoria ('me mostre as pizzas')",
		},
		Dicas: []string{"Não aceita reserva online — chegue cedo ou deixe o nome na lista de espera no local", "Bom para grupos e para ir com crianças"},
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
	"lavva",
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
		"lavva":          "LAVVA",
	}
	return rotulos[local]
}

// LocalReservavel diz se o local está na lista.
func LocalReservavel(local string) bool {
	return slices.Contains(RestaurantesReservaveis, local)
}

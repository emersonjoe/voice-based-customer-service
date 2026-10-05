// Resolvedor de nomes falados: liga o que o cliente disse (com dica ruim,
// pedaço faltando ou apelido) ao local ou tópico certo do guia.
package matarazzo

import "strings"

// aliasLocal mapeia apelidos falados ao identificador do local. A parte
// semântica da ferramenta: "o italiano" é o Mata Città, "o coreano" é o
// LAVVA, "o hotel" é o Rosewood.
var aliasLocal = map[string]string{
	"rosewood":   "hotel_rosewood",
	"hotel":      "hotel_rosewood",
	"o hotel":    "hotel_rosewood",
	"italiano":   "mata_citta",
	"matacita":   "mata_citta",
	"mata":       "mata_citta",
	"matt":       "mata_citta",
	"matta":      "mata_citta",
	"lavva":      "lavva",
	"lava":       "lavva",
	"coreano":    "lavva",
	"coreia":     "lavva",
	"steakhouse": "lavva",
}

// resolverPorAliases procura o falado nos apelidos primeiro (palavra
// inteira ou contido), devolvendo o identificador do local.
func resolverPorAliases(falado string) (string, bool) {
	alvo := Normaliza(falado)
	if alvo == "" {
		return "", false
	}
	for alias, id := range aliasLocal {
		aliasN := Normaliza(alias)
		if aliasN == alvo || strings.Contains(alvo, aliasN) || strings.Contains(aliasN, alvo) {
			return id, true
		}
	}
	return "", false
}

// LocalResolvido é o resultado de ResolverLocal.
type LocalResolvido struct {
	ID         string
	Nome       string
	Aproximado bool // true quando casou por aproximação — o agente confirma com o cliente
}

// resolverNomeTolerante casa o falado contra nomes de candidatos usando
// contém, tokens e Levenshtein.
func resolverNomeTolerante(falado string, candidatos map[string]string) (string, bool, bool) {
	alvo := Normaliza(falado)
	if alvo == "" {
		return "", false, false
	}
	melhor, melhorPontos := "", 0.0
	for id, nome := range candidatos {
		nomeN := Normaliza(nome)
		idN := Normaliza(id)
		switch {
		case nomeN == alvo || idN == alvo:
			return id, false, true
		case strings.Contains(nomeN, alvo) || strings.Contains(alvo, nomeN):
			pontos := float64(len(nomeN)) / float64(len(alvo))
			if pontos > melhorPontos {
				melhor, melhorPontos = id, pontos
			}
		}
	}
	if melhor != "" {
		return melhor, true, true
	}
	falados := tokens(alvo)
	for id, nome := range candidatos {
		if pontuaNome(falados, tokens(Normaliza(nome))) >= 0.6 {
			return id, true, true
		}
	}
	return "", false, false
}

// ResolverLocal resolve o local falado para um dos locais conhecidos
// (reserváveis + Mata Città, que não é reservável mas é consultável).
// Aproximado=true pede confirmação ao cliente.
func ResolverLocal(falado string) (LocalResolvido, bool) {
	candidatos := map[string]string{}
	for _, id := range RestaurantesReservaveis {
		candidatos[id] = Rotulo(id)
	}
	candidatos[RestauranteMataCitta] = "Mata Città"

	if id, ok := resolverPorAliases(falado); ok {
		nome := candidatos[id]
		if nome == "" {
			nome = id
		}
		return LocalResolvido{ID: id, Nome: nome, Aproximado: true}, true
	}
	if id, aproximado, ok := resolverNomeTolerante(falado, candidatos); ok {
		return LocalResolvido{ID: id, Nome: candidatos[id], Aproximado: aproximado}, true
	}
	return LocalResolvido{}, false
}

// ResolverTopico resolve o tópico do guia falado (com apelidos e erros).
func ResolverTopico(falado string) (Topico, bool) {
	if TopicoValido(Topico(Normaliza(falado))) {
		return Topico(Normaliza(falado)), true
	}
	if id, ok := resolverPorAliases(falado); ok {
		if TopicoValido(Topico(id)) {
			return Topico(id), true
		}
	}
	candidatos := map[string]string{}
	for topico, entrada := range Guia {
		candidatos[string(topico)] = entrada.Titulo
	}
	if id, _, ok := resolverNomeTolerante(falado, candidatos); ok {
		return Topico(id), true
	}
	return "", false
}

// CategoriaMaisProxima acha a categoria cujo nome mais se aproxima do
// falado (contém ou tokens tolerantes) — para a ferramenta não negar
// "pizas" quando a categoria é "Pizze".
func CategoriaMaisProxima(restaurante, falado string) (CategoriaCardapio, bool) {
	if cat, ok := CategoriaPorNome(restaurante, falado); ok {
		return cat, true
	}
	alvo := Normaliza(falado)
	if alvo == "" {
		return CategoriaCardapio{}, false
	}
	falados := tokens(alvo)
	for _, cat := range Cardapios[restaurante] {
		if pontuaNome(falados, tokens(Normaliza(cat.Nome))) >= 0.6 {
			return cat, true
		}
		if strings.Contains(Normaliza(cat.Nome), alvo) || strings.Contains(alvo, Normaliza(cat.Nome)) {
			return cat, true
		}
	}
	return CategoriaCardapio{}, false
}

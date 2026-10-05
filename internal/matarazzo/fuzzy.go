// Camada de tolerância para a fala: a transcrição chega com dicas erradas
// ("picana"), pedaços faltando ("bife") e apelidos ("o italiano", "lava").
// A ordem de tentativa das buscas é: igualdade/contém → tokens casados
// (com Levenshtein por token) → sinônimos/aliases.
package matarazzo

import (
	"slices"
	"strings"
)

// Levenshtein devolve a distância de edição entre duas palavras
// normalizadas — quantas letras mudam para uma virar a outra.
func Levenshtein(a, b string) int {
	if a == b {
		return 0
	}
	if len(a) == 0 {
		return len(b)
	}
	if len(b) == 0 {
		return len(a)
	}
	anterior := make([]int, len(b)+1)
	atual := make([]int, len(b)+1)
	for j := 0; j <= len(b); j++ {
		anterior[j] = j
	}
	for i := 1; i <= len(a); i++ {
		atual[0] = i
		for j := 1; j <= len(b); j++ {
			custo := 1
			if a[i-1] == b[j-1] {
				custo = 0
			}
			atual[j] = min(anterior[j]+1, atual[j-1]+1, anterior[j-1]+custo)
		}
		anterior, atual = atual, anterior
	}
	return anterior[len(b)]
}

// limiarDeToken aceita erro de dica proporcional ao tamanho da palavra:
// curto tolera 1 letra, longo 2 ou 3.
func limiarDeToken(palavra string) int {
	switch tamanho := len(palavra); {
	case tamanho <= 4:
		return 1
	case tamanho <= 8:
		return 2
	default:
		return 3
	}
}

// tokens separa um texto normalizado em palavras.
func tokens(s string) []string {
	return strings.Fields(s)
}

// casaToken diz se dois tokens casam: contido (quase prefixo, para caber
// "bife" dentro de "bife ancho") ou com erro de digitação/fala dentro do
// limiar.
func casaToken(falado, doNome string) bool {
	if falado == "" || doNome == "" {
		return false
	}
	if strings.Contains(doNome, falado) || strings.Contains(falado, doNome) {
		// contém só vale se for início da palavra: "ita" em "italiano" sim,
		// "tta" no meio de "burrata" não.
		if strings.HasPrefix(doNome, falado) || strings.HasPrefix(falado, doNome) {
			return true
		}
	}
	return Levenshtein(falado, doNome) <= limiarDeToken(doNome)
}

// pontuaNome mede quantos tokens falados casam com o nome: 1.0 = todos.
func pontuaNome(falados, doNome []string) float64 {
	if len(falados) == 0 {
		return 0
	}
	casados := 0
	for _, f := range falados {
		for _, n := range doNome {
			if casaToken(f, n) {
				casados++
				break
			}
		}
	}
	return float64(casados) / float64(len(falados))
}

// BuscarPratosTolerante procura pratos aceitando pedaços do nome e erros de
// fala ("picana" → Picanha, "bife" → os bifes, "tiramusu" → Tiramisu).
// Devolve no máximo 8, do mais parecido ao menos.
func BuscarPratosTolerante(restaurante, termo string) []PratoEncontrado {
	alvo := Normaliza(termo)
	if alvo == "" {
		return nil
	}
	// 1) contém exato (comportamento antigo, melhor qualidade)
	if diretos := BuscarPratos(restaurante, alvo); len(diretos) > 0 {
		return diretos
	}
	// 2) tolerante por tokens
	falados := tokens(alvo)
	type candidato struct {
		prato  PratoEncontrado
		pontos float64
	}
	var candidatos []candidato
	vistos := map[int]bool{}
	for _, c := range Cardapios[restaurante] {
		for _, item := range c.Itens {
			chave := item.Codigo
			if vistos[chave] {
				continue
			}
			pontos := pontuaNome(falados, tokens(Normaliza(item.Nome)))
			if pontos >= 0.5 {
				vistos[chave] = true
				candidatos = append(candidatos, candidato{
					prato:  PratoEncontrado{Categoria: c.Nome, Codigo: item.Codigo, Nome: item.Nome},
					pontos: pontos,
				})
			}
		}
	}
	slices.SortStableFunc(candidatos, func(a, b candidato) int {
		if a.pontos > b.pontos {
			return -1
		}
		if a.pontos < b.pontos {
			return 1
		}
		return strings.Compare(a.prato.Nome, b.prato.Nome)
	})
	if len(candidatos) > 8 {
		candidatos = candidatos[:8]
	}
	achados := make([]PratoEncontrado, 0, len(candidatos))
	for _, c := range candidatos {
		achados = append(achados, c.prato)
	}
	return achados
}

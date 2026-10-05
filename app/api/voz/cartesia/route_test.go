package cartesia

import (
	"strings"
	"testing"
)

func TestAgenteValido(t *testing.T) {
	ok := []string{
		"abc123xy",
		"agent_2901m32wxnddfaat2bzqtsbrn6zd",
		"agent_cFk5fF9mx1QLvw1FGQ4h8r",
		strings.Repeat("a", 80),
	}
	for _, agente := range ok {
		if !agenteValido(agente) {
			t.Errorf("agenteValido(%q) = false; queria true", encurta(agente))
		}
	}
	falha := []string{
		"",
		"ab",
		"a b",
		"a.b",
		"a/b",
		"agente\n",
		strings.Repeat("a", 81),
	}
	for _, agente := range falha {
		if agenteValido(agente) {
			t.Errorf("agenteValido(%q) = true; queria false", encurta(agente))
		}
	}
}

func encurta(s string) string {
	if len(s) > 40 {
		return s[:37] + "…"
	}
	return s
}

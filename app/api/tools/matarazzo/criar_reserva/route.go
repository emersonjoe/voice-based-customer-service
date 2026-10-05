// A ferramenta criar_reserva_matarazzo registra a reserva do hotel
// Rosewood ou de um dos restaurantes do complexo. É um atendimento do tipo
// reserva (protocolo RS-…), sempre no agente matarazzo.
package criar_reserva

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/emersonjoe/trilha"

	"github.com/emersonjoe/voice-based-customer-service/internal/apiutil"
	"github.com/emersonjoe/voice-based-customer-service/internal/atendimento"
	"github.com/emersonjoe/voice-based-customer-service/internal/limite"
	"github.com/emersonjoe/voice-based-customer-service/internal/matarazzo"
)

type resposta struct {
	Status    string `json:"status"`
	Protocolo string `json:"protocolo,omitempty"`
	Agente    string `json:"agente,omitempty"`
	Local     string `json:"local,omitempty"`
	Mensagem  string `json:"mensagem"`
}

func falha(status, mensagem string) resposta { return resposta{Status: status, Mensagem: mensagem} }

func POST(c *trilha.Ctx) error {
	w, r := c.Writer(), c.Request()
	if !apiutil.Limita(w, r, trilha.Use[*limite.Limitador](c), "reserva_matarazzo") {
		return nil
	}
	corpo, err := apiutil.Corpo(r)
	if err != nil {
		return c.JSON(http.StatusBadRequest, falha("corpo_invalido",
			"Não entendi o corpo da requisição: esperava JSON com tipo, data, horario, pessoas e nome_hospede."))
	}

	local := strings.ToLower(strings.TrimSpace(apiutil.Texto(corpo, "tipo")))
	nome := strings.TrimSpace(apiutil.Texto(corpo, "nome_hospede"))
	data := strings.TrimSpace(apiutil.Texto(corpo, "data"))
	horario := strings.TrimSpace(apiutil.Texto(corpo, "horario"))
	pessoas, errPessoas := strconv.Atoi(apiutil.Texto(corpo, "pessoas"))

	switch {
	case !matarazzo.LocalReservavel(local):
		switch local {
		case "mata_citta":
			return c.JSON(http.StatusOK, falha("sem_reserva_online",
				"O Mata Città não aceita reserva online — a entrada é por lista de espera no local. Sugira chegar cedo ou deixar o nome no balcão."))
		case "gui":
			return c.JSON(http.StatusOK, falha("tipo_invalido",
				"Para reservas, use os locais da lista: hotel_rosewood, le_jardin, blaise, taraz, rabo_di_galo e lavva."))
		}
		return c.JSON(http.StatusOK, falha("tipo_invalido",
			"Os locais reserváveis são: hotel_rosewood, le_jardin, blaise, taraz, rabo_di_galo e lavva."))
	case len(nome) < 3 || len(nome) > 120:
		return c.JSON(http.StatusOK, falha("nome_invalido",
			"Preciso do nome completo do hóspede para a reserva."))
	case !dataValida(data):
		return c.JSON(http.StatusOK, falha("data_invalida",
			"A data precisa estar no formato AAAA-MM-DD e ser futura, até dezoito meses de hoje."))
	case !horarioValido(horario):
		return c.JSON(http.StatusOK, falha("horario_invalido",
			"O horário precisa estar no formato HH:MM, entre 06:00 e 23:59."))
	case errPessoas != nil || pessoas < 1 || pessoas > 12:
		return c.JSON(http.StatusOK, falha("pessoas_invalido",
			"O número de pessoas precisa ser de um a doze."))
	}

	quando, _ := time.Parse("2006-01-02", data)
	detalhes := map[string]string{
		"local":   matarazzo.Rotulo(local),
		"tipo":    local,
		"data":    data,
		"horario": horario,
		"pessoas": strconv.Itoa(pessoas),
	}
	if tel := apiutil.Texto(corpo, "telefone"); tel != "" {
		detalhes["telefone"] = tel
	}
	if obs := strings.TrimSpace(apiutil.Texto(corpo, "observacoes")); obs != "" {
		detalhes["observacoes"] = obs
	}

	store := trilha.Use[*atendimento.Store](c)
	rotulo := matarazzo.Rotulo(local)
	resumo := "Reserva: " + rotulo + " — " + quando.Format("02/01") + " " + horario + " (" + strconv.Itoa(pessoas) + " pessoas)"
	att := store.Criar(atendimento.TipoReserva, atendimento.AgenteMatarazzo, "", nome, resumo,
		"Reserva feita pelo concierge de voz da Cidade Matarazzo.",
		atendimento.PrioridadeBaixa, origem(r),
		apiutil.Texto(corpo, "conversation_id"), detalhes)

	c.Log().Info("ferramenta criar_reserva_matarazzo: reserva registrada", "protocolo", att.Protocolo)
	return c.JSON(http.StatusOK, resposta{
		Status:    "solicitacao_registrada",
		Protocolo: att.Protocolo,
		Agente:    atendimento.AgenteMatarazzo,
		Local:     rotulo,
		Mensagem: "Reserva no " + rotulo + " registrada para " + quando.Format("02/01/2006") +
			" às " + horario + ", para " + strconv.Itoa(pessoas) + " pessoa(s), no nome de " + nome +
			". O protocolo é " + att.Protocolo + ".",
	})
}

// origem distingue a chamada do browser (voz: fetch manda Sec-Fetch-Mode)
// da da Cartesia (webhook, que é um servidor e não manda esses headers).
func origem(r *http.Request) string {
	if r.Header.Get("Sec-Fetch-Mode") != "" {
		return "voz"
	}
	return "webhook"
}

func dataValida(s string) bool {
	quando, err := time.Parse("2006-01-02", s)
	if err != nil {
		return false
	}
	hoje := time.Now().Truncate(24 * time.Hour)
	return !quando.Before(hoje) && quando.Before(hoje.AddDate(1, 6, 0))
}

func horarioValido(s string) bool {
	h, err := time.Parse("15:04", s)
	if err != nil {
		return false
	}
	abertura, _ := time.Parse("15:04", "06:00")
	fechamento, _ := time.Parse("15:04", "23:59")
	return !h.Before(abertura) && !h.After(fechamento)
}

// A ferramenta enviar_segunda_via_boleto localiza o cliente pelo CPF,
// pega a fatura em aberto e registra o envio da segunda via pelo canal
// pedido. Serve ao client tool do browser e ao webhook da ElevenLabs.
package enviar_segunda_via_boleto

import (
	"net/http"
	"strings"
	"time"

	"github.com/emersonjoe/trilha"

	"github.com/emersonjoe/voice-based-customer-service/internal/apiutil"
	"github.com/emersonjoe/voice-based-customer-service/internal/atendimento"
	"github.com/emersonjoe/voice-based-customer-service/internal/limite"
)

type resposta struct {
	Status         string `json:"status"`
	Protocolo      string `json:"protocolo,omitempty"`
	Cliente        string `json:"cliente,omitempty"`
	Mensagem       string `json:"mensagem"`
	Valor          string `json:"valor,omitempty"`
	Vencimento     string `json:"vencimento,omitempty"`
	LinhaDigitavel string `json:"linha_digitavel,omitempty"`
	EnviadoPara    string `json:"enviado_para,omitempty"`
}

func falha(status, mensagem string) resposta { return resposta{Status: status, Mensagem: mensagem} }

func POST(c *trilha.Ctx) error {
	w, r := c.Writer(), c.Request()
	if !apiutil.Limita(w, r, trilha.Use[*limite.Limitador](c), "segunda_via") {
		return nil
	}
	corpo, err := apiutil.Corpo(r)
	if err != nil {
		return c.JSON(http.StatusBadRequest, falha("corpo_invalido",
			"Não entendi o corpo da requisição: esperava JSON com o cpf do cliente."))
	}

	cpf := atendimento.SoDigitos(apiutil.Texto(corpo, "cpf"))
	if !atendimento.CPFValido(cpf) {
		return c.JSON(http.StatusOK, falha("cpf_invalido",
			"Para localizar o cadastro preciso do CPF, com os onze números."))
	}

	store := trilha.Use[*atendimento.Store](c)
	cliente, ok := store.ClientePorCPF(cpf)
	if !ok {
		return c.JSON(http.StatusOK, falha("cliente_nao_encontrado",
			"Não localizei um cadastro com esse CPF. Peça para conferir os números e tente de novo."))
	}

	fatura, ok := cliente.FaturaAberta(time.Now())
	if !ok {
		return c.JSON(http.StatusOK, resposta{
			Status:   "sem_fatura_aberta",
			Cliente:  cliente.Nome,
			Mensagem: "Cadastro de " + cliente.Nome + " localizado. Boa notícia: não há nenhuma fatura em aberto, então não é preciso segunda via.",
		})
	}

	// O e-mail do cadastro é o destino padrão; um e-mail informado na
	// conversa substitui, desde que pareça um e-mail.
	destino := cliente.Email
	if pedido := strings.TrimSpace(apiutil.Texto(corpo, "email")); strings.Contains(pedido, "@") && len(pedido) <= 120 {
		destino = pedido
	}
	canal := strings.ToLower(apiutil.Texto(corpo, "canal"))
	if canal == "" || (canal != "email" && canal != "whatsapp") {
		canal = "email"
	}
	if canal == "whatsapp" {
		destino = cliente.Telefone
	}

	detalhes := map[string]string{
		"canal":        canal,
		"enviado_para": destino,
		"valor":        atendimento.BRL(fatura.ValorCentavos),
		"vencimento":   fatura.Vencimento.Format("02/01/2006"),
		"linha":        fatura.LinhaDigitavel,
	}
	att := store.Criar(atendimento.TipoSegundaVia, cpf,
		"Segunda via enviada por "+canal,
		"Segunda via da fatura de "+fatura.Vencimento.Format("01/2006")+" enviada pelo agente de voz.",
		atendimento.PrioridadeBaixa, origem(r),
		apiutil.Texto(corpo, "conversation_id"), detalhes)

	c.Log().Info("ferramenta enviar_segunda_via_boleto: envio registrado", "protocolo", att.Protocolo)
	venc := fatura.Vencimento.Format("02/01/2006")
	mensagem := "Cadastro de " + cliente.Nome + " localizado. Segunda via enviada para " + destino +
		". O valor é " + atendimento.BRL(fatura.ValorCentavos) +
		", com vencimento em " + venc + ". O protocolo do atendimento é " + att.Protocolo + "."
	return c.JSON(http.StatusOK, resposta{
		Status:         "sucesso",
		Protocolo:      att.Protocolo,
		Cliente:        cliente.Nome,
		Valor:          atendimento.BRL(fatura.ValorCentavos),
		Vencimento:     venc,
		LinhaDigitavel: fatura.LinhaDigitavel,
		EnviadoPara:    destino,
		Mensagem:       mensagem,
	})
}

// origem distingue a chamada do browser (voz: fetch manda Sec-Fetch-Mode)
// da da ElevenLabs (webhook, que é um servidor e não manda esses headers).
func origem(r *http.Request) string {
	if r.Header.Get("Sec-Fetch-Mode") != "" {
		return "voz"
	}
	return "webhook"
}

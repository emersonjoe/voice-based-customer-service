// A ferramenta solicitar_religue_confirmacao registra o pedido de religue
// depois que o cliente confirma (ou comprova) o pagamento da fatura vencida.
// Serve ao client tool do browser e ao webhook da ElevenLabs.
package solicitar_religue_confirmacao

import (
	"net/http"
	"time"

	"github.com/emersonjoe/trilha"

	"github.com/emersonjoe/voice-based-customer-service/internal/apiutil"
	"github.com/emersonjoe/voice-based-customer-service/internal/atendimento"
	"github.com/emersonjoe/voice-based-customer-service/internal/limite"
)

type resposta struct {
	Status    string `json:"status"`
	Protocolo string `json:"protocolo,omitempty"`
	Cliente   string `json:"cliente,omitempty"`
	Cpf       string `json:"cpf,omitempty"`
	Mensagem  string `json:"mensagem"`
	Prazo     string `json:"prazo,omitempty"`
}

func falha(status, mensagem string) resposta { return resposta{Status: status, Mensagem: mensagem} }

func POST(c *trilha.Ctx) error {
	w, r := c.Writer(), c.Request()
	if !apiutil.Limita(w, r, trilha.Use[*limite.Limitador](c), "religue") {
		return nil
	}
	corpo, err := apiutil.Corpo(r)
	if err != nil {
		return c.JSON(http.StatusBadRequest, falha("corpo_invalido",
			"Não entendi o corpo da requisição: esperava JSON com o cpf e a confirmação do pagamento."))
	}

	// A voz às vezes derrapa um dígito do CPF: tenta reparar contra o
	// cadastro antes de negar.
	store := trilha.Use[*atendimento.Store](c)
	cpf, cpfReparado := apiutil.CPFDaFerramenta(corpo, store)
	if !atendimento.CPFValido(cpf) {
		recebido := atendimento.SoDigitos(apiutil.Texto(corpo, "cpf"))
		if recebido == "" {
			return c.JSON(http.StatusOK, falha("cpf_invalido",
				"Para localizar o cadastro preciso do CPF, com os onze números."))
		}
		return c.JSON(http.StatusOK, falha("cpf_invalido",
			"O CPF que recebi ("+recebido+") não é válido. Confira os onze números com o cliente, falando devagar, um por um."))
	}

	// Sem confirmação de pagamento não há religue: a resposta diz o que
	// falta, e o agente repassa ao cliente.
	if !apiutil.Bool(corpo, "pagamento_confirmado") {
		return c.JSON(http.StatusOK, falha("comprovacao_pendente",
			"Para religar o sinal, o pagamento precisa estar confirmado. Peça o comprovante ou a confirmação "+
				"do pagamento e chame esta ferramenta de novo com pagamento_confirmado=true."))
	}

	cliente, ok := store.ClientePorCPF(cpf)
	if !ok {
		return c.JSON(http.StatusOK, falha("cliente_nao_encontrado",
			"Não localizei um cadastro com esse CPF. Peça para conferir os números e tente de novo."))
	}
	aviso := ""
	if cpfReparado {
		aviso = "Corrigi o CPF para " + atendimento.FormataCPF(cpf) + " contra o cadastro — confirme com o cliente. "
	}

	fatura, temFatura := cliente.FaturaAberta(time.Now())
	if !temFatura {
		return c.JSON(http.StatusOK, resposta{
			Status:   "sem_debitos",
			Cliente:  cliente.Nome,
			Mensagem: "Cadastro de " + cliente.Nome + " localizado, sem faturas vencidas. Se o sinal continua fora, o caminho é abrir um chamado técnico em vez de um religue.",
		})
	}

	detalhes := map[string]string{
		"forma_pagamento": apiutil.Texto(corpo, "forma_pagamento"),
		"data_pagamento":  apiutil.Texto(corpo, "data_pagamento"),
		"comprovante":     apiutil.Texto(corpo, "comprovante"),
		"valor":           atendimento.BRL(fatura.ValorCentavos),
		"vencimento":      fatura.Vencimento.Format("02/01/2006"),
	}
	att := store.Criar(atendimento.TipoReligue, cpf, "",
		"Religue solicitado com comprovação de pagamento",
		"Cliente confirmou o pagamento da fatura vencida e pediu o religue do sinal.",
		atendimento.PrioridadeAlta, origem(r),
		apiutil.Texto(corpo, "conversation_id"), detalhes)

	c.Log().Info("ferramenta solicitar_religue_confirmacao: pedido registrado", "protocolo", att.Protocolo)
	return c.JSON(http.StatusOK, resposta{
		Status:    "solicitacao_registrada",
		Protocolo: att.Protocolo,
		Cliente:   cliente.Nome,
		Cpf:       atendimento.FormataCPF(cpf),
		Prazo:     "até 2 horas úteis",
		Mensagem:  aviso + "Cadastro de " + cliente.Nome + " localizado. Pedido de religue registrado no protocolo " + att.Protocolo + ". A equipe técnica religa o sinal em até 2 horas úteis.",
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

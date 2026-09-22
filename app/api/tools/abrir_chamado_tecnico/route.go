// A ferramenta abrir_chamado_tecnico registra o chamado técnico que o
// agente de voz abre depois de entender o problema do cliente. O mesmo
// endpoint atende o client tool do browser e o webhook da ElevenLabs.
package abrir_chamado_tecnico

import (
	"net/http"
	"strings"

	"github.com/emersonjoe/trilha"

	"github.com/emersonjoe/voice-based-customer-service/internal/apiutil"
	"github.com/emersonjoe/voice-based-customer-service/internal/atendimento"
	"github.com/emersonjoe/voice-based-customer-service/internal/limite"
)

type resposta struct {
	Status     string `json:"status"`
	Protocolo  string `json:"protocolo,omitempty"`
	Cliente    string `json:"cliente,omitempty"`
	Cpf        string `json:"cpf,omitempty"`
	Mensagem   string `json:"mensagem"`
	Prazo      string `json:"prazo,omitempty"`
	Prioridade string `json:"prioridade,omitempty"`
}

func falha(status, mensagem string) resposta { return resposta{Status: status, Mensagem: mensagem} }

func POST(c *trilha.Ctx) error {
	w, r := c.Writer(), c.Request()
	if !apiutil.Limita(w, r, trilha.Use[*limite.Limitador](c), "abrir_chamado") {
		return nil
	}
	corpo, err := apiutil.Corpo(r)
	if err != nil {
		return c.JSON(http.StatusBadRequest, falha("corpo_invalido",
			"Não entendi o corpo da requisição: esperava JSON com nome, cpf e descricao."))
	}

	nome := strings.TrimSpace(apiutil.Texto(corpo, "nome"))
	descricao := strings.TrimSpace(apiutil.Texto(corpo, "descricao"))
	problema := strings.ToLower(apiutil.Texto(corpo, "problema"))
	prioridade := atendimento.Prioridade(strings.ToLower(apiutil.Texto(corpo, "prioridade")))

	// A voz às vezes derrapa um dígito do CPF: tenta reparar contra o
	// cadastro antes de negar.
	store := trilha.Use[*atendimento.Store](c)
	cpf, cpfReparado := apiutil.CPFDaFerramenta(corpo, store)
	if !atendimento.CPFValido(cpf) {
		recebido := atendimento.SoDigitos(apiutil.Texto(corpo, "cpf"))
		if recebido == "" {
			return c.JSON(http.StatusOK, falha("cpf_invalido",
				"Preciso do CPF do cliente, com os onze números."))
		}
		return c.JSON(http.StatusOK, falha("cpf_invalido",
			"O CPF que recebi ("+recebido+") não é válido. Confira os onze números com o cliente, falando devagar, um por um."))
	}
	if descricao == "" || len(descricao) > 1000 {
		return c.JSON(http.StatusOK, falha("descricao_invalida",
			"Preciso de uma descrição do problema, com até mil caracteres."))
	}

	// O nome é opcional quando o CPF está no cadastro: a base sabe quem é o
	// cliente, e a conversa não precisa cobrar o dado duas vezes.
	if len(nome) < 3 || len(nome) > 120 {
		if cliente, ok := store.ClientePorCPF(cpf); ok {
			nome = cliente.Nome
		}
	}
	if len(nome) < 3 || len(nome) > 120 {
		return c.JSON(http.StatusOK, falha("nome_invalido",
			"Não consegui identificar o cliente: o CPF não está no cadastro e falta o nome completo. "+
				"Peça o nome completo do cliente e tente de novo."))
	}

	if prioridade == "" {
		prioridade = atendimento.PrioridadeMedia
	}
	if !prioridade.Valido() {
		prioridade = atendimento.PrioridadeMedia
	}

	detalhes := map[string]string{"problema": problema}
	if tel := apiutil.Texto(corpo, "telefone"); tel != "" {
		detalhes["telefone"] = tel
	}
	if email := apiutil.Texto(corpo, "email"); strings.Contains(email, "@") {
		detalhes["email"] = email
	}

	att := store.Criar(atendimento.TipoChamadoTecnico, cpf, nome, resumoDo(problema, descricao),
		descricao, prioridade, origem(r),
		apiutil.Texto(corpo, "conversation_id"), detalhes)

	c.Log().Info("ferramenta abrir_chamado_tecnico: protocolo criado", "protocolo", att.Protocolo)
	aviso := ""
	if cpfReparado {
		aviso = "Corrigi o CPF para " + atendimento.FormataCPF(cpf) + " contra o cadastro — confirme com o cliente. "
	}
	return c.JSON(http.StatusOK, resposta{
		Status:     "sucesso",
		Protocolo:  att.Protocolo,
		Cliente:    att.ClienteNome,
		Cpf:        atendimento.FormataCPF(cpf),
		Prazo:      prazoDe(prioridade),
		Prioridade: string(prioridade),
		Mensagem: aviso + "Chamado " + att.Protocolo + " aberto para " + att.ClienteNome +
			", com prioridade " + prioridade.Rotulo() + ". " + prazoDe(prioridade) +
			", e pode acompanhar tudo pelo protocolo.",
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

func resumoDo(problema, descricao string) string {
	rotulos := map[string]string{
		"sem_sinal": "Sem sinal", "sem_internet": "Sem internet", "sem_conexao": "Sem conexão",
		"lentidao": "Lentidão na conexão", "wifi": "Problema com wi-fi",
		"tv": "Problema com o sinal de TV", "outro": "Problema técnico",
	}
	if r, ok := rotulos[problema]; ok {
		return r
	}
	if len(descricao) > 60 {
		return descricao[:60] + "…"
	}
	return descricao
}

func prazoDe(p atendimento.Prioridade) string {
	switch p {
	case atendimento.PrioridadeAlta:
		return "O atendimento acontece em até 24 horas úteis"
	case atendimento.PrioridadeBaixa:
		return "O atendimento acontece em até 5 dias úteis"
	}
	return "O atendimento acontece em até 48 horas úteis"
}

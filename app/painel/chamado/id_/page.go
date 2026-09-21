// A tela de um atendimento: o dossiê completo do que a voz registrou e os
// formulários de mudança de status.
package chamado_

import (
	"time"

	"github.com/emersonjoe/trilha"
	"github.com/emersonjoe/trilha/h"
	"github.com/emersonjoe/trilha/ui"

	"github.com/emersonjoe/voice-based-customer-service/internal/atendimento"

	painel "github.com/emersonjoe/voice-based-customer-service/app/painel"
)

// Page renderiza GET /painel/chamado/{id}.
func Page(c *trilha.Ctx) (h.Node, error) {
	store := trilha.Use[*atendimento.Store](c)
	a, ok := store.PorID(c.Param("id"))
	if !ok {
		c.SetTitle("Atendimento não encontrado")
		c.Status(404)
		return ui.Empty(ui.EmptyOpts{
			Icon:   "search",
			Title:  "Atendimento não encontrado",
			Hint:   "O protocolo pode ter sido aberto em outra base de dados — esta é a da demonstração.",
			Action: ui.ButtonLink("/painel", h.Text("Voltar ao painel")),
		}), nil
	}

	c.SetTitle(a.Protocolo + " · Central de atendimentos")

	detalhes := dossie(a)
	if detalhes == nil {
		detalhes = []h.Node{h.Li(h.Class("wh-evento"), h.Em(h.Text("Sem detalhes adicionais.")))}
	}

	return ui.Stack(
		ui.PageHeader(a.Protocolo,
			h.P(h.Class("wh-sub"), h.Text(a.Tipo.Rotulo()+" · aberto por "+a.Origem)),
			ui.ButtonLink("/painel", ui.Outline(), ui.Icon("arrow-left"), h.Text("Voltar")),
		),
		ui.Grid(
			ui.Card(
				ui.CardHeader(ui.CardTitle("Atendimento"), ui.CardDescription(a.Resumo)),
				ui.CardContent(
					h.P(h.Class("wh-descricao"), h.Text(a.Descricao)),
					h.Ul(h.Class("wh-dados"),
						dado("Status", "", painel.BadgeStatus(a.Status)),
						dado("Prioridade", "", painel.BadgePrioridade(a.Prioridade)),
						dado("Criado", painel.Quando(c, a.CriadoEm), nil),
						dado("Atualizado", painel.Quando(c, a.AtualizadoEm), nil),
						dado("Conversa", a.ConversaID, nil),
					),
				),
				ui.CardFooter(acoes(c, a)...),
			),
			ui.Card(
				ui.CardHeader(ui.CardTitle("Cliente"), ui.CardDescription("Base fictícia da demonstração")),
				ui.CardContent(dadosCliente(c, store, a)),
			),
		),
		ui.Card(
			ui.CardHeader(ui.CardTitle("Detalhes do fluxo"), ui.CardDescription("O que a ferramenta registrou")),
			ui.CardContent(h.Ul(append([]h.Node{h.Class("wh-dados")}, detalhes...)...)),
		),
	), nil
}

func dadosCliente(c *trilha.Ctx, store *atendimento.Store, a *atendimento.Atendimento) h.Node {
	cliente, ok := store.ClientePorCPF(a.ClienteCPF)
	if !ok {
		return h.Ul(h.Class("wh-dados"),
			dado("CPF", atendimento.FormataCPF(a.ClienteCPF), nil),
			dado("Nome", a.ClienteNome, nil),
		)
	}
	var faturas h.Node
	agora := time.Now()
	if f, tem := cliente.FaturaAberta(agora); tem {
		estado := "em aberto"
		if f.Vencida(agora) {
			estado = "vencida"
		}
		faturas = h.Text(atendimento.BRL(f.ValorCentavos) + " · vence " + f.Vencimento.Format("02/01/2006") + " (" + estado + ")")
	} else {
		faturas = h.Text("nenhuma em aberto")
	}
	return h.Ul(h.Class("wh-dados"),
		dado("Nome", cliente.Nome, nil),
		dado("CPF", atendimento.FormataCPF(cliente.CPF), nil),
		dado("Plano", cliente.Plano, nil),
		dado("Telefone", cliente.Telefone, nil),
		dado("E-mail", cliente.Email, nil),
		dado("Endereço", cliente.Endereco, nil),
		dado("Fatura", "", faturas),
	)
}

// dossie lista os detalhes que a ferramenta guardou (valor, canal, forma de
// pagamento…), com rótulos amigáveis.
func dossie(a *atendimento.Atendimento) []h.Node {
	rotulos := map[string]string{
		"problema":        "Problema relatado",
		"telefone":        "Telefone informado",
		"email":           "E-mail informado",
		"canal":           "Canal de envio",
		"enviado_para":    "Enviado para",
		"valor":           "Valor",
		"vencimento":      "Vencimento",
		"linha":           "Linha digitável",
		"forma_pagamento": "Forma de pagamento",
		"data_pagamento":  "Data do pagamento",
		"comprovante":     "Comprovante",
	}
	var nodes []h.Node
	for _, k := range []string{
		"problema", "telefone", "email", "canal", "enviado_para", "valor",
		"vencimento", "linha", "forma_pagamento", "data_pagamento", "comprovante",
	} {
		v, ok := a.Detalhes[k]
		if !ok || v == "" {
			continue
		}
		nodes = append(nodes, dado(rotulos[k], v, nil))
	}
	return nodes
}

// acoes devolve os formulários de transição de status, cada um com o CSRF
// da página e o caminho de volta.
func acoes(c *trilha.Ctx, a *atendimento.Atendimento) []h.Node {
	form := func(destino atendimento.Status, rotulo string) h.Node {
		return h.Form(
			h.Method("post"), h.Action("/api/chamados/"+a.ID+"/status"), h.Class("wh-acao"),
			trilha.CSRFInput(c),
			h.Input(h.Type("hidden"), h.Name("status"), h.Value(string(destino))),
			h.Input(h.Type("hidden"), h.Name("voltar"), h.Value("/painel/chamado/"+a.ID)),
			ui.Button(h.Type("submit"), h.Text(rotulo)),
		)
	}
	switch a.Status {
	case atendimento.StatusAberto:
		return []h.Node{
			form(atendimento.StatusEmAtendimento, "Iniciar atendimento"),
			form(atendimento.StatusResolvido, "Marcar resolvido"),
		}
	case atendimento.StatusEmAtendimento:
		return []h.Node{form(atendimento.StatusResolvido, "Marcar resolvido"), form(atendimento.StatusAberto, "Voltar para aberto")}
	case atendimento.StatusResolvido:
		return []h.Node{form(atendimento.StatusAberto, "Reabrir")}
	}
	return nil
}

func dado(rotulo, valor string, no h.Node) h.Node {
	corpo := no
	if corpo == nil {
		corpo = h.Text(valor)
	}
	return h.Li(h.Class("wh-dado"),
		h.Span(h.Class("wh-dado-rotulo"), h.Text(rotulo)),
		h.Span(h.Class("wh-dado-valor"), corpo),
	)
}

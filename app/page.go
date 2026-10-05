package app

import (
	"os"

	"github.com/emersonjoe/trilha"
	"github.com/emersonjoe/trilha/h"
	"github.com/emersonjoe/trilha/ui"
)

// Page renders GET /: a vitrine da POC — hero, os três fluxos de atendimento,
// o widget de voz (ElevenLabs) e o roteiro da demonstração.
func Page(c *trilha.Ctx) (h.Node, error) {
	c.SetTitle("WaveHub · Demonstração de atendimento por voz com IA")
	agenteID := os.Getenv("ELEVENLABS_AGENT_ID")
	agenteCartesiaID := os.Getenv("CARTESIA_AGENT_ID")

	return ui.Stack(
		hero(),
		ui.Grid(
			h.Section(h.Class("wh-flows"),
				h.H2(h.Class("wh-section-title"), h.Text("Três fluxos, um atendente")),
				fluxo("triangle-alert", "Chamado técnico",
					"O cliente descreve o problema; a assistente pergunta o essencial, registra o chamado e devolve protocolo e prazo."),
				fluxo("calendar", "Segunda via do boleto",
					"Identificação por CPF, consulta da fatura em aberto e envio do boleto por e-mail ou WhatsApp, na hora."),
				fluxo("circle-check", "Religue com comprovação",
					"O sinal só volta com o pagamento confirmado: a assistente exige o comprovante antes de abrir o pedido."),
			),
			WidgetVoz(WidgetVozOpts{
				ProvedorFixo:         "",
				Titulo:               "Fale com a assistente",
				Descricao:            "Clique, permita o microfone e peça: “minha internet caiu e quero abrir um chamado”.",
				AgentePadrao:         agenteID,
				AgenteCartesiaPadrao: agenteCartesiaID,
			}),
		),
		comoFunciona(),
		roteiroDemo(),
	), nil
}

func hero() h.Node {
	return h.Section(h.Class("wh-hero"),
		ui.Badge(h.Class("wh-hero-badge"), h.Text("Prova de conceito · Atendimento por voz com IA")),
		ui.H1(
			h.Text("O cliente fala. A assistente "),
			h.Span(h.Class("wh-grad"), h.Text("resolve na hora")),
			h.Text("."),
		),
		ui.Lead(h.Text(
			"Chamado técnico, segunda via do boleto e religue com comprovação — "+
				"sem fila, sem tecla. A voz é da ElevenLabs; o atendimento aparece no painel em tempo real.")),
		h.Div(h.Class("wh-cta"),
			ui.ButtonLink("/painel", ui.Lg(), h.Text("Abrir o painel ao vivo"), ui.Icon("arrow-right")),
			ui.ButtonLink("/agente", ui.Lg(), ui.Outline(), h.Text("Configurar o agente")),
		),
	)
}

func fluxo(icone, titulo, texto string) h.Node {
	return ui.Card(
		ui.CardHeader(
			h.Span(h.Class("wh-fluxo-icone"), ui.Icon(icone)),
			ui.CardTitle(titulo),
		),
		ui.CardContent(h.P(h.Class("wh-fluxo-texto"), h.Text(texto))),
	)
}

func comoFunciona() h.Node {
	passo := func(n, titulo, texto string) h.Node {
		return h.Div(h.Class("wh-passo"),
			h.Span(h.Class("wh-passo-n"), h.Text(n)),
			h.Div(h.Strong(h.Text(titulo)), h.P(h.Class("wh-passo-texto"), h.Text(texto))),
		)
	}
	return h.Section(h.Class("wh-como"),
		h.H2(h.Class("wh-section-title"), h.Text("Como funciona")),
		ui.Grid(
			passo("1", "Cliente fala", "A conversa corre por voz, com latência de conversa real, pela plataforma de agentes da ElevenLabs."),
			passo("2", "Agente entende e pergunta", "O prompt dirige a conversa: o essencial é perguntado, o que não é do escopo é dito com clareza."),
			passo("3", "Ferramenta executa", "Chamado técnico, segunda via ou religue: a ferramenta chama a API deste app e registra o atendimento."),
			passo("4", "Painel atualiza", "O painel acompanha em tempo real: status, prioridade, protocolo e o histórico de cada atendimento."),
		),
	)
}

func roteiroDemo() h.Node {
	cliente := func(nome, cpf, roteiro string) h.Node {
		return h.Div(h.Class("wh-demo-linha"),
			h.Strong(h.Text(nome)), h.Text(" "),
			ui.Kbd(cpf),
			h.P(h.Class("wh-demo-roteiro"), h.Text(roteiro)),
		)
	}
	return ui.Card(h.Class("wh-demo-card"),
		ui.CardHeader(
			ui.CardTitle("Roteiro da demonstração"),
			ui.CardDescription("Base fictícia de clientes para apresentar sem medo: nada aqui é dado real."),
		),
		ui.CardContent(
			ui.Grid(
				cliente("Ana Beatriz Souza", "111.444.777-35", "Fatura vencida há 4 dias — teste o religue com comprovação e a segunda via."),
				cliente("Carlos Eduardo Ramos", "529.982.247-25", "Fatura em dia — teste o chamado técnico relatando que a internet caiu."),
				cliente("Marina Lopes", "168.995.350-09", "Fatura a vencer — peça a segunda via e mostre o envio por e-mail."),
			),
			h.P(h.Class("wh-demo-dica"),
				ui.Icon("info"),
				h.Text(" Abra o painel ao lado e deixe-o atualizando: cada ferramenta acionada pela voz aparece lá em segundos."),
			),
		),
		ui.CardFooter(ui.ButtonLink("/painel", ui.Outline(), h.Text("Ir ao painel"), ui.Icon("arrow-right"))),
	)
}

// A página /matarazzo: o concierge de voz do Complexo Cidade Matarazzo,
// com a identidade visual do complexo (dourado, creme e preto, serifas)
// e o widget de voz fixado no provedor Cartesia.
package matarazzo

import (
	"os"

	"github.com/emersonjoe/trilha"
	"github.com/emersonjoe/trilha/h"
	"github.com/emersonjoe/trilha/ui"

	app "github.com/emersonjoe/voice-based-customer-service/app"
	guia "github.com/emersonjoe/voice-based-customer-service/internal/matarazzo"
)

// Page renderiza GET /matarazzo.
func Page(c *trilha.Ctx) (h.Node, error) {
	c.SetTitle("Guia Matarazzo · Concierge por voz")

	return h.Div(h.Class("mt-page"),
		h.Section(h.Class("wh-hero mt-hero"),
			ui.Badge(h.Class("wh-hero-badge mt-chip"), h.Text("Prova de conceito · Concierge por voz")),
			ui.H1(
				h.Text("O concierge da "),
				h.Span(h.Class("mt-dourado"), h.Text("Cidade Matarazzo")),
				h.Text("."),
			),
			ui.Lead(h.Text(
				"Hotel Rosewood, restaurantes, lojas e eventos num só complexo histórico. "+
					"Fale com o Gui: ele apresenta cada canto e registra a sua reserva.")),
			h.Div(h.Class("wh-cta"),
				ui.ButtonLink("/painel", ui.Lg(), h.Class("mt-botao"), h.Text("Ver o painel ao vivo"), ui.Icon("arrow-right")),
				ui.ButtonLink("/agente", ui.Lg(), ui.Outline(), h.Text("Como funciona")),
			),
		),
		ui.Grid(
			h.Section(h.Class("wh-flows mt-flows"),
				h.H2(h.Class("mt-titulo"), h.Text("O que dá para pedir ao Gui")),
				cartaoGuia("hotel_rosewood"),
				cartaoGuia("restaurantes"),
				cartaoGuia("mata_citta"),
				cartaoGuia("lojas"),
				cartaoGuia("eventos"),
				cartaoGuia("como_chegar"),
			),
			ui.Stack(
				app.WidgetVoz(app.WidgetVozOpts{
					ProvedorFixo:         "cartesia",
					Titulo:               "Fale com o Gui",
					Descricao:            "Clique, permita o microfone e peça: “quero jantar no Taraz sexta às 20 horas para duas pessoas”.",
					AgenteCartesiaPadrao: os.Getenv("MATARAZZO_AGENT_ID"),
				}),
				ui.Card(h.Class("mt-card"),
					ui.CardHeader(
						ui.CardTitle("Roteiro da demonstração"),
						ui.CardDescription("Reservas fictícias, no nome de quem você quiser"),
					),
					ui.CardContent(h.Ul(h.Class("wh-dados"),
						h.Li(h.Class("wh-dado"),
							h.Span(h.Class("wh-dado-rotulo"), h.Text("Apresentação")),
							h.Span(h.Class("wh-dado-valor"), h.Text("“o que tem na Cidade Matarazzo?”"))),
						h.Li(h.Class("wh-dado"),
							h.Span(h.Class("wh-dado-rotulo"), h.Text("Reserva")),
							h.Span(h.Class("wh-dado-valor"), h.Text("“quero jantar no Taraz sexta às 20h para duas pessoas, no nome de Marina”"))),
						h.Li(h.Class("wh-dado"),
							h.Span(h.Class("wh-dado-rotulo"), h.Text("Mata Città")),
							h.Span(h.Class("wh-dado-valor"), h.Text("“dá para reservar o Mata Città?” — a resposta honesta é lista de espera"))),
					)),
				),
			),
		),
		secaoConfig(),
	), nil
}

func cartaoGuia(topico guia.Topico) h.Node {
	entrada, ok := guia.Guia[topico]
	if !ok {
		return nil
	}
	var dicas []h.Node
	for _, dica := range entrada.Dicas {
		dicas = append(dicas, h.Li(h.Text(dica)))
	}
	return ui.Card(h.Class("mt-card"),
		ui.CardHeader(ui.CardTitle(entrada.Titulo)),
		ui.CardContent(
			h.P(h.Class("wh-fluxo-texto"), h.Text(entrada.Texto)),
			h.Ul(append([]h.Node{h.Class("mt-dicas")}, dicas...)...),
		),
	)
}

// secaoConfig documenta o agente já configurado nesta POC: prompt,
// saudação e os payloads das duas client tools, copiáveis.
func secaoConfig() h.Node {
	ficha := func(id, titulo, esquema string) h.Node {
		return ui.Card(
			ui.CardHeader(ui.CardTitle(titulo)),
			ui.CardContent(copiavel(id, esquema)),
			ui.CardFooter(ui.Button(ui.Sm(), ui.Outline(), h.Class("wh-copy"), h.Data("copy-target", id), h.Text("Copiar JSON da tool"))),
		)
	}
	return ui.Stack(
		ui.Card(h.Class("mt-card"),
			ui.CardHeader(
				ui.CardTitle("Configurar o agente (já configurado nesta POC)"),
				ui.CardDescription("Standard Agent na Cartesia · voz Helena (pt-BR) · timezone America/Sao_Paulo"),
			),
			ui.CardContent(
				copiavel("mt-prompt", PROMPT_MATARAZZO),
				h.P(h.Class("mt-rotulo"), h.Text("Welcome message")),
				copiavel("mt-boas-vindas", PRIMEIRA_MATARAZZO),
				h.P(h.Class("wh-rota"), ui.Kbd("MATARAZZO_AGENT_ID="+os.Getenv("MATARAZZO_AGENT_ID"))),
			),
			ui.CardFooter(ui.Button(ui.Sm(), ui.Outline(), h.Class("wh-copy"), h.Data("copy-target", "mt-prompt"), h.Text("Copiar o prompt"))),
		),
		ui.Grid(
			ficha("mt-tool-guia", "consultar_guia_matarazzo", ESQUEMA_MT_GUIA),
			ficha("mt-tool-reserva", "criar_reserva_matarazzo", ESQUEMA_MT_RESERVA),
		),
	)
}

// copiavel desenha um bloco <pre> com o texto em data-copy; o botão
// .wh-copy com data-copy-target aponta para ele (mesmo mecanismo da /agente).
func copiavel(id, texto string) h.Node {
	return h.Pre(h.ID(id), h.Class("wh-pre"), h.Data("copy", texto), h.Code(h.Text(texto)))
}

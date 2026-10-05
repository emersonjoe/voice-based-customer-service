package app

import (
	"github.com/emersonjoe/trilha/h"
	"github.com/emersonjoe/trilha/ui"
)

// WidgetVoz é o cartão do assistente de voz, reutilizado pelas páginas de
// cada agente. Os ids são fixos (voice.js os procura); o que muda por
// página são os defaults de agente e o provedor fixado.
func WidgetVoz(o WidgetVozOpts) h.Node {
	var barras []h.Node
	for i := 1; i <= 7; i++ {
		barras = append(barras, h.Span(h.Class("wh-bar")))
	}
	orb := []h.Node{h.ID("wh-orb"), h.Class("wh-orb")}
	orb = append(orb, barras...)

	selectField := ui.Field("wh-provedor", "Provedor de voz",
		ui.Select(h.ID("wh-provedor"), h.Name("provedor"), h.Aria("label", "Provedor de voz"),
			h.Option(h.Value("elevenlabs"), h.Selected(), h.Text("ElevenLabs")),
			h.Option(h.Value("cartesia"), h.Text("Cartesia"))),
		ui.Help("Os dois ficam prontos para a demonstração — troque, recarregue e compare."))
	if o.ProvedorFixo != "" {
		// voice.js trava no provedor fixo; o bloco sai do caminho visual
		selectField = h.Div(h.Hidden(), selectField)
	}

	agentField := ui.Field("wh-agent-id", "Agent ID do provedor",
		ui.Input(h.ID("wh-agent-id"), h.Name("agente_id"),
			h.Placeholder("cole aqui o ID do agente publicado"),
			h.Value(o.AgentePadrao)),
		ui.Help("Gerado na plataforma do provedor — o passo a passo está na página Agente IA. Fica salvo apenas neste navegador."))

	return ui.Card(h.Class("wh-voice-card"),
		ui.CardHeader(
			ui.CardTitle(o.Titulo),
			ui.CardDescription(o.Descricao),
		),
		ui.CardContent(
			h.Div([]h.Node{
				h.ID("wh-voice"), h.Class("wh-voice"),
				h.Data("agent-id", o.AgentePadrao),
				h.Data("agent-id-cartesia", o.AgenteCartesiaPadrao),
				h.Data("provedor-fixo", o.ProvedorFixo),
				h.Div(orb...),
				h.P(h.ID("wh-status"), h.Class("wh-status"), h.Text("Toque em iniciar para falar")),
				h.Div(h.Class("wh-controls"),
					ui.Button(h.ID("wh-toggle"), ui.Lg(), h.Text("Iniciar conversa")),
				),
				selectField,
				agentField,
				h.Div(h.Class("wh-feed-wrap"),
					h.P(h.Class("wh-feed-title"), h.Text("Ferramentas acionadas")),
					h.Ul(h.ID("wh-feed"), h.Class("wh-feed"),
						h.Li(h.Class("wh-feed-empty"), h.Text("As ações que a assistente executar aparecem aqui em tempo real."))),
				),
				h.Div(h.Class("wh-textrow"),
					ui.Input(h.ID("wh-text"), h.Name("mensagem"),
						h.Placeholder("Sem microfone? Escreva sua mensagem…")),
					ui.Button(h.ID("wh-send"), ui.Outline(), h.Text("Enviar")),
				),
			}...),
		),
	)
}

// WidgetVozOpts parametriza o cartão de voz por página.
type WidgetVozOpts struct {
	ProvedorFixo         string // "" = seletor visível; "cartesia" trava o provedor
	Titulo               string
	Descricao            string
	AgentePadrao         string // default do campo (ElevenLabs)
	AgenteCartesiaPadrao string // default quando o provedor é Cartesia
}

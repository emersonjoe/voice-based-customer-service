package app

import (
	"github.com/emersonjoe/trilha"
	"github.com/emersonjoe/trilha/h"
	"github.com/emersonjoe/trilha/ui"
)

// Layout é o layout raiz: o documento <html> em volta de todas as páginas
// públicas, com a identidade visual da WaveHub.
func Layout(c *trilha.Ctx, children h.Node) (h.Node, error) {
	title := c.Title()
	if title == "" {
		title = "WaveHub · Atendimento por voz com IA"
	}
	return h.Html(h.Lang(c.Locale()),
		h.Head(
			h.Meta(h.Charset("utf-8")),
			h.Meta(h.Name("viewport"), h.Content("width=device-width, initial-scale=1")),
			h.Meta(h.Name("description"), h.Content("Prova de conceito de atendimento por voz com IA: chamado técnico, segunda via de boleto e religue com comprovação. Identidade WaveHub, voz ElevenLabs, construído com Trilha.")),
			h.Title(h.Text(title)),
			ui.Head(c), // public/ui.theme.css (o tema), ui.css e ui.js
			h.Link(h.Rel("stylesheet"), h.Href(c.Asset("/style.css"))),
			h.Script(h.Src(c.Asset("/voice.js")), h.Type("module")),
			h.Script(h.Src(c.Asset("/copy.js")), h.Defer()),
		),
		h.Body(ui.Body(),
			ui.Header(
				ui.Brand("/", "WaveHub"),
				ui.Nav(
					ui.NavLink("/", "Demonstração", c.Request().URL.Path == "/"),
					ui.NavLink("/painel", "Painel", c.Request().URL.Path == "/painel"),
					ui.NavLink("/agente", "Agente IA", c.Request().URL.Path == "/agente"),
				),
				ui.Spacer(),
				ui.Badge(h.Class("wh-chip"), h.Text("POC")),
				ui.ThemeToggle(),
			),
			h.Main(ui.Container(children)),
			ui.Container(h.Footer(h.Class("wh-footer"),
				h.Small(h.Text("Prova de conceito · construída com "),
					h.A(h.Href("https://github.com/emersonjoe/trilha"), h.Text("Trilha")),
					h.Text(" (Go, sem dependências) · voz por "),
					h.A(h.Href("https://elevenlabs.io"), h.Text("ElevenLabs")),
					h.Text(" · identidade "),
					h.A(h.Href("https://wavehub.dev"), h.Text("WaveHub")),
				),
			)),
			ui.Flashes(c), // c.Flash(ui.FlashSuccess, "…") aparece aqui depois do redirect
		),
	), nil
}

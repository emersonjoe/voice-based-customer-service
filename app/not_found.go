package app

import (
	"github.com/emersonjoe/trilha"
	"github.com/emersonjoe/trilha/h"
	"github.com/emersonjoe/trilha/ui"
)

// NotFound renderiza a 404 com o tom da casa.
func NotFound(c *trilha.Ctx) (h.Node, error) {
	c.SetTitle("Página não encontrada · WaveHub")
	c.Status(404)
	return ui.Empty(ui.EmptyOpts{
		Icon:   "search",
		Title:  "Esta página não existe",
		Hint:   "O link pode estar velho — a demonstração, o painel e o guia do agente estão sempre no menu.",
		Action: ui.ButtonLink("/", h.Text("Voltar à demonstração")),
	}), nil
}

package painel

import (
	"github.com/emersonjoe/trilha"
	"github.com/emersonjoe/trilha/h"
	"github.com/emersonjoe/trilha/ui"
)

// Layout é a moldura do painel: a barra lateral do kit ui em volta das
// telas de gestão, dentro do documento raiz da WaveHub.
func Layout(c *trilha.Ctx, children h.Node) (h.Node, error) {
	return ui.Shell(c, ui.ShellOpts{
		Brand: ui.Brand("/painel", "WaveHub · POC"),
		Nav: []ui.NavGroup{
			{Label: "Atendimento", Items: []ui.NavItem{
				{Href: "/painel", Label: "Visão geral", Icon: "house"},
				{Href: "/agente", Label: "Agente de voz", Icon: "settings"},
			}},
		},
		Header: []h.Node{ui.ThemeToggle()},
		User: ui.UserMenu{
			Name:   "Equipe WaveHub",
			Detail: "Demonstração",
			Items: []h.Node{
				ui.MenuLink("/", h.Text("Abrir a demonstração")),
				ui.MenuLink("/agente", h.Text("Configurar o agente")),
			},
		},
	}, children), nil
}

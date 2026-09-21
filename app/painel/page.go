package painel

import (
	"fmt"
	"time"

	"github.com/emersonjoe/trilha"
	"github.com/emersonjoe/trilha/h"
	"github.com/emersonjoe/trilha/ui"

	"github.com/emersonjoe/voice-based-customer-service/internal/atendimento"
)

// filtro é o contrato da URL da listagem: busca, facetas e página. O
// ListParams embutido carrega page, sort, dir e q do endereço.
type filtro struct {
	trilha.ListParams
	Tipo   string `form:"tipo"`
	Status string `form:"status"`
}

// Page renderiza GET /painel: resumo com atualização periódica e a tabela
// de atendimentos com busca, filtro e ordem no endereço.
func Page(c *trilha.Ctx) (h.Node, error) {
	c.SetTitle("Central de atendimentos · WaveHub")

	var f filtro
	if err := c.Bind(&f); err != nil {
		return nil, err
	}

	store := trilha.Use[*atendimento.Store](c)

	// O fragmento #resumo é o que o ui.Poll pede a cada 6s: o painel vivo
	// durante a demonstração, sem recarregar a página.
	if c.Fragment() == "resumo" {
		return resumo(c, store), nil
	}

	filtroAtendimento := atendimento.Filtro{ListParams: f.ListParams}
	if atendimento.Tipo(f.Tipo).Valido() {
		filtroAtendimento.Tipo = atendimento.Tipo(f.Tipo)
	}
	if atendimento.Status(f.Status).Valido() {
		filtroAtendimento.Status = atendimento.Status(f.Status)
	}
	atts, total := store.Listar(filtroAtendimento)

	return ui.Stack(
		ui.PageHeader("Central de atendimentos",
			h.P(h.Class("wh-sub"), h.Text("Tudo o que o agente de voz registra, em tempo real — chamados, segundas vias e religues.")),
			ui.ButtonLink("/agente", ui.Outline(), h.Text("Configurar o agente")),
		),
		h.Div(h.ID("resumo"), resumo(c, store), ui.Poll("6s", "")),
		ui.DataTable(c, colunas(c), atts, ui.ListState{
			Params:  f.ListParams,
			Total:   total,
			ID:      "atendimentos",
			Search:  "Protocolo, cliente, resumo…",
			Filters: facetas(f),
			Empty: ui.Empty(ui.EmptyOpts{
				Icon:  "loader-circle",
				Title: "Nenhum atendimento para o filtro atual",
				Hint:  "Fale com a assistente na demonstração: o que ela registrar aparece aqui em segundos.",
				Action: ui.ButtonLink("/", h.Text(
					"Abrir a demonstração")),
			}),
			RowHref: func(i int) string { return "/painel/chamado/" + atts[i].ID },
			Caption: "Atendimentos abertos pelo agente de voz",
		}),
	), nil
}

// resumo é o fragmento pollado: os contadores e os últimos eventos.
func resumo(c *trilha.Ctx, store *atendimento.Store) h.Node {
	porStatus := store.ContagemPorStatus()
	eventos := store.Ultimos(5)

	var lista []h.Node
	for _, a := range eventos {
		lista = append(lista, h.Li(h.Class("wh-evento"),
			BadgeTipo(a.Tipo),
			h.Strong(h.Text(a.Protocolo)),
			h.Span(h.Class("wh-evento-resumo"), h.Text(a.Resumo)),
			h.Small(h.Class("wh-evento-quando"), h.Text(Quando(c, a.CriadoEm))),
			h.Span(h.Class("wh-evento-origem"), h.Text(a.Origem)),
		))
	}
	if lista == nil {
		lista = []h.Node{h.Li(h.Class("wh-evento"), h.Em(h.Text("Nada ainda.")))}
	}

	return ui.Stack(
		ui.Grid(
			ui.Stat("Abertos", fmt.Sprintf("%d", porStatus[atendimento.StatusAberto])),
			ui.Stat("Em atendimento", fmt.Sprintf("%d", porStatus[atendimento.StatusEmAtendimento])),
			ui.Stat("Resolvidos hoje", fmt.Sprintf("%d", store.ResolvidosHoje(time.Now()))),
			ui.Stat("Total registrado", fmt.Sprintf("%d", store.Total())),
		),
		ui.Card(
			ui.CardHeader(
				ui.CardTitle("Últimos eventos"),
				ui.CardDescription("Atualiza a cada 6 segundos — deixe aberto durante a chamada de voz."),
			),
			ui.CardContent(h.Ul(append([]h.Node{h.Class("wh-eventos")}, lista...)...)),
		),
	)
}

func facetas(f filtro) h.Node {
	opcoes := func(atual string, pares [][2]string) []h.Node {
		var ops []h.Node
		for _, par := range pares {
			op := []h.Node{h.Value(par[0]), h.Text(par[1])}
			if atual == par[0] || (atual == "" && par[0] == "") {
				op = append(op, h.Selected())
			}
			ops = append(ops, h.Option(op...))
		}
		return ops
	}
	return h.Div(h.Class("wh-facetas"),
		ui.Select(append([]h.Node{
			h.ID("f-tipo"), h.Name("tipo"), h.Aria("label", "Filtrar por tipo"),
		}, opcoes(f.Tipo, tipos)...)...),
		ui.Select(append([]h.Node{
			h.ID("f-status"), h.Name("status"), h.Aria("label", "Filtrar por status"),
		}, opcoes(f.Status, statusOpcoes)...)...),
	)
}

var tipos = [][2]string{
	{"", "Todos os tipos"},
	{string(atendimento.TipoChamadoTecnico), "Chamado técnico"},
	{string(atendimento.TipoSegundaVia), "Segunda via"},
	{string(atendimento.TipoReligue), "Religue"},
}

var statusOpcoes = [][2]string{
	{"", "Todos os status"},
	{string(atendimento.StatusAberto), "Abertos"},
	{string(atendimento.StatusEmAtendimento), "Em atendimento"},
	{string(atendimento.StatusResolvido), "Resolvidos"},
}

func colunas(c *trilha.Ctx) ui.Columns[*atendimento.Atendimento] {
	return ui.Columns[*atendimento.Atendimento]{
		{Key: "protocolo", Label: "Protocolo", Sort: true, Cell: func(a *atendimento.Atendimento) h.Node {
			return h.Strong(h.Text(a.Protocolo))
		}},
		{Key: "cliente", Label: "Cliente", Cell: func(a *atendimento.Atendimento) h.Node {
			if a.ClienteNome == "" {
				return h.Text("—")
			}
			return h.Span(h.Text(a.ClienteNome))
		}},
		{Key: "tipo", Label: "Tipo", Cell: func(a *atendimento.Atendimento) h.Node {
			return BadgeTipo(a.Tipo)
		}},
		{Key: "resumo", Label: "Resumo", Cell: func(a *atendimento.Atendimento) h.Node {
			r := a.Resumo
			if len(r) > 48 {
				r = r[:48] + "…"
			}
			return h.Span(h.Text(r))
		}},
		{Key: "prioridade", Label: "Prioridade", Cell: func(a *atendimento.Atendimento) h.Node {
			return BadgePrioridade(a.Prioridade)
		}},
		{Key: "status", Label: "Status", Sort: true, Cell: func(a *atendimento.Atendimento) h.Node {
			return BadgeStatus(a.Status)
		}},
		{Key: "criado_em", Label: "Criado", Sort: true, Cell: func(a *atendimento.Atendimento) h.Node {
			return h.Span(h.Class("wh-quando"), h.Text(Quando(c, a.CriadoEm)))
		}},
	}
}

func BadgeTipo(t atendimento.Tipo) h.Node {
	return ui.Badge(h.Class("wh-badge wh-tipo-"+string(t)), h.Text(t.Rotulo()))
}

func BadgeStatus(s atendimento.Status) h.Node {
	return ui.Badge(h.Class("wh-badge wh-status-"+string(s)), h.Text(s.Rotulo()))
}

func BadgePrioridade(p atendimento.Prioridade) h.Node {
	if p == "" {
		p = atendimento.PrioridadeMedia
	}
	return ui.Badge(h.Class("wh-badge wh-prio-"+string(p)), h.Text(p.Rotulo()))
}

// quando formata o instante no fuso da aplicação, curto para a tabela.
func Quando(c *trilha.Ctx, t time.Time) string {
	hj := time.Now().In(c.Location())
	l := t.In(c.Location())
	if l.Format("02/01") == hj.Format("02/01") {
		return "hoje, " + l.Format("15:04")
	}
	return l.Format("02/01 15:04")
}

package agente

import (
	"github.com/emersonjoe/trilha"
	"github.com/emersonjoe/trilha/h"
	"github.com/emersonjoe/trilha/ui"
)

// Page renderiza GET /agente: o passo a passo para montar o agente na
// plataforma da ElevenLabs — prompt, ferramentas, voz e o plano B por
// webhook. Os blocos copiáveis vivem em <pre data-copy>.
func Page(c *trilha.Ctx) (h.Node, error) {
	c.SetTitle("Configurar o agente · WaveHub")

	return ui.Stack(
		ui.PageHeader("Configurar o agente na ElevenLabs",
			h.P(h.Class("wh-sub"), h.Text("Do zero ao agente falando português em cerca de dez minutos. Tudo o que está aqui cola direto na plataforma.")),
			ui.ButtonLink("/", ui.Outline(), h.Text("Voltar à demonstração")),
		),
		passo("1", "Crie o agente",
			"Na plataforma de agentes (elevenlabs.io → Agents), crie um agente novo, publique-o como público e copie o Agent ID para o campo do widget na demonstração."),
		passo("2", "Cole o prompt do sistema",
			"O texto abaixo dirige a conversa: os três fluxos, o que perguntar e o que nunca inventar."),
		ui.Card(
			ui.CardHeader(ui.CardTitle("Prompt do sistema"), ui.CardDescription("Agents → System prompt")),
			ui.CardContent(
				copiavel("prompt-agente", PROMPT_SISTEMA),
			),
			ui.CardFooter(ui.Button(ui.Sm(), ui.Outline(), h.Class("wh-copy"), h.Data("copy-target", "prompt-agente"), h.Text("Copiar prompt"))),
		),
		passo("3", "Cadastre as três ferramentas (client tools)",
			"Em Tools, adicione cada ferramenta como “Client tool” com o nome, a descrição e o esquema de parâmetros abaixo. São o browser ou o seu servidor que as executam — este app expõe as três rotas."),
		ferramentas(),
		passo("4", "Voz, modelo e primeira mensagem",
			"Idioma português (pt-BR), voz masculina ou feminina em PT-BR (ex.: Gabriel, Antônio ou Ana), modelo Eleven Turbo v2.5, otimizado para latência. Primeira mensagem sugerida abaixo."),
		ui.Card(
			ui.CardHeader(ui.CardTitle("Primeira mensagem"), ui.CardDescription("O que a assistente fala ao atender")),
			ui.CardContent(copiavel("primeira-msg", PRIMEIRA_MENSAGEM)),
			ui.CardFooter(ui.Button(ui.Sm(), ui.Outline(), h.Class("wh-copy"), h.Data("copy-target", "primeira-msg"), h.Text("Copiar mensagem"))),
		),
		webhookAlternativo(),
		agentePrivado(),
		checklist(),
	), nil
}

func passo(n, titulo, texto string) h.Node {
	return h.Div(h.Class("wh-passo wh-passo-guia"),
		h.Span(h.Class("wh-passo-n"), h.Text(n)),
		h.Div(h.Strong(h.Text(titulo)), h.P(h.Class("wh-passo-texto"), h.Text(texto))),
	)
}

func ferramentas() h.Node {
	ficha := func(nome, quandoUsar, descricao, esquema, alvo string) h.Node {
		id := "esquema-" + nome
		return ui.Card(
			ui.CardHeader(ui.CardTitle(nome), ui.CardDescription(quandoUsar)),
			ui.CardContent(
				h.P(h.Class("wh-fluxo-texto"), h.Text(descricao)),
				copiavel(id, esquema),
				h.P(h.Class("wh-rota"), ui.Kbd("POST "+alvo)),
			),
			ui.CardFooter(ui.Button(ui.Sm(), ui.Outline(), h.Class("wh-copy"), h.Data("copy-target", id), h.Text("Copiar esquema JSON"))),
		)
	}
	return ui.Grid(
		ficha("abrir_chamado_tecnico",
			"Chamar quando o cliente relatar um problema técnico e a assistente já tiver nome, CPF e descrição.",
			"Abre o chamado e devolve protocolo e prazo na resposta — o agente lê esses campos para o cliente.",
			ESQUEMA_CHAMADO, "/api/tools/abrir_chamado_tecnico"),
		ficha("enviar_segunda_via_boleto",
			"Chamar quando o cliente pedir a segunda vez do boleto, após colher o CPF.",
			"Localiza o cadastro, pega a fatura em aberto e devolve valor, vencimento, linha digitável e destino do envio.",
			ESQUEMA_SEGUNDA_VIA, "/api/tools/enviar_segunda_via_boleto"),
		ficha("solicitar_religue_confirmacao",
			"Chamar para o religue — e só com pagamento_confirmado=true depois que o cliente comprovar o pagamento.",
			"Sem confirmação, responde comprovacao_pendente; com ela, registra o pedido com prazo de até 2 horas úteis.",
			ESQUEMA_RELIGUE, "/api/tools/solicitar_religue_confirmacao"),
	)
}

func webhookAlternativo() h.Node {
	return ui.Card(
		ui.CardHeader(
			ui.CardTitle("Alternativa: ferramentas por webhook"),
			ui.CardDescription("Para executar no servidor em vez do browser"),
		),
		ui.CardContent(
			h.P(h.Class("wh-fluxo-texto"), h.Text(
				"As mesmas ferramentas funcionam como Webhook tool: cadastre o endereço público deste app "+
					"(em produção, atrás de HTTPS) e a ElevenLabs chama o servidor diretamente. O corpo e a resposta "+
					"são os mesmos do esquema acima, com conversation_id incluído.")),
			h.P(h.Class("wh-rota"), ui.Kbd("https://SEU-DOMINIO/api/tools/abrir_chamado_tecnico")),
			h.P(h.Class("wh-fluxo-texto"), h.Text(
				"Para uma apresentação local, prefira as client tools: o widget já fala com este app na mesma origem, "+
					"sem expor nada à internet.")),
		),
	)
}

// agentePrivado explica o fluxo de sessão assinada, para quem ligou a
// autenticação do agente na ElevenLabs.
func agentePrivado() h.Node {
	item := func(texto string) h.Node {
		return h.Li(h.Class("wh-check"), ui.Icon("circle-check"), h.Span(h.Text(texto)))
	}
	return ui.Card(
		ui.CardHeader(
			ui.CardTitle("Agente privado (autenticação habilitada)"),
			ui.CardDescription("Quando o agente não é público na ElevenLabs"),
		),
		ui.CardContent(h.Ul(h.Class("wh-checks"),
			item("Exporte ELEVENLABS_API_KEY no ambiente do servidor — é a chave da sua conta ElevenLabs, e nunca vai para o navegador."),
			item("O widget então pede a sessão ao servidor: GET /api/voz/assinada?agente=… devolve uma URL assinada, curta-viva."),
			item("Sem a chave configurada, o widget tenta o modo público com o Agent ID direto — e falha com erro de conexão se o agente for privado."),
			item("A URL assinada vale como credencial de conversa: não aparece em logs."),
		)),
	)
}

func checklist() h.Node {
	item := func(texto string) h.Node {
		return h.Li(h.Class("wh-check"), ui.Icon("circle-check"), h.Span(h.Text(texto)))
	}
	return ui.Card(
		ui.CardHeader(ui.CardTitle("Checklist antes de apresentar"), ui.CardDescription("Cinco minutos que salvam a demo")),
		ui.CardContent(h.Ul(h.Class("wh-checks"),
			item("Agente publicado (público) e Agent ID colado no widget."),
			item("Microfone permitido no navegador — teste uma frase curta antes da reunião."),
			item("Painel aberto em outra janela, atualizando a cada 6 segundos."),
			item("Falar perto do microfone e em ambiente silencioso; a frase de teste: “minha internet caiu e quero abrir um chamado”."),
			item("Sem internet, nada funciona: a conversa é um WebSocket com a ElevenLabs. Leve um plano de dados móveis."),
		)),
		ui.CardFooter(ui.ButtonLink("/", h.Text("Ir para a demonstração"), ui.Icon("arrow-right"))),
	)
}

// copiavel desenha um bloco <pre> com o texto completo em data-copy; o
// botão .wh-copy com data-copy-target aponta para ele.
func copiavel(id, texto string) h.Node {
	return h.Pre(h.ID(id), h.Class("wh-pre"), h.Data("copy", texto), h.Code(h.Text(texto)))
}

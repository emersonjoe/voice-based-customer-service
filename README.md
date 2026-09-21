# voice-based-customer-service

Prova de conceito de **atendimento por voz com IA** para a [WaveHub](https://wavehub.dev):
o cliente fala com a assistente e ela executa os três fluxos de sempre —
**chamado técnico**, **segunda via do boleto** e **religue com comprovação de
pagamento** — enquanto um painel acompanha tudo em tempo real.

A identidade visual é a da WaveHub (gradiente roxo→fúcsia sobre quase preto,
acentos amarelos), a voz é da plataforma de agentes da
[ElevenLabs](https://elevenlabs.io) e o servidor é feito com
[**Trilha**](https://github.com/emersonjoe/trilha), framework Go com
roteamento por arquivos e **zero dependências fora da biblioteca padrão**.

| Página | O que mostra |
|---|---|
| `/` | Demonstração: hero, os três fluxos e o widget de voz |
| `/painel` | Central de atendimentos: contadores, eventos ao vivo (poll de 6s) e tabela com busca/filtro/ordem |
| `/painel/chamado/{id}` | Dossiê do atendimento: cliente, detalhes do fluxo e ações de status |
| `/agente` | Passo a passo para montar o agente na ElevenLabs (prompt e esquemas copiáveis) |

## Como funciona

1. O agente conversa por voz com o cliente (WebSocket com a ElevenLabs).
2. Quando o fluxo pede uma ação, a **client tool** executa no navegador e
   chama a API deste app — o mesmo endpoint serve como **webhook** para a
   ElevenLabs chamar o servidor diretamente em produção.
3. A resposta JSON volta para o agente, que informa protocolo e prazo ao
   cliente.
4. O painel mostra o atendimento registrado segundos depois.

## Rodando

Pré-requisitos: [Go 1.22+](https://go.dev) e a CLI do Trilha
(`go install github.com/emersonjoe/trilha/cmd/trilha@latest`).

```bash
trilha dev --addr :3000
```

Abra `http://localhost:3000`. Sem nenhuma configuração extra o app já sobe
com uma base fictícia de clientes e atendimentos de exemplo.

### Configurar o agente de voz

1. Siga o guia da página **`/agente`**: cria o agente na ElevenLabs, cola o
   prompt do sistema, cadastra as três client tools (esquemas JSON prontos
   para copiar) e publica.
2. Cole o **Agent ID** no campo do widget (ou exporte
   `ELEVENLABS_AGENT_ID` para vir preenchido por padrão).
3. Permita o microfone e fale: *"minha internet caiu e quero abrir um
   chamado"*.

### Variáveis de ambiente

| Variável | Para que serve |
|---|---|
| `TRILHA_SECRET` | Assinatura de cookies e flash (obrigatória fora de dev) — gere com `trilha secret` |
| `ELEVENLABS_AGENT_ID` | Agent ID público sugerido no widget |
| `DATA_DIR` | Onde o JSON de atendimentos vive (padrão `data/`) |

Copie `.env.example` para `.env` — que está fora do git — e preencha.

## API

| Rota | Método | O que faz |
|---|---|---|
| `/api/tools/abrir_chamado_tecnico` | POST | Abre o chamado técnico (nome, CPF, problema, descrição, prioridade) |
| `/api/tools/enviar_segunda_via_boleto` | POST | Localiza o cliente pelo CPF e registra o envio da segunda via |
| `/api/tools/solicitar_religue_confirmacao` | POST | Registra o religue — só com `pagamento_confirmado=true` |
| `/api/chamados` | GET | Listagem JSON (paginação, busca, filtros) |
| `/api/chamados/{id}/status` | POST | Mudança de status pelos formulários do painel (com CSRF) |

As respostas das ferramentas são pensadas para o agente falar: em vez de
error HTTP para falha de negócio, voltam com `status` explicativo
(`cpf_invalido`, `cliente_nao_encontrado`, `comprovacao_pendente`,
`sem_fatura_aberta`, `sem_debitos`) e a mensagem pronta.

## Segurança (postura OWASP para repo público e VPS)

- **Nenhum segredo no código** — tudo via ambiente; `.env` fora do git,
  `.env.example` documentando o contrato.
- **Dados 100% fictícios** — clientes e atendimentos semeados são exemplos
  clássicos de documentação (CPF inválidos no mundo real); `data/` gerado em
  execução não é versionado.
- **CSP apertada** — `default-src 'self'`, sem `unsafe-inline` em script; as
  únicas origens extras são as da conversa de voz: REST/token
  (`api.elevenlabs.io`) e sinalização WebRTC LiveKit
  (`livekit.rtc.elevenlabs.io`), sempre em `connect-src`.
- **Permissions-Policy mínimo** — `microphone=(self)` apenas, porque o
  widget precisa do microfone na mesma origem.
- **CSRF** nas escritas do painel (`KindPage` no ramo `/api/chamados`);
  redirect de volta aceita só caminho local.
- **Limitador de taxa** por IP (janela fixa, em memória) nos endpoints
  públicos de ferramenta e listagem; leitura de `X-Forwarded-For` correta
  atrás de proxy.
- **Validação de entrada** em toda ferramenta (CPF com dígitos
  verificadores, limites de tamanho, enums fechados) e corpo lido com
  tolerância a campos desconhecidos — o webhook da ElevenLabs manda envelope
  extra.
- **Headers de plataforma** do Trilha: `X-Content-Type-Options`,
  `X-Frame-Options: DENY`, HSTS em HTTPS, COOP, Referrer-Policy, escape de
  HTML por padrão, limite de corpo, cookies assinados.
- **Logs sem PII** — as rotas logam só protocolo, nunca CPF ou descrição.
- O painel é aberto de propósito na POC; em produção, ponha-o atrás de
  autenticação (a receita `trilha add login` do próprio framework) ou de um
  proxy com SSO.

## Deploy em VPS

Binário único (public/ embutido) com a própria CLI:

```bash
trilha check && trilha build
TRILHA_SECRET=$(trilha secret) DATA_DIR=/var/lib/wavehub/data ./bin/voice-based-customer-service
```

Ou com o Dockerfile pronto (build roda `trilha check` como portão,
runtime sem root e volume para os dados):

```bash
docker build -t voice-based-customer-service .
docker run -p 3000:3000 -v wavehub-data:/var/lib/app/data \
  -e TRILHA_SECRET="$(trilha secret)" voice-based-customer-service
```

Na frente, HTTPS com Caddy (`reverse_proxy localhost:3000`) — o
`X-Forwarded-For` que o limitador de taxa lê vem daí. Microfone exige
contexto seguro: `localhost` ou HTTPS.

## Estrutura

```
app/                          # uma pasta = uma URL
├── page.go                   # / — demonstração + widget de voz
├── agente/page.go            # /agente — guia da ElevenLabs
├── painel/                   # /painel — central de atendimentos (ui.Shell)
├── api/tools/…               # as três ferramentas do agente
└── api/chamados/…            # listagem JSON + mudança de status
internal/atendimento/         # modelo, store JSON, CPF, regras dos fluxos
internal/limite/              # limitador de taxa por IP
internal/apiutil/             # corpo JSON tolerante + rate limit
public/                       # style.css (tema WaveHub), voice.js, copy.js
public/vendor/                # @elevenlabs/client local, sha256 no vendor.lock
```

## Roteiro da demonstração

| Cliente fictício | CPF | Cenário |
|---|---|---|
| Ana Beatriz Souza | `111.444.777-35` | Fatura vencida — religue com comprovação |
| Carlos Eduardo Ramos | `529.982.247-25` | Fatura em dia — chamado técnico |
| Marina Lopes | `168.995.350-09` | Fatura a vencer — segunda via |

Deixe o `/painel` aberto em outra janela: cada ferramenta acionada pela voz
aparece lá em segundos (poll de 6s).

## Créditos

- [Trilha](https://github.com/emersonjoe/trilha) — framework Go
  ([documentação](https://emersonjoe.github.io/trilha), [receitas](https://emersonjoe.github.io/trilha/pt/receitas),
  [referência](https://emersonjoe.github.io/trilha/pt/referencia)) e os
  projetos de exemplo do repositório (blog, cadastro, orçamento, assistente).
- [ElevenLabs](https://elevenlabs.io) — plataforma de agentes de voz e SDK
  `@elevenlabs/client`.
- [WaveHub](https://wavehub.dev) — identidade visual e o caso de negócio.

---
id: TASK-003
title: rota GET /api/voz/cartesia cunhando access token
status: ready
spec: 001-provedor-de-voz-selecionavel-elevenlabs-cartesia
milestone: m1
covers:
  - R3
depends_on: []
acceptance:
  - 503 sem chave, 502 upstream, 200 com access_token curto
  - token nunca em log; rota rate-limited
checks:
  - go test ./app/api/voz/cartesia/ && trilha check
created: "2026-10-05T13:01:46Z"
updated: "2026-10-05T13:03:17Z"
---

### Objetivo

Rota servidor que cunha o access token da Cartesia (grant `agent`), espelhando o desenho de `app/api/voz/assinada/route.go` (a rota ElevenLabs). Ler antes: `app/api/voz/assinada/route.go` inteiro — é o modelo.

### Arquivos novos

- `app/api/voz/cartesia/route.go` — pacote `cartesia`, função `GET(c *trilha.Ctx) error`
- `app/api/voz/cartesia/route_test.go` — testes de tabela do validador

### Contrate da rota

`GET /api/voz/cartesia?agente=<agent_id opcional>`

Respostas:
- 200 `{"disponivel":true,"access_token":"...","agent_id":"...","expires_in":120}`
- 503 sem `CARTESIA_API_KEY`: `{"disponivel":false,"mensagem":"Servidor sem CARTESIA_API_KEY: o provedor Cartesia não está disponível; volte para a ElevenLabs."}`
- 400 `agente` com formato inválido: `{"disponivel":false,"mensagem":"Agent ID com formato inesperado."}`
- 502 upstream falhou (rede): `{"disponivel":false,"mensagem":"Não consegui falar com a Cartesia para cunhar o token. Tente de novo."}`
- 502 upstream respondeu não-200: `{"disponivel":false,"mensagem":"A Cartesia recusou o token (status NNN). Confira a chave e se o agente existe nesta conta."}`

### Implementação (seguir de perto)

```go
package cartesia

import (
    "io"
    "net/http"
    "net/url"
    "os"
    "strconv"
    "strings"
    "time"

    "github.com/emersonjoe/trilha"

    "github.com/emersonjoe/voice-based-customer-service/internal/apiutil"
    "github.com/emersonjoe/voice-based-customer-service/internal/limite"
)

const (
    endpoint   = "https://api.cartesia.ai/access-token"
    versao     = "2026-08-14" // header Cartesia-Version
    expiraEm   = 120          // segundos: curto por desenho
)

var cliente = &http.Client{Timeout: 10 * time.Second}

func GET(c *trilha.Ctx) error {
    if !apiutil.Limita(c.Writer(), c.Request(), trilha.Use[*limite.Limitador](c), "cartesia-token") {
        return nil
    }
    chave := os.Getenv("CARTESIA_API_KEY")
    if chave == "" { /* 503 como no contrato acima; return nil */ }
    agente := c.Query("agente")
    if agente == "" { agente = os.Getenv("CARTESIA_AGENT_ID") }
    if !agenteValido(agente) { /* 400 */ }

    corpo := strings.NewReader(`{"grants":{"agent":true},"expires_in":` + strconv.Itoa(expiraEm) + `}`)
    req, err := http.NewRequestWithContext(c.Context(), http.MethodPost, endpoint, corpo)
    if err != nil { return err }
    req.Header.Set("Authorization", "Bearer "+chave)
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("Cartesia-Version", versao)

    resp, err := cliente.Do(req)
    if err != nil { c.Log().Warn("voz: Cartesia indisponível ao cunhar token"); /* 502 rede */ }
    defer resp.Body.Close()
    bruto, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<12))
    if resp.StatusCode != http.StatusOK { c.Log().Warn("voz: Cartesia recusou token", "status", resp.StatusCode); /* 502 status NNN */ }

    var respostaJSON struct {
        Token       string `json:"token"`
        AccessToken string `json:"access_token"`
    }
    if err := json.Unmarshal(bruto, &respostaJSON); err != nil { /* 502 resposta inesperada */ }
    token := respostaJSON.Token
    if token == "" { token = respostaJSON.AccessToken } // tolerância ao nome do campo
    if token == "" { /* 502 resposta inesperada */ }

    return c.JSON(http.StatusOK, map[string]any{
        "disponivel":   true,
        "access_token": token, // credencial curta: não logar
        "agent_id":     agente,
        "expires_in":   expiraEm,
    })
}

// agenteValido: 8–80 caracteres de [A-Za-z0-9_-].
func agenteValido(agente string) bool { /* mesma forma do validar da rota assinada */ }
```

Preencher os comentários `/* ... */` com `c.JSON(...)` + `return nil` no formato do contrato. IMPORTAR `encoding/json`.

### Testes (`route_test.go`, mesmo pacote)

Tabela para `agenteValido`: ok `"abc123xy"`, `"agent_2901m32wxnddfaat2bzqtsbrn6zd"`, `strings.Repeat("a",80)`; falha `""`, `"ab"`, `"a b"`, `"a.b"`, `strings.Repeat("a",81)`, string com `/`.

### Verificação

1. `go test ./app/api/voz/cartesia/`
2. `trilha gen && trilha check --fix`
3. Com `./dev.sh` rodando SEM a chave: `curl -s localhost:3210/api/voz/cartesia` → JSON com `disponivel:false` e status 503 (`curl -s -o /dev/null -w "%{http_code}"`).
4. (Opcional, dono do repo) chave dummy na env e subir: deve dar 502 com "recusou o token".

### Critérios de aceite

- Contrato de resposta exatamente como acima; token nunca em log.
- Rota rate-limited (mesmo `Limitador`).
- `trilha check` verde.

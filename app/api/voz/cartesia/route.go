// O access token da Cartesia para agentes privados: o widget pede uma
// credencial curta-viva ao servidor, que a cunha com a chave de API. A
// chave nunca chega ao navegador.
package cartesia

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/emersonjoe/trilha"

	"github.com/emersonjoe/voice-based-customer-service/internal/apiutil"
	"github.com/emersonjoe/voice-based-customer-service/internal/limite"
)

const (
	endpoint = "https://api.cartesia.ai/access-token"
	versao   = "2026-08-14" // header Cartesia-Version
	expiraEm = 120          // segundos: curto por desenho
)

var cliente = &http.Client{Timeout: 10 * time.Second}

func GET(c *trilha.Ctx) error {
	if !apiutil.Limita(c.Writer(), c.Request(), trilha.Use[*limite.Limitador](c), "cartesia-token") {
		return nil
	}

	chave := os.Getenv("CARTESIA_API_KEY")
	if chave == "" {
		return c.JSON(http.StatusServiceUnavailable, map[string]any{
			"disponivel": false,
			"mensagem":   "Servidor sem CARTESIA_API_KEY: o provedor Cartesia não está disponível; volte para a ElevenLabs.",
		})
	}

	agente := c.Query("agente")
	if agente == "" {
		agente = os.Getenv("CARTESIA_AGENT_ID")
	}
	if !agenteValido(agente) {
		return c.JSON(http.StatusBadRequest, map[string]any{
			"disponivel": false,
			"mensagem":   "Agent ID com formato inesperado.",
		})
	}

	corpo := strings.NewReader(`{"grants":{"agent":true},"expires_in":` + strconv.Itoa(expiraEm) + `}`)
	req, err := http.NewRequestWithContext(c.Context(), http.MethodPost, endpoint, corpo)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+chave)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Cartesia-Version", versao)

	resp, err := cliente.Do(req)
	if err != nil {
		c.Log().Warn("voz: Cartesia indisponível ao cunhar token")
		return c.JSON(http.StatusBadGateway, map[string]any{
			"disponivel": false,
			"mensagem":   "Não consegui falar com a Cartesia para cunhar o token. Tente de novo.",
		})
	}
	defer resp.Body.Close()

	bruto, err := io.ReadAll(io.LimitReader(resp.Body, 1<<12))
	if err != nil || resp.StatusCode != http.StatusOK {
		c.Log().Warn("voz: Cartesia recusou o token", "status", resp.StatusCode)
		return c.JSON(http.StatusBadGateway, map[string]any{
			"disponivel": false,
			"mensagem":   "A Cartesia recusou o token (status " + strconv.Itoa(resp.StatusCode) + "). Confira a chave e se o agente existe nesta conta.",
		})
	}

	var respostaJSON struct {
		Token       string `json:"token"`
		AccessToken string `json:"access_token"`
	}
	if err := json.Unmarshal(bruto, &respostaJSON); err != nil {
		return c.JSON(http.StatusBadGateway, map[string]any{
			"disponivel": false,
			"mensagem":   "Resposta inesperada da Cartesia ao cunhar o token.",
		})
	}
	token := respostaJSON.Token
	if token == "" {
		token = respostaJSON.AccessToken
	}
	if token == "" {
		return c.JSON(http.StatusBadGateway, map[string]any{
			"disponivel": false,
			"mensagem":   "Resposta inesperada da Cartesia ao cunhar o token.",
		})
	}

	// O token é uma credencial curta-viva: não vai para o log, só para o
	// widget.
	return c.JSON(http.StatusOK, map[string]any{
		"disponivel":   true,
		"access_token": token,
		"agent_id":     agente,
		"expires_in":   expiraEm,
	})
}

// agenteValido limita o Agent ID a um formato validável antes de ir na
// query para a Cartesia.
func agenteValido(agente string) bool {
	if len(agente) < 8 || len(agente) > 80 {
		return false
	}
	for _, r := range agente {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '_', r == '-':
		default:
			return false
		}
	}
	return true
}

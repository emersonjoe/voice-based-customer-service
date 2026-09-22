// A URL de sessão assinada para agentes privados da ElevenLabs: o widget
// pede uma URL curta-viva ao servidor, que assina com a chave de API. A
// chave nunca chega ao navegador.
package assinada

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/emersonjoe/trilha"

	"github.com/emersonjoe/voice-based-customer-service/internal/apiutil"
	"github.com/emersonjoe/voice-based-customer-service/internal/limite"
)

const endpoint = "https://api.elevenlabs.io/v1/convai/conversation/get-signed-url"

var cliente = &http.Client{Timeout: 10 * time.Second}

func GET(c *trilha.Ctx) error {
	if !apiutil.Limita(c.Writer(), c.Request(), trilha.Use[*limite.Limitador](c), "voz-assinada") {
		return nil
	}

	chave := os.Getenv("ELEVENLABS_API_KEY")
	if chave == "" {
		return c.JSON(http.StatusServiceUnavailable, map[string]any{
			"disponivel": false,
			"mensagem":   "Servidor sem ELEVENLABS_API_KEY: a sessão assinada não está disponível; o widget tenta o modo público.",
		})
	}

	agente := c.Query("agente")
	if agente == "" {
		agente = os.Getenv("ELEVENLABS_AGENT_ID")
	}
	if !agenteValido(agente) {
		return c.JSON(http.StatusBadRequest, map[string]any{
			"disponivel": false,
			"mensagem":   "Agent ID com formato inesperado.",
		})
	}

	req, err := http.NewRequestWithContext(c.Context(), http.MethodGet,
		endpoint+"?agent_id="+url.QueryEscape(agente), nil)
	if err != nil {
		return err
	}
	req.Header.Set("xi-api-key", chave)
	resp, err := cliente.Do(req)
	if err != nil {
		c.Log().Warn("voz: ElevenLabs indisponível ao assinar sessão")
		return c.JSON(http.StatusBadGateway, map[string]any{
			"disponivel": false,
			"mensagem":   "Não consegui falar com a ElevenLabs para assinar a sessão. Tente de novo.",
		})
	}
	defer resp.Body.Close()

	corpo, err := io.ReadAll(io.LimitReader(resp.Body, 1<<12))
	if err != nil || resp.StatusCode != http.StatusOK {
		c.Log().Warn("voz: assinatura recusada pela ElevenLabs", "status", resp.StatusCode)
		return c.JSON(http.StatusBadGateway, map[string]any{
			"disponivel": false,
			"mensagem":   "A ElevenLabs recusou a assinatura da sessão (status " + itoa(resp.StatusCode) + "). Confira se o agente existe nesta conta.",
		})
	}

	var corpoJSON struct {
		SignedURL string `json:"signed_url"`
	}
	if err := json.Unmarshal(corpo, &corpoJSON); err != nil || corpoJSON.SignedURL == "" {
		return c.JSON(http.StatusBadGateway, map[string]any{
			"disponivel": false,
			"mensagem":   "Resposta inesperada da ElevenLabs ao assinar a sessão.",
		})
	}

	// A URL é curta-viva e vale como credencial de conversa: não vai para o
	// log, só para o widget.
	return c.JSON(http.StatusOK, map[string]any{
		"disponivel": true,
		"signed_url": corpoJSON.SignedURL,
	})
}

// agenteValido limita o Agent ID a um formato validável antes de ir na
// query para a ElevenLabs.
func agenteValido(agente string) bool {
	if len(agente) < 8 || len(agente) > 80 {
		return false
	}
	for _, r := range agente {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '_':
		default:
			return false
		}
	}
	return true
}

func itoa(n int) string {
	return strconv.Itoa(n)
}

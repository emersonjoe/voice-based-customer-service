// Package app é onde as rotas vivem, e setup.go é o que roda antes da
// primeira requisição.
package app

import (
	"os"
	"path/filepath"
	"time"

	"github.com/emersonjoe/trilha"

	"github.com/emersonjoe/voice-based-customer-service/internal/atendimento"
	"github.com/emersonjoe/voice-based-customer-service/internal/limite"
)

// Setup roda uma vez, antes da primeira requisição: é onde as dependências
// ficam disponíveis com trilha.Provide e a configuração que não é flag é
// ajustada. Nenhum segredo mora no código: tudo vem de ambiente.
func Setup(a *trilha.App) error {
	cfg := a.Config()

	cfg.Locales = []string{"pt-BR"}

	// O widget de voz usa o microfone na mesma origem; a sessão de conversa
	// fala com a ElevenLabs: REST/token em api.elevenlabs.io e sinalização
	// WebRTC (LiveKit) em livekit.rtc.elevenlabs.io, por WSS e HTTPS. Os
	// worklets de áudio do SDK são blobs (código inline dele) e o resampler
	// está self-hosted em public/vendor. CSP e Permissions ficam no mínimo
	// necessário — nada de curingas nem CDN.
	cfg.Security.PermissionsPolicy = "camera=(), microphone=(self), display-capture=(self), geolocation=(), payment=(), usb=()"
	cfg.Security.CSPExtra = map[string][]string{
		"connect-src": {
			"wss://api.elevenlabs.io",
			"https://api.elevenlabs.io",
			"wss://livekit.rtc.elevenlabs.io",
			"https://livekit.rtc.elevenlabs.io",
		},
		"script-src": {"blob:"},
		"worker-src": {"blob:"},
	}

	dir := os.Getenv("DATA_DIR")
	if dir == "" {
		dir = "data"
	}
	store, err := atendimento.Abrir(filepath.Join(dir, "atendimentos.json"))
	if err != nil {
		return err
	}
	trilha.Provide(a, store)

	// Os endpoints de ferramenta são públicos por desenho (a ElevenLabs
	// chama de fora no modo webhook), então eles carregam um limitador de
	// taxa por IP.
	trilha.Provide(a, limite.Novo(30, time.Minute))
	return nil
}

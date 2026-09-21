// Package limite implementa um limitador de taxa por chave (IP) em janela
// fixa, todo em memória: o suficiente para uma POC exposta numa VPS não virar
// funil aberto, sem adicionar dependência ao projeto.
package limite

import (
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

type balde struct {
	inicio time.Time
	feitos int
}

// Limitador permite no máximo max chamadas por chave dentro de cada janela.
type Limitador struct {
	mu     sync.Mutex
	max    int
	janela time.Duration
	baldes map[string]*balde
	agora  func() time.Time // injetável nos testes
}

// Novo cria o limitador. max é o teto de chamadas por janela por chave.
func Novo(max int, janela time.Duration) *Limitador {
	return &Limitador{
		max:    max,
		janela: janela,
		baldes: map[string]*balde{},
		agora:  time.Now,
	}
}

// Permite contabiliza a chamada da chave e diz se ela passa. Chaves velhas
// são esquecidas conforme o mapa cresce.
func (l *Limitador) Permite(chave string) bool {
	agora := l.agora()
	l.mu.Lock()
	defer l.mu.Unlock()

	if len(l.baldes) > 8192 {
		for k, b := range l.baldes {
			if agora.Sub(b.inicio) >= l.janela {
				delete(l.baldes, k)
			}
		}
	}

	b, ok := l.baldes[chave]
	if !ok || agora.Sub(b.inicio) >= l.janela {
		l.baldes[chave] = &balde{inicio: agora, feitos: 1}
		return true
	}
	if b.feitos >= l.max {
		return false
	}
	b.feitos++
	return true
}

// ChaveIP escolhe a chave do limitador para uma requisição: o primeiro salto
// do X-Forwarded-For quando a app roda atrás de proxy (Caddy/nginx na VPS),
// e o RemoteAddr nos outros casos.
func ChaveIP(r *http.Request) string {
	if xf := r.Header.Get("X-Forwarded-For"); xf != "" {
		if primeiro, _, _ := strings.Cut(xf, ","); primeiro != "" {
			return strings.TrimSpace(primeiro)
		}
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

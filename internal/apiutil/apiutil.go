// Package apiutil reúne o que as rotas de ferramenta compartilham: leitura
// tolerante do corpo JSON e a chave do limitador de taxa.
package apiutil

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/emersonjoe/voice-based-customer-service/internal/atendimento"
	"github.com/emersonjoe/voice-based-customer-service/internal/limite"
)

// CPFDaFerramenta normaliza o CPF recebido e, quando ele veio errado da voz
// (dígito perdido, repetido ou trocado), tenta repará-lo contra o cadastro.
// O booleano diz se houve reparo — a resposta deve pedir confirmação ao
// cliente antes de seguir.
func CPFDaFerramenta(corpo map[string]any, store *atendimento.Store) (string, bool) {
	cpf := atendimento.SoDigitos(Texto(corpo, "cpf"))
	if atendimento.CPFValido(cpf) {
		return cpf, false
	}
	if consertado := store.ReparaCPF(cpf); consertado != "" {
		return consertado, true
	}
	return cpf, false
}

// Corpo decodifica o JSON da requisição num mapa, ignorando campos
// desconhecidos. A tolerância é de propósito: o mesmo endpoint recebe
// escrita do browser (client tool) e da ElevenLabs (webhook), e o webhook
// vem com campos de envelope que não são parâmetro da ferramenta.
func Corpo(r *http.Request) (map[string]any, error) {
	raw, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()
	if len(strings.TrimSpace(string(raw))) == 0 {
		return map[string]any{}, nil
	}
	var m map[string]any
	if err := json.Unmarshal(raw, &m); err != nil {
		return nil, err
	}
	return m, nil
}

// Texto extrai um campo textual, aceitando que o remetente mande número ou
// booleano onde a ferramenta espera string.
func Texto(m map[string]any, k string) string {
	v, ok := m[k]
	if !ok || v == nil {
		return ""
	}
	switch t := v.(type) {
	case string:
		return strings.TrimSpace(t)
	case float64:
		return strconv.FormatFloat(t, 'f', -1, 64)
	case bool:
		return strconv.FormatBool(t)
	default:
		return fmt.Sprintf("%v", t)
	}
}

// Bool extrai um campo booleano, aceitando "true"/"sim"/"1".
func Bool(m map[string]any, k string) bool {
	switch strings.ToLower(Texto(m, k)) {
	case "true", "1", "sim", "yes", "confirmado":
		return true
	}
	return false
}

// Limita aplica o limitador de taxa ao IP da requisição e devolve a resposta
// 429 pronta quando a cota estourou.
func Limita(w http.ResponseWriter, r *http.Request, l *limite.Limitador, balde string) bool {
	if l.Permite(balde + ":" + limite.ChaveIP(r)) {
		return true
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Retry-After", "60")
	w.WriteHeader(http.StatusTooManyRequests)
	_, _ = w.Write([]byte(`{"status":"muitas_solicitacoes","mensagem":"Muitas solicitações seguidas. Aguarde um minuto e tente de novo."}`))
	return false
}

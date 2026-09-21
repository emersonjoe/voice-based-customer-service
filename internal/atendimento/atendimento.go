// Package atendimento guarda o modelo e o repositório da POC: uma base
// fictícia de clientes e os atendimentos (chamado técnico, segunda via de
// boleto e religue) que o agente de voz abre. A persistência é um arquivo
// JSON — zero dependências, como o resto do projeto — e todo dado semeado
// aqui é inventado, para o repositório público não carregar informação real.
package atendimento

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/emersonjoe/trilha"
)

// Tipo classifica o atendimento pelos três fluxos da POC.
type Tipo string

const (
	TipoChamadoTecnico Tipo = "chamado_tecnico"
	TipoSegundaVia     Tipo = "segunda_via"
	TipoReligue        Tipo = "religue"
)

// Validos são os tipos que a API aceita.
var Validos = []Tipo{TipoChamadoTecnico, TipoSegundaVia, TipoReligue}

// Validodevolve se o tipo existe.
func (t Tipo) Valido() bool { return slices.Contains(Validos, t) }

// Prefixo do protocolo: CH-2026-0001, BV-2026-0001, RG-2026-0001.
func (t Tipo) Prefixo() string {
	switch t {
	case TipoChamadoTecnico:
		return "CH"
	case TipoSegundaVia:
		return "BV"
	case TipoReligue:
		return "RG"
	}
	return "AT"
}

// Rotulo é o nome do tipo como a tela mostra.
func (t Tipo) Rotulo() string {
	switch t {
	case TipoChamadoTecnico:
		return "Chamado técnico"
	case TipoSegundaVia:
		return "Segunda via de boleto"
	case TipoReligue:
		return "Religue"
	}
	return string(t)
}

// Status é o ciclo de vida do atendimento.
type Status string

const (
	StatusAberto        Status = "aberto"
	StatusEmAtendimento Status = "em_atendimento"
	StatusResolvido     Status = "resolvido"
)

// StatusValidos são os status que a API do painel aceita.
var StatusValidos = []Status{StatusAberto, StatusEmAtendimento, StatusResolvido}

// Valido devolve se o status existe.
func (s Status) Valido() bool { return slices.Contains(StatusValidos, s) }

// Rotulo é o nome do status como a tela mostra.
func (s Status) Rotulo() string {
	switch s {
	case StatusAberto:
		return "Aberto"
	case StatusEmAtendimento:
		return "Em atendimento"
	case StatusResolvido:
		return "Resolvido"
	}
	return string(s)
}

// Prioridade é a fila do chamado técnico.
type Prioridade string

const (
	PrioridadeBaixa Prioridade = "baixa"
	PrioridadeMedia Prioridade = "media"
	PrioridadeAlta  Prioridade = "alta"
)

// PrioridadesValidas são as prioridades que a API aceita.
var PrioridadesValidas = []Prioridade{PrioridadeBaixa, PrioridadeMedia, PrioridadeAlta}

// Valido devolve se a prioridade existe.
func (p Prioridade) Valido() bool { return slices.Contains(PrioridadesValidas, p) }

func (p Prioridade) Rotulo() string {
	switch p {
	case PrioridadeBaixa:
		return "Baixa"
	case PrioridadeMedia:
		return "Média"
	case PrioridadeAlta:
		return "Alta"
	}
	return string(p)
}

// rank ordena prioridade de alta para baixa.
func (p Prioridade) rank() int {
	switch p {
	case PrioridadeAlta:
		return 0
	case PrioridadeMedia:
		return 1
	}
	return 2
}

// Fatura é uma cobrança do cliente. Na POC há no máximo duas por cliente.
type Fatura struct {
	ValorCentavos  int64     `json:"valor_centavos"`
	Vencimento     time.Time `json:"vencimento"`
	Paga           bool      `json:"paga"`
	LinhaDigitavel string    `json:"linha_digitavel"`
}

// Vencida diz se a fatura em aberto passou do vencimento.
func (f Fatura) Vencida(agora time.Time) bool { return !f.Paga && f.Vencimento.Before(agora) }

// Cliente é a base fictícia da demo. O CPF (só dígitos) é a chave.
type Cliente struct {
	CPF      string   `json:"cpf"`
	Nome     string   `json:"nome"`
	Email    string   `json:"email"`
	Telefone string   `json:"telefone"`
	Plano    string   `json:"plano"`
	Endereco string   `json:"endereco"`
	Faturas  []Fatura `json:"faturas"`
}

// Atendimento é o registro que o agente de voz (ou o painel) cria.
type Atendimento struct {
	ID           string            `json:"id"`
	Protocolo    string            `json:"protocolo"`
	Tipo         Tipo              `json:"tipo"`
	ClienteCPF   string            `json:"cliente_cpf"`
	ClienteNome  string            `json:"cliente_nome"`
	Resumo       string            `json:"resumo"`
	Descricao    string            `json:"descricao"`
	Prioridade   Prioridade        `json:"prioridade"`
	Status       Status            `json:"status"`
	Origem       string            `json:"origem"` // voz | painel | webhook
	ConversaID   string            `json:"conversa_id,omitempty"`
	Detalhes     map[string]string `json:"detalhes,omitempty"`
	CriadoEm     time.Time         `json:"criado_em"`
	AtualizadoEm time.Time         `json:"atualizado_em"`
}

// Filtro carrega busca, facetas e página da listagem. O ListParams embutido é
// o contrato do ui.DataTable: página, ordem e busca vêm da URL.
type Filtro struct {
	trilha.ListParams
	Tipo   Tipo
	Status Status
}

// Store é o repositório: carrega de data/atendimentos.json, serve em memória
// e grava atomicamente a cada escrita.
type Store struct {
	mu           sync.Mutex
	path         string
	Seq          map[string]int      `json:"seq"`
	Clientes     map[string]*Cliente `json:"clientes"`
	Atendimentos []*Atendimento      `json:"atendimentos"`
}

// Abrir carrega o repositório do caminho, semeando a base de demonstração
// quando o arquivo ainda não existe.
func Abrir(path string) (*Store, error) {
	s := &Store{path: path, Seq: map[string]int{}}
	raw, err := os.ReadFile(path)
	switch {
	case os.IsNotExist(err):
		s.semear()
		return s, s.salvar()
	case err != nil:
		return nil, fmt.Errorf("atendimento: lendo %s: %w", path, err)
	}
	if err := json.Unmarshal(raw, s); err != nil {
		return nil, fmt.Errorf("atendimento: %s corrompido: %w", path, err)
	}
	if s.Clientes == nil {
		s.Clientes = map[string]*Cliente{}
	}
	if s.Seq == nil {
		s.Seq = map[string]int{}
	}
	return s, nil
}

// salvar grava com escrita temporária + rename, para um crash no meio da
// escrita não corromper o arquivo.
func (s *Store) salvar() error {
	raw, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	tmp := s.path + ".tmp"
	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(tmp, raw, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, s.path)
}

// Criar registra um atendimento e devolve a cópia criada.
func (s *Store) Criar(t Tipo, cpf, resumo, descricao string, p Prioridade, origem, conversaID string, detalhes map[string]string) *Atendimento {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.Seq == nil {
		s.Seq = map[string]int{}
	}
	cpf = SoDigitos(cpf)
	s.Seq[t.Prefixo()]++
	ano, _, _ := time.Now().Date()
	att := &Atendimento{
		ID:           strconv.FormatInt(time.Now().UnixNano(), 36),
		Protocolo:    fmt.Sprintf("%s-%d-%04d", t.Prefixo(), ano, s.Seq[t.Prefixo()]),
		Tipo:         t,
		ClienteCPF:   cpf,
		ClienteNome:  s.nomeDe(cpf),
		Resumo:       resumo,
		Descricao:    descricao,
		Prioridade:   p,
		Status:       StatusAberto,
		Origem:       origem,
		ConversaID:   conversaID,
		Detalhes:     detalhes,
		CriadoEm:     time.Now(),
		AtualizadoEm: time.Now(),
	}
	if att.Prioridade == "" {
		att.Prioridade = PrioridadeMedia
	}
	s.Atendimentos = append([]*Atendimento{att}, s.Atendimentos...)
	_ = s.salvar()
	return att
}

// nomeDe consulta o cliente sem tocar o travas — chamada só com o mutex preso.
func (s *Store) nomeDe(cpf string) string {
	if c, ok := s.Clientes[cpf]; ok {
		return c.Nome
	}
	return ""
}

// Listar aplica o filtro, a busca e a ordem da URL e devolve a página pedida
// mais o total antes da paginação.
func (s *Store) Listar(f Filtro) ([]*Atendimento, int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	q := strings.ToLower(strings.TrimSpace(f.Q))
	var todas []*Atendimento
	for _, a := range s.Atendimentos {
		if f.Tipo != "" && a.Tipo != f.Tipo {
			continue
		}
		if f.Status != "" && a.Status != f.Status {
			continue
		}
		if q != "" && !casa(a, q) {
			continue
		}
		todas = append(todas, a)
	}
	total := len(todas)

	ordem := f.Sort
	if ordem == "" {
		ordem = "criado_em"
	}
	desc := f.Dir == "desc" || f.Sort == "" // o padrão da tela é o mais novo primeiro
	rankStatus := map[Status]int{StatusAberto: 0, StatusEmAtendimento: 1, StatusResolvido: 2}
	slices.SortStableFunc(todas, func(a, b *Atendimento) int {
		cmp := 0
		switch ordem {
		case "protocolo":
			cmp = strings.Compare(a.Protocolo, b.Protocolo)
		case "status":
			cmp = rankStatus[a.Status] - rankStatus[b.Status]
		case "prioridade":
			cmp = a.Prioridade.rank() - b.Prioridade.rank()
		default:
			cmp = a.CriadoEm.Compare(b.CriadoEm)
		}
		if desc {
			return -cmp
		}
		return cmp
	})

	fim := f.Offset() + f.PerPage
	if f.PerPage <= 0 { // sem Bind não há defaults: devolve tudo
		fim = total
	}
	if fim > total {
		fim = total
	}
	inicio := f.Offset()
	if inicio < 0 {
		inicio = 0
	}
	if inicio > total {
		return nil, total
	}
	return todas[inicio:fim], total
}

func casa(a *Atendimento, q string) bool {
	campos := []string{a.Protocolo, a.ClienteNome, a.Resumo, a.Descricao, FormataCPF(a.ClienteCPF)}
	for _, c := range campos {
		if strings.Contains(strings.ToLower(c), q) {
			return true
		}
	}
	return false
}

// PorID devolve o atendimento com aquele id.
func (s *Store) PorID(id string) (*Atendimento, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, a := range s.Atendimentos {
		if a.ID == id {
			copia := *a
			return &copia, true
		}
	}
	return nil, false
}

// AtualizarStatus avança o ciclo de vida e grava.
func (s *Store) AtualizarStatus(id string, st Status) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, a := range s.Atendimentos {
		if a.ID != id {
			continue
		}
		a.Status = st
		a.AtualizadoEm = time.Now()
		_ = s.salvar()
		return true
	}
	return false
}

// ContagemPorStatus resume o painel.
func (s *Store) ContagemPorStatus() map[Status]int {
	s.mu.Lock()
	defer s.mu.Unlock()
	m := map[Status]int{}
	for _, a := range s.Atendimentos {
		m[a.Status]++
	}
	return m
}

// ResolvidosHoje conta os resolvidos desde a meia-noite local.
func (s *Store) ResolvidosHoje(agora time.Time) int {
	s.mu.Lock()
	defer s.mu.Unlock()
	inicio := time.Date(agora.Year(), agora.Month(), agora.Day(), 0, 0, 0, 0, agora.Location())
	n := 0
	for _, a := range s.Atendimentos {
		if a.Status == StatusResolvido && a.AtualizadoEm.After(inicio) {
			n++
		}
	}
	return n
}

// Ultimos devolve os n atendimentos mais recentes, para o feed do painel.
func (s *Store) Ultimos(n int) []*Atendimento {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.Atendimentos) < n {
		n = len(s.Atendimentos)
	}
	fora := make([]*Atendimento, 0, n)
	for _, a := range s.Atendimentos[:n] {
		copia := *a
		fora = append(fora, &copia)
	}
	return fora
}

// Total devolve quantos atendimentos existem.
func (s *Store) Total() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.Atendimentos)
}

// ClientePorCPF normaliza o CPF e devolve o cliente da base fictícia.
func (s *Store) ClientePorCPF(cpf string) (*Cliente, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	c, ok := s.Clientes[SoDigitos(cpf)]
	if !ok {
		return nil, false
	}
	copia := *c
	return &copia, true
}

// FaturaAberta devolve a fatura em aberto mais relevante: a vencida, ou a
// mais próxima do vencimento.
func (c *Cliente) FaturaAberta(agora time.Time) (Fatura, bool) {
	var achada *Fatura
	for i := range c.Faturas {
		f := &c.Faturas[i]
		if f.Paga {
			continue
		}
		if achada == nil || f.Vencimento.Before(achada.Vencimento) {
			achada = f
		}
	}
	if achada == nil {
		return Fatura{}, false
	}
	return *achada, true
}

// SoDigitos mantém só os caracteres numéricos de um CPF digitado.
func SoDigitos(s string) string {
	var b strings.Builder
	for _, r := range s {
		if r >= '0' && r <= '9' {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// CPFValido confere os dígitos verificadores e rejeita sequências repetidas.
func CPFValido(s string) bool {
	cpf := SoDigitos(s)
	if len(cpf) != 11 {
		return false
	}
	todosIguais := true
	for i := 1; i < 11; i++ {
		if cpf[i] != cpf[0] {
			todosIguais = false
			break
		}
	}
	if todosIguais {
		return false
	}
	digito := func(n int) byte {
		soma := 0
		for i := 0; i < n; i++ {
			soma += int(cpf[i]-'0') * (n + 1 - i)
		}
		resto := (soma * 10) % 11
		if resto == 10 {
			resto = 0
		}
		return byte('0' + resto)
	}
	return digito(9) == cpf[9] && digito(10) == cpf[10]
}

// FormataCPF põe os pontos e o traço de volta: 000.000.000-00.
func FormataCPF(cpf string) string {
	d := SoDigitos(cpf)
	if len(d) != 11 {
		return cpf
	}
	return d[0:3] + "." + d[3:6] + "." + d[6:9] + "-" + d[9:]
}

// BRL formata centavos como reais: 11990 → "R$ 119,90".
func BRL(centavos int64) string {
	reais := centavos / 100
	cent := centavos % 100
	mil := reais / 1000
	resto := reais % 1000
	if mil > 0 {
		return fmt.Sprintf("R$ %d.%03d,%02d", mil, resto, cent)
	}
	return fmt.Sprintf("R$ %d,%02d", reais, cent)
}

// semear escreve a base fictícia da demo. Os CPFs são os exemplos clássicos
// da documentação de validação — válidos no algoritmo, vazios no mundo real.
func (s *Store) semear() {
	agora := time.Now()
	s.Clientes = map[string]*Cliente{
		"11144477735": {
			CPF: "11144477735", Nome: "Ana Beatriz Souza",
			Email: "ana.souza@example.com", Telefone: "(21) 98888-0101",
			Plano: "Fibra 500 Mega", Endereco: "Rua das Palmeiras, 120 — Rio de Janeiro/RJ",
			Faturas: []Fatura{{
				ValorCentavos:  11990,
				Vencimento:     agora.AddDate(0, 0, -4),
				LinhaDigitavel: "34191.09008 11144.477735 16899.535009 5 91240000011990",
			}},
		},
		"52998224725": {
			CPF: "52998224725", Nome: "Carlos Eduardo Ramos",
			Email: "carlos.ramos@example.com", Telefone: "(21) 97777-0202",
			Plano: "Fibra 700 Mega + WaveTV", Endereco: "Av. Atlântica, 1500 — Niterói/RJ",
			Faturas: []Fatura{{
				ValorCentavos:  14990,
				Vencimento:     agora.AddDate(0, 0, 12),
				Paga:           true,
				LinhaDigitavel: "34191.09008 52998.224725 11144.477735 6 91240000014990",
			}},
		},
		"16899535009": {
			CPF: "16899535009", Nome: "Marina Lopes",
			Email: "marina.lopes@example.com", Telefone: "(21) 96666-0303",
			Plano: "Fibra 250 Mega", Endereco: "Rua do Farol, 45 — Rio de Janeiro/RJ",
			Faturas: []Fatura{{
				ValorCentavos:  8990,
				Vencimento:     agora.AddDate(0, 0, 8),
				LinhaDigitavel: "34191.09008 16899.535009 52998.224725 7 91240000008990",
			}},
		},
	}

	ch := s.Criar(TipoChamadoTecnico, "52998224725", "Sem conexão desde a madrugada",
		"A internet caiu por volta das 3h. A luz do modem fica vermelha e o wi-fi não aparece em nenhum celular da casa.",
		PrioridadeAlta, "voz", "demo-seed-1", map[string]string{"problema": "sem_conexao"})
	ch.Status = StatusEmAtendimento
	ch.AtualizadoEm = agora.Add(-40 * time.Minute)

	bv := s.Criar(TipoSegundaVia, "16899535009", "Segunda via enviada por e-mail",
		"Cliente pediu a segunda via da fatura de setembro por e-mail.",
		PrioridadeBaixa, "voz", "demo-seed-2", map[string]string{
			"canal": "email", "enviado_para": "marina.lopes@example.com",
			"valor": BRL(8990), "vencimento": agora.AddDate(0, 0, 8).Format("02/01/2006"),
		})
	bv.Status = StatusResolvido
	bv.AtualizadoEm = agora.Add(-3 * time.Hour)

	s.Criar(TipoReligue, "11144477735", "Religue solicitado com comprovação de pagamento",
		"Cliente confirmou o pagamento via Pix e enviou o comprovante pelo WhatsApp.",
		PrioridadeAlta, "painel", "", map[string]string{
			"forma_pagamento": "pix", "valor": BRL(11990),
			"vencimento": agora.AddDate(0, 0, -4).Format("02/01/2006"),
		})
}

package atendimento

import (
	"testing"
	"time"

	"github.com/emersonjoe/trilha"
)

func TestCPFValido(t *testing.T) {
	// Os CPFs semeados são exemplos clássicos da documentação: o algoritmo
	// precisa aceitá-los e recusar o resto.
	validos := []string{"11144477735", "529.982.247-25", "168.995.350-09"}
	for _, cpf := range validos {
		if !CPFValido(cpf) {
			t.Errorf("CPFValido(%q) = false; queria true", cpf)
		}
	}
	invalidos := []string{"", "123", "1114447773", "11144477736", "11111111111", "abcdefghijk"}
	for _, cpf := range invalidos {
		if CPFValido(cpf) {
			t.Errorf("CPFValido(%q) = true; queria false", cpf)
		}
	}
}

func TestCPFSemeadosSaoValidos(t *testing.T) {
	s := &Store{}
	s.semear()
	for cpf := range s.Clientes {
		if !CPFValido(cpf) {
			t.Errorf("cliente semeado com CPF inválido: %s", cpf)
		}
	}
	if len(s.Clientes) != 3 {
		t.Errorf("semear gerou %d clientes; queria 3", len(s.Clientes))
	}
}

func TestCriarEListar(t *testing.T) {
	s := &Store{Seq: map[string]int{}}
	s.semear()

	a := s.Criar(TipoChamadoTecnico, "111.444.777-35", "Sem sinal", "Modem apagado.", PrioridadeAlta, "voz", "", nil)
	if a.Protocolo[:3] != "CH-" {
		t.Errorf("protocolo = %s; queria prefixo CH-", a.Protocolo)
	}
	if a.ClienteNome != "Ana Beatriz Souza" {
		t.Errorf("cliente nome = %q; queria o nome da base", a.ClienteNome)
	}

	b := s.Criar(TipoReligue, "11144477735", "Religue", "Pagou e pediu religue.", PrioridadeAlta, "voz", "", nil)
	if b.Protocolo[:3] != "RG-" {
		t.Errorf("protocolo = %s; queria prefixo RG-", b.Protocolo)
	}

	todos, total := s.Listar(Filtro{})
	if total != 5 { // 3 semeados + 2 criados
		t.Errorf("total = %d; queria 5", total)
	}
	if todos[0].ID != b.ID {
		t.Errorf("o mais novo deveria vir primeiro")
	}

	porTipo, n := s.Listar(Filtro{Tipo: TipoChamadoTecnico})
	if n != 2 || porTipo[0].Tipo != TipoChamadoTecnico {
		t.Errorf("filtro por tipo devolveu %d; queria 2", n)
	}

	porBusca, n := s.Listar(Filtro{ListParams: trilha.ListParams{Q: "marina"}})
	if n != 1 || porBusca[0].ClienteNome != "Marina Lopes" {
		t.Errorf("busca por marina devolveu %d; queria 1", n)
	}
}

func TestAtualizarStatus(t *testing.T) {
	s := &Store{Seq: map[string]int{}}
	s.semear()
	a := s.Criar(TipoChamadoTecnico, "52998224725", "Lentidão", "Está arrastando.", PrioridadeMedia, "voz", "", nil)

	if !s.AtualizarStatus(a.ID, StatusEmAtendimento) {
		t.Fatal("atualização falhou para id existente")
	}
	got, ok := s.PorID(a.ID)
	if !ok || got.Status != StatusEmAtendimento {
		t.Errorf("status = %v; queria em_atendimento", got.Status)
	}
	if s.AtualizarStatus("id-inexistente", StatusResolvido) {
		t.Error("atualizou id inexistente")
	}
}

func TestFaturas(t *testing.T) {
	s := &Store{}
	s.semear()
	agora := time.Now()

	ana, ok := s.ClientePorCPF("111.444.777-35")
	if !ok {
		t.Fatal("Ana não encontrada")
	}
	f, ok := ana.FaturaAberta(agora)
	if !ok || !f.Vencida(agora) {
		t.Error("Ana deveria ter fatura vencida em aberto")
	}

	carlos, _ := s.ClientePorCPF("52998224725")
	if _, ok := carlos.FaturaAberta(agora); ok {
		t.Error("Carlos não deveria ter fatura em aberto")
	}
}

func TestBRL(t *testing.T) {
	casos := map[int64]string{
		11990:   "R$ 119,90",
		8990:    "R$ 89,90",
		1000000: "R$ 10.000,00",
		5:       "R$ 0,05",
	}
	for centavos, quer := range casos {
		if got := BRL(centavos); got != quer {
			t.Errorf("BRL(%d) = %q; queria %q", centavos, got, quer)
		}
	}
}

func TestAbrirArquivo(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/atendimentos.json"

	s, err := Abrir(path)
	if err != nil {
		t.Fatalf("Abrir: %v", err)
	}
	s.Criar(TipoSegundaVia, "16899535009", "2ª via", "Por e-mail.", PrioridadeBaixa, "voz", "", nil)

	deNovo, err := Abrir(path)
	if err != nil {
		t.Fatalf("reabrir: %v", err)
	}
	if _, total := deNovo.Listar(Filtro{}); total != 4 { // 3 semeados + 1
		t.Errorf("total após reabrir = %d; queria 4", total)
	}
}

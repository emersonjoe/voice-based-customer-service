package atendimento

import "testing"

func TestReparaCPF(t *testing.T) {
	s := &Store{}
	s.semear()

	// Fala comeu um dígito: 10 dígitos, casa com o cadastro.
	if got := s.ReparaCPF("5299822472"); got != "52998224725" {
		t.Errorf("ReparaCPF(5299822472) = %q; queria 52998224725", got)
	}
	// Dígito verificador ouvido errado.
	if got := s.ReparaCPF("52998224755"); got != "52998224725" {
		t.Errorf("ReparaCPF(52998224755) = %q; queria 52998224725", got)
	}
	// Fala repetiu um dígito: 12 dígitos.
	if got := s.ReparaCPF("529998224725"); got != "52998224725" {
		t.Errorf("ReparaCPF(529998224725) = %q; queria 52998224725", got)
	}
	// CPF já válido passa sem mexer.
	if got := s.ReparaCPF("529.982.247-25"); got != "52998224725" {
		t.Errorf("ReparaCPF válido = %q; queria o mesmo CPF", got)
	}
	// Dígito ouvido errado, CPF no cadastro: o cadastro decide.
	if got := s.ReparaCPF("52998224720"); got != "52998224725" {
		t.Errorf("ReparaCPF(52998224720) = %q; queria 52998224725", got)
	}
	// Fora do cadastro, mas a correção é única: aceita.
	if got := s.ReparaCPF("12345678900"); got != "12345678909" {
		t.Errorf("ReparaCPF(12345678900) = %q; queria 12345678909", got)
	}
	// Fora do cadastro e ambíguo (fala comeu um dígito): não inventa.
	if got := s.ReparaCPF("1234567890"); got != "" {
		t.Errorf("ReparaCPF(1234567890) = %q; queria vazio (ambíguo)", got)
	}
	// Lixo que não vira CPF.
	if got := s.ReparaCPF("123"); got != "" {
		t.Errorf("ReparaCPF(123) = %q; queria vazio", got)
	}
}

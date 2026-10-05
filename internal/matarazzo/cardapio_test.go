package matarazzo

import "testing"

func TestBuscarNoCardapio(t *testing.T) {
	// sem acento e em minúsculas: acha nas duas categorias
	achados := BuscarPratos(RestauranteMataCitta, "carbonara")
	if len(achados) != 2 {
		t.Fatalf("buscar carbonara = %d resultados; queria 2 (pizza e pasta seca)", len(achados))
	}
	if achados[0].Categoria != "Pizze" || achados[1].Categoria != "Pasta seca" {
		t.Errorf("categorias inesperadas: %s, %s", achados[0].Categoria, achados[1].Categoria)
	}

	// acento e capitalização não atrapalham
	if acho := BuscarPratos(RestauranteMataCitta, "CAMARÃO"); len(acho) != 1 {
		t.Errorf("buscar CAMARÃO = %d resultados; queria 1", len(acho))
	}
	if acho := BuscarPratos(RestauranteMataCitta, "tiramisu"); len(acho) != 1 {
		t.Errorf("buscar tiramisu = %d resultados; queria 1", len(acho))
	}

	// prato que não existe
	if achados := BuscarPratos(RestauranteMataCitta, "hambúrguer"); len(achados) != 0 {
		t.Errorf("buscar hambúrguer devolveu %d resultados; queria 0", len(achados))
	}
}

func TestCategoriaPorNome(t *testing.T) {
	if cat, ok := CategoriaPorNome(RestauranteMataCitta, "PASTA FRESCA"); !ok || len(cat.Itens) != 7 {
		t.Errorf("CategoriaPorNome(PASTA FRESCA) ok=%v itens=%d; queria ok e 7", ok, len(cat.Itens))
	}
	if _, ok := CategoriaPorNome(RestauranteMataCitta, "hamburgueria"); ok {
		t.Error("categoria inexistente não deveria ser encontrada")
	}
}

func TestCategoriasCardapio(t *testing.T) {
	if cats := CategoriasCardapio(RestauranteMataCitta); len(cats) != 8 {
		t.Errorf("CategoriasCardapio(mata_citta) = %d categorias; queria 8", len(cats))
	}
	if cats := CategoriasCardapio(RestauranteLavva); len(cats) != 4 {
		t.Errorf("CategoriasCardapio(lavva) = %d categorias; queria 4", len(cats))
	}
}

func TestBuscarNoLavva(t *testing.T) {
	achados := BuscarPratos(RestauranteLavva, "wagyu")
	if len(achados) != 6 {
		t.Errorf("buscar wagyu no LAVVA = %d resultados; queria 6", len(achados))
	}
	if achados := BuscarPratos(RestauranteLavva, "galbi"); len(achados) != 1 {
		t.Errorf("buscar galbi = %d resultados; queria 1", len(achados))
	}
	if cat, ok := CategoriaPorNome(RestauranteLavva, "signature cocktails"); !ok || len(cat.Itens) != 12 {
		t.Errorf("CategoriaPorNome(signature cocktails) ok=%v itens=%d; queria ok e 12", ok, len(cat.Itens))
	}
}

func TestBuscarTolerante(t *testing.T) {
	casos := map[string]int{
		"tiramusu": 1,
		"camarao":  1,
		"bife":     1, // no Mata Città só existe o Bife à Parmegiana
	}
	for termo, quer := range casos {
		if got := len(BuscarPratosTolerante(RestauranteMataCitta, termo)); got != quer {
			t.Errorf("tolerante(%q) = %d resultados; queria %d", termo, got, quer)
		}
	}
}

func TestToleranteCrossRestaurante(t *testing.T) {
	// picana é do LAVVA: no Mata Città não tem, no LAVVA tem 2
	if achados := len(BuscarPratosTolerante(RestauranteMataCitta, "picana")); achados != 0 {
		t.Errorf("picana no Mata Città = %d; queria 0", achados)
	}
	if achados := len(BuscarPratosTolerante(RestauranteLavva, "picana")); achados != 2 {
		t.Errorf("picana no LAVVA = %d; queria 2 (Angus e Wagyu)", achados)
	}
}

func TestResolverLocalEFalas(t *testing.T) {
	casos := map[string]string{
		"lava":       RestauranteLavva,
		"lavva":      RestauranteLavva,
		"o coreano":  RestauranteLavva,
		"mata sita":  RestauranteMataCitta,
		"o italiano": RestauranteMataCitta,
		"rosewood":   "hotel_rosewood",
		"o hotel":    "hotel_rosewood",
	}
	for falado, quer := range casos {
		resolvido, ok := ResolverLocal(falado)
		if !ok || resolvido.ID != quer {
			t.Errorf("ResolverLocal(%q) = %q, %v; queria %q", falado, resolvido.ID, ok, quer)
		}
	}
	if _, ok := ResolverLocal("pizzaria do joão"); ok {
		t.Error("local aleatório não deveria resolver")
	}
}

func TestResolverTopicoFalas(t *testing.T) {
	if topico, ok := ResolverTopico("mata sita"); !ok || topico != TopicoMataCitta {
		t.Errorf("ResolverTopico(mata sita) = %q, %v; queria mata_citta", topico, ok)
	}
	if topico, ok := ResolverTopico("hotel"); !ok || topico != "hotel_rosewood" {
		t.Errorf("ResolverTopico(hotel) = %q, %v; queria hotel_rosewood", topico, ok)
	}
}

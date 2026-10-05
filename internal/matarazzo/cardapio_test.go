package matarazzo

import "testing"

func TestBuscarNoCardapio(t *testing.T) {
	// sem acento e em minúsculas: acha nas duas categorias
	achados := BuscarNoCardapio("carbonara")
	if len(achados) != 2 {
		t.Fatalf("buscar carbonara = %d resultados; queria 2 (pizza e pasta seca)", len(achados))
	}
	if achados[0].Categoria != "Pizze" || achados[1].Categoria != "Pasta seca" {
		t.Errorf("categorias inesperadas: %s, %s", achados[0].Categoria, achados[1].Categoria)
	}

	// acento e capitalização não atrapalham
	if acho := BuscarNoCardapio("CAMARÃO"); len(acho) != 1 {
		t.Errorf("buscar CAMARÃO = %d resultados; queria 1", len(acho))
	}
	if acho := BuscarNoCardapio("tiramisu"); len(acho) != 1 {
		t.Errorf("buscar tiramisu = %d resultados; queria 1", len(acho))
	}

	// prato que não existe
	if achados := BuscarNoCardapio("hambúrguer"); len(achados) != 0 {
		t.Errorf("buscar hambúrguer devolveu %d resultados; queria 0", len(achados))
	}
}

func TestCategoriaPorNome(t *testing.T) {
	if cat, ok := CategoriaPorNome("PASTA FRESCA"); !ok || len(cat.Itens) != 7 {
		t.Errorf("CategoriaPorNome(PASTA FRESCA) ok=%v itens=%d; queria ok e 7", ok, len(cat.Itens))
	}
	if _, ok := CategoriaPorNome("hamburgueria"); ok {
		t.Error("categoria inexistente não deveria ser encontrada")
	}
}

func TestCategoriasCardapio(t *testing.T) {
	cats := CategoriasCardapio()
	if len(cats) != 8 {
		t.Errorf("CategoriasCardapio = %d categorias; queria 8", len(cats))
	}
}

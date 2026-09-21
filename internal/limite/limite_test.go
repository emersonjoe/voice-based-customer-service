package limite

import (
	"net/http/httptest"
	"testing"
	"time"
)

func TestPermite(t *testing.T) {
	agora := time.Now()
	l := Novo(3, time.Minute)
	l.agora = func() time.Time { return agora }

	for i := 0; i < 3; i++ {
		if !l.Permite("10.0.0.1") {
			t.Fatalf("chamada %d deveria passar", i+1)
		}
	}
	if l.Permite("10.0.0.1") {
		t.Error("a 4ª chamada na mesma janela deveria ser barrada")
	}
	if !l.Permite("10.0.0.2") {
		t.Error("outro IP não deveria ser afetado")
	}

	agora = agora.Add(2 * time.Minute)
	if !l.Permite("10.0.0.1") {
		t.Error("janela nova deveria permitir de novo")
	}
}

func TestChaveIP(t *testing.T) {
	r := httptest.NewRequest("POST", "/api/tools/x", nil)
	r.RemoteAddr = "203.0.113.7:51200"
	if got := ChaveIP(r); got != "203.0.113.7" {
		t.Errorf("ChaveIP = %q; queria o host do RemoteAddr", got)
	}

	r.Header.Set("X-Forwarded-For", "198.51.100.9, 10.0.0.1")
	if got := ChaveIP(r); got != "198.51.100.9" {
		t.Errorf("ChaveIP = %q; queria o primeiro salto do X-Forwarded-For", got)
	}
}

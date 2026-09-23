package main

import (
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
)

// servirEm sobe um servidor de teste e aponta PORT para ele.
func servirEm(t *testing.T, status int) {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/health" {
			http.NotFound(w, r)
			return
		}
		w.WriteHeader(status)
	}))
	t.Cleanup(srv.Close)
	_, porta, _ := net.SplitHostPort(srv.Listener.Addr().String())
	t.Setenv("PORT", porta)
}

func TestHealthcheckAPISaudavel(t *testing.T) {
	servirEm(t, http.StatusOK)
	if got := healthcheck(); got != 0 {
		t.Errorf("healthcheck() = %d; esperado 0", got)
	}
}

func TestHealthcheckAPIComErro(t *testing.T) {
	servirEm(t, http.StatusInternalServerError)
	if got := healthcheck(); got != 1 {
		t.Errorf("healthcheck() = %d; esperado 1", got)
	}
}

func TestHealthcheckAPIForaDoAr(t *testing.T) {
	t.Setenv("PORT", "1") // porta sem ninguem escutando
	if got := healthcheck(); got != 1 {
		t.Errorf("healthcheck() = %d; esperado 1", got)
	}
}

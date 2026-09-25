package main

import (
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
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

func TestNovoServidorTemPrazos(t *testing.T) {
	srv := novoServidor(":0", http.NotFoundHandler())
	prazos := map[string]time.Duration{
		"ReadHeaderTimeout": srv.ReadHeaderTimeout,
		"ReadTimeout":       srv.ReadTimeout,
		"WriteTimeout":      srv.WriteTimeout,
		"IdleTimeout":       srv.IdleTimeout,
	}
	for nome, valor := range prazos {
		if valor <= 0 {
			t.Errorf("%s deve ter prazo (slowloris); veio %v", nome, valor)
		}
	}
	if srv.MaxHeaderBytes <= 0 || srv.MaxHeaderBytes > http.DefaultMaxHeaderBytes {
		t.Errorf("MaxHeaderBytes deve ficar abaixo do padrão de 1 MiB; veio %d", srv.MaxHeaderBytes)
	}
}

func TestPoolDoAmbiente(t *testing.T) {
	casos := []struct {
		nome             string
		abertas, ociosas string
		esperado         configPool
	}{
		{"padrão seguro", "", "", configPool{abertas: 20, ociosas: 10}},
		{"ajuste por env", "40", "15", configPool{abertas: 40, ociosas: 15}},
		{"ociosas não passam de abertas", "5", "", configPool{abertas: 5, ociosas: 5}},
		{"zero ociosas é permitido", "", "0", configPool{abertas: 20, ociosas: 0}},
		{"valor inválido fica no padrão", "abc", "-1", configPool{abertas: 20, ociosas: 10}},
		{"abertas zero fica no padrão", "0", "", configPool{abertas: 20, ociosas: 10}},
	}
	for _, c := range casos {
		t.Run(c.nome, func(t *testing.T) {
			t.Setenv("DB_MAX_OPEN_CONNS", c.abertas)
			t.Setenv("DB_MAX_IDLE_CONNS", c.ociosas)
			if got := poolDoAmbiente(); got != c.esperado {
				t.Errorf("poolDoAmbiente() = %+v; esperado %+v", got, c.esperado)
			}
		})
	}
}

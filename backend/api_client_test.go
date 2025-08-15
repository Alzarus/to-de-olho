package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// Testa retry/backoff do CamaraClient
func TestCamaraClient_Retry(t *testing.T) {
	attempts := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 2 { // primeira falha
			w.WriteHeader(500)
			return
		}
		w.WriteHeader(200)
		w.Write([]byte(`{"dados": []}`))
	}))
	defer srv.Close()

	c := NewCamaraClient(srv.URL, 2*time.Second, 5, 5)
	ctx := context.Background()

	_, err := c.FetchDeputados(ctx, "", "", "")
	if err != nil {
		t.Fatalf("esperava sucesso após retry, erro: %v", err)
	}

	if attempts < 2 {
		t.Fatalf("esperava ao menos 2 tentativas, teve %d", attempts)
	}
}

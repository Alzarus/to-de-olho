package senado

import (
	"context"
	"testing"
	"time"
)

// TestContract_ListaSenadores verifica se a API Legislativa retorna dados compativeis com nosso struct
// Este teste realiza uma requisicao REAL para validar o contrato com a API externa.
// Em CI/CD, idealmente seria executado periodicamente, nao a cada push (devido a rate limits).
func TestContract_ListaSenadores(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping contract test in short mode")
	}

	client := NewLegisClient() // Client default
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	senadores, err := client.ListarSenadoresAtuais(ctx)
	if err != nil {
		t.Fatalf("Contrato quebrado ou API indisponivel: %v", err)
	}

	if len(senadores) == 0 {
		t.Errorf("API retornou lista vazia de senadores")
	}

	// Validar campos obrigatorios do primeiro item
	primeiro := senadores[0]
	if primeiro.IdentificacaoParlamentar.CodigoParlamentar == "" {
		t.Errorf("Contrato quebrado: CodigoParlamentar vazio")
	}
	if primeiro.IdentificacaoParlamentar.NomeParlamentar == "" {
		t.Errorf("Contrato quebrado: NomeParlamentar vazio")
	}
}

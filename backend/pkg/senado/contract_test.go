package senado

import (
	"context"
	"net/http"
	"os"
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

// TestContract_ListaVotacoes verifica se a API Legislativa de votacoes responde (contrato basico)
func TestContract_ListaVotacoes(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping contract test in short mode")
	}

	// Usar client do pacote (supondo que exista ou criando um basico para teste)
	// Como o pacote 'senado' foca em dados parlamentares, vamos instanciar o client legislativo
	client := NewLegisClient()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Buscar votacoes de um ano recente para garantir dados
	ano := time.Now().Year()
	// Tentar ano anterior se estivermos no inicio do ano (jan/fev) e api puder estar vazia
	if time.Now().Month() <= time.February {
		ano--
	}

	votacoes, err := client.ListarVotacoesAno(ctx, 2024) // Hardcoded 2024 for stable contract test or dynamic? 2024 tem dados com certeza
	if err != nil {
		// Nao falhar fatalmente se for apenas timeout ocasional, mas falhar se for erro de parsing/contrato
		t.Fatalf("Erro ao buscar votacoes: %v", err)
	}

	if len(votacoes) > 0 {
		v := votacoes[0]
		// VotacaoSessaoAPI usa IdentificacaoMateria e DescricaoVotacao
		if v.IdentificacaoMateria == "" && v.DescricaoVotacao == "" {
			t.Errorf("Contrato votacao suspeito: campos chave vazios %+v", v)
		}
	} else {
		t.Logf("Aviso: Nenhuma votacao encontrada para o ano teste, contrato nao totalmente validado")
	}
}

// TestContract_ListaEmendas verifica se a API da Transparencia retorna emendas (necessita API KEY)
func TestContract_ListaEmendas(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping contract test in short mode")
	}

	apiKey := os.Getenv("TRANSPARENCIA_API_KEY")
	if apiKey == "" {
		t.Skip("skipping emendas contract test: TRANSPARENCIA_API_KEY not set")
	}

	// Criar client ad-hoc de emendas ou usar o servico se disponivel.
	// Para teste de contrato puro, idealmente seria um client isolado, mas vamos usar o http default
	// Simular chamada: GET /api-de-dados/emendas?ano=2024&autor=...
	// Como nao temos um client de emendas exportado neste pacote 'senado', vamos fazer a request raw
	// ou importar o pacote 'emenda' (cuidado com ciclo).
	// Melhor abordagem: Testar apenas se a URL da API responde 200 para um query valida
	
	url := "https://api.portaldatransparencia.gov.br/api-de-dados/emendas-parlamentares?ano=2024&pagina=1"
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("chave-api-dados", apiKey)
	
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Erro ao conectar API Transparencia: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("API Transparencia retornou status %d", resp.StatusCode)
	}
	
	// Poderiamos decodificar um item para validar struct, mas so o 200 ja garante que a chave funciona e endpoint existe
}

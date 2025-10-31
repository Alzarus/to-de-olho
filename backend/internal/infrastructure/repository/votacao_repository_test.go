package repository

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"to-de-olho-backend/internal/domain"
	"to-de-olho-backend/internal/infrastructure/db"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVotacaoRepository(t *testing.T) {
	// Setup do banco de dados de teste
	// Integration tests require a running Postgres test database.
	// Provide connection string via TEST_DATABASE_URL environment variable.
	connStr := os.Getenv("TEST_DATABASE_URL")
	if connStr == "" {
		t.Skip("Skipping votacao repository tests: set TEST_DATABASE_URL to run")
		return
	}

	testDB, err := db.NewPostgreSQL(connStr)
	if err != nil {
		t.Skipf("test database not available: %v", err)
		return
	}
	defer testDB.Close()

	repo := NewVotacaoRepository(testDB.GetDB())
	ctx := context.Background()

	t.Run("CreateVotacao", func(t *testing.T) {
		numericID := int64(12345)
		votacao := &domain.Votacao{
			IDCamara:         fmt.Sprintf("%d", numericID),
			IDVotacaoCamara:  &numericID,
			Titulo:           "PEC da Blindagem Fiscal",
			Ementa:           "Proposta de Emenda à Constituição que estabelece blindagem fiscal...",
			DataVotacao:      time.Now(),
			Aprovacao:        "Aprovada",
			PlacarSim:        250,
			PlacarNao:        150,
			PlacarAbstencao:  50,
			PlacarOutros:     10,
			TipoProposicao:   "PEC",
			NumeroProposicao: "123",
			AnoProposicao:    &[]int{2024}[0],
			Relevancia:       "alta",
			Payload: map[string]interface{}{
				"original_data": map[string]interface{}{
					"id":  12345,
					"uri": "https://dadosabertos.camara.leg.br/api/v2/votacoes/12345",
				},
			},
		}

		err := repo.CreateVotacao(ctx, votacao)
		assert.NoError(t, err)
		assert.NotZero(t, votacao.ID)
		assert.NotZero(t, votacao.CreatedAt)
	})

	t.Run("GetVotacaoByID", func(t *testing.T) {
		// Primeiro, criar uma votação
		numericID := int64(12346)
		votacao := &domain.Votacao{
			IDCamara:         fmt.Sprintf("%d", numericID),
			IDVotacaoCamara:  &numericID,
			Titulo:           "PL do Marco Legal das Startups",
			Ementa:           "Estabelece marco legal para startups no Brasil",
			DataVotacao:      time.Now(),
			Aprovacao:        "Aprovada",
			PlacarSim:        280,
			PlacarNao:        120,
			PlacarAbstencao:  30,
			TipoProposicao:   "PL",
			NumeroProposicao: "456",
			Relevancia:       "média",
			Payload:          map[string]interface{}{"test": true},
		}

		err := repo.CreateVotacao(ctx, votacao)
		require.NoError(t, err)

		// Buscar a votação criada
		found, err := repo.GetVotacaoByID(ctx, votacao.ID)
		assert.NoError(t, err)
		assert.Equal(t, votacao.Titulo, found.Titulo)
		assert.Equal(t, votacao.Ementa, found.Ementa)
		assert.Equal(t, votacao.PlacarSim, found.PlacarSim)
	})

	t.Run("CreateVotoDeputado", func(t *testing.T) {
		// Criar votação primeiro
		numericID := int64(12347)
		votacao := &domain.Votacao{
			IDCamara:        fmt.Sprintf("%d", numericID),
			IDVotacaoCamara: &numericID,
			Titulo:          "Votação de Teste",
			DataVotacao:     time.Now(),
			Aprovacao:       "Aprovada",
			TipoProposicao:  "PL",
			Relevancia:      "baixa",
			Payload:         map[string]interface{}{},
		}
		err := repo.CreateVotacao(ctx, votacao)
		require.NoError(t, err)

		// Criar voto do deputado
		voto := &domain.VotoDeputado{
			IDVotacao:  votacao.ID,
			IDDeputado: 178957, // ID exemplo
			Voto:       "Sim",
			Payload:    map[string]interface{}{"deputado_nome": "João Silva"},
		}

		err = repo.CreateVotoDeputado(ctx, voto)
		assert.NoError(t, err)
		assert.NotZero(t, voto.ID)
	})

	t.Run("CreateOrientacaoPartido", func(t *testing.T) {
		// Criar votação primeiro
		numericID := int64(12348)
		votacao := &domain.Votacao{
			IDCamara:        fmt.Sprintf("%d", numericID),
			IDVotacaoCamara: &numericID,
			Titulo:          "Orientação de Teste",
			DataVotacao:     time.Now(),
			Aprovacao:       "Rejeitada",
			TipoProposicao:  "PEC",
			Relevancia:      "alta",
			Payload:         map[string]interface{}{},
		}
		err := repo.CreateVotacao(ctx, votacao)
		require.NoError(t, err)

		// Criar orientação do partido
		orientacao := &domain.OrientacaoPartido{
			IDVotacao:  votacao.ID,
			Partido:    "PT",
			Orientacao: "Não",
		}

		err = repo.CreateOrientacaoPartido(ctx, orientacao)
		assert.NoError(t, err)
		assert.NotZero(t, orientacao.ID)
	})

	t.Run("ListVotacoes", func(t *testing.T) {
		filtros := domain.FiltrosVotacao{
			Aprovacao: "Aprovada",
		}

		paginacao := domain.Pagination{
			Page:  1,
			Limit: 10,
		}

		votacoes, total, err := repo.ListVotacoes(ctx, filtros, paginacao)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, total, 0)
		assert.LessOrEqual(t, len(votacoes), 10)

		// Verificar se todas as votações retornadas são aprovadas
		for _, v := range votacoes {
			assert.Equal(t, "Aprovada", v.Aprovacao)
		}
	})

	t.Run("GetVotacaoDetalhada", func(t *testing.T) {
		// Criar votação com votos e orientações
		numericID := int64(12349)
		votacao := &domain.Votacao{
			IDCamara:        fmt.Sprintf("%d", numericID),
			IDVotacaoCamara: &numericID,
			Titulo:          "Votação Completa",
			DataVotacao:     time.Now(),
			Aprovacao:       "Aprovada",
			TipoProposicao:  "PL",
			Relevancia:      "alta",
			Payload:         map[string]interface{}{},
		}
		err := repo.CreateVotacao(ctx, votacao)
		require.NoError(t, err)

		// Adicionar alguns votos
		votos := []*domain.VotoDeputado{
			{IDVotacao: votacao.ID, IDDeputado: 1001, Voto: "Sim"},
			{IDVotacao: votacao.ID, IDDeputado: 1002, Voto: "Não"},
			{IDVotacao: votacao.ID, IDDeputado: 1003, Voto: "Abstenção"},
		}

		for _, voto := range votos {
			err = repo.CreateVotoDeputado(ctx, voto)
			require.NoError(t, err)
		}

		// Adicionar orientações
		orientacoes := []*domain.OrientacaoPartido{
			{IDVotacao: votacao.ID, Partido: "PT", Orientacao: "Sim"},
			{IDVotacao: votacao.ID, Partido: "PSDB", Orientacao: "Não"},
		}

		for _, orientacao := range orientacoes {
			err = repo.CreateOrientacaoPartido(ctx, orientacao)
			require.NoError(t, err)
		}

		// Buscar votação detalhada
		detalhada, err := repo.GetVotacaoDetalhada(ctx, votacao.ID)
		assert.NoError(t, err)
		assert.Equal(t, votacao.Titulo, detalhada.Votacao.Titulo)
		assert.Len(t, detalhada.Votos, 3)
		assert.Len(t, detalhada.Orientacoes, 2)
	})

	t.Run("UpsertVotacao", func(t *testing.T) {
		numericID := int64(99999)
		votacao := &domain.Votacao{
			IDCamara:        fmt.Sprintf("%d", numericID),
			IDVotacaoCamara: &numericID,
			Titulo:          "Votação Original",
			DataVotacao:     time.Now(),
			Aprovacao:       "Aprovada",
			PlacarSim:       100,
			TipoProposicao:  "PL",
			Relevancia:      "baixa",
			Payload:         map[string]interface{}{},
		}

		// Primeira inserção
		err := repo.UpsertVotacao(ctx, votacao)
		assert.NoError(t, err)
		originalID := votacao.ID

		// Atualização
		votacao.Titulo = "Votação Atualizada"
		votacao.PlacarSim = 150

		err = repo.UpsertVotacao(ctx, votacao)
		assert.NoError(t, err)
		assert.Equal(t, originalID, votacao.ID) // ID deve ser o mesmo

		// Verificar se foi atualizada
		found, err := repo.GetVotacaoByID(ctx, votacao.ID)
		assert.NoError(t, err)
		assert.Equal(t, "Votação Atualizada", found.Titulo)
		assert.Equal(t, 150, found.PlacarSim)
	})
}

func TestVotacaoValidation(t *testing.T) {
	t.Run("ValidateVotacao", func(t *testing.T) {
		// Votação válida
		numericID := int64(123)
		votacao := &domain.Votacao{
			IDCamara:        fmt.Sprintf("%d", numericID),
			IDVotacaoCamara: &numericID,
			Titulo:          "Título válido",
			DataVotacao:     time.Now(),
			Aprovacao:       "Aprovada",
			Relevancia:      "alta",
			PlacarSim:       10,
			PlacarNao:       5,
			PlacarAbstencao: 2,
		}
		assert.NoError(t, votacao.Validate())

		// Votação inválida - sem ID textual
		votacaoInvalida := *votacao
		votacaoInvalida.IDCamara = ""
		assert.Error(t, votacaoInvalida.Validate())

		// Votação inválida - sem título
		votacaoInvalida = *votacao
		votacaoInvalida.Titulo = ""
		assert.Error(t, votacaoInvalida.Validate())

		// Votação inválida - aprovação inválida
		votacaoInvalida = *votacao
		votacaoInvalida.Aprovacao = "Talvez"
		assert.Error(t, votacaoInvalida.Validate())
	})

	t.Run("ValidateVotoDeputado", func(t *testing.T) {
		voto := &domain.VotoDeputado{
			IDVotacao:  1,
			IDDeputado: 123,
			Voto:       "Sim",
		}
		assert.NoError(t, voto.Validate())

		// Voto inválido
		votoInvalido := *voto
		votoInvalido.Voto = "Talvez"
		assert.Error(t, votoInvalido.Validate())
	})

	t.Run("ValidateOrientacaoPartido", func(t *testing.T) {
		orientacao := &domain.OrientacaoPartido{
			IDVotacao:  1,
			Partido:    "PT",
			Orientacao: "Sim",
		}
		assert.NoError(t, orientacao.Validate())

		// Orientação inválida
		orientacaoInvalida := *orientacao
		orientacaoInvalida.Orientacao = "Talvez"
		assert.Error(t, orientacaoInvalida.Validate())
	})
}

func TestVotacaoAnalytics(t *testing.T) {
	t.Run("TotalVotos", func(t *testing.T) {
		votacao := &domain.Votacao{
			PlacarSim:       100,
			PlacarNao:       50,
			PlacarAbstencao: 25,
			PlacarOutros:    5,
		}
		assert.Equal(t, 180, votacao.TotalVotos())
	})

	t.Run("PorcentagemAprovacao", func(t *testing.T) {
		votacao := &domain.Votacao{
			PlacarSim:       60,
			PlacarNao:       30,
			PlacarAbstencao: 10,
			PlacarOutros:    0,
		}
		assert.Equal(t, 60.0, votacao.PorcentagemAprovacao())
	})

	t.Run("CalcularDisciplinaPartidaria", func(t *testing.T) {
		vp := &domain.VotacaoPartido{
			Partido:          "PT",
			Orientacao:       "Sim",
			VotaramFavor:     8,
			VotaramContra:    1,
			VotaramAbstencao: 1,
			TotalMembros:     10,
		}

		vp.CalcularDisciplina()
		assert.Equal(t, 80.0, vp.Disciplina) // 8/10 = 80%
	})
}

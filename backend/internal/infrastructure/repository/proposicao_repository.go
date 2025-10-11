package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"

	"to-de-olho-backend/internal/domain"

	"github.com/jackc/pgx/v5/pgxpool"
)

type ProposicaoRepository struct {
	db DB
}

func NewProposicaoRepository(db *pgxpool.Pool) *ProposicaoRepository {
	return &ProposicaoRepository{db: db}
}

// ListProposicoes busca proposições no cache/banco aplicando filtros
func (r *ProposicaoRepository) ListProposicoes(ctx context.Context, filtros *domain.ProposicaoFilter) ([]domain.Proposicao, int, error) {
	logger := slog.Default()

	if r == nil || r.db == nil {
		logger.Warn("repository não inicializado")
		return nil, 0, nil
	}

	// Construir query dinâmica baseada nos filtros
	whereClause, args := r.buildWhereClause(filtros)
	baseQuery := "SELECT payload FROM proposicoes_cache"

	var query string
	if whereClause != "" {
		query = fmt.Sprintf("%s WHERE %s", baseQuery, whereClause)
	} else {
		query = baseQuery
	}

	// Adicionar ordenação e paginação
	if filtros != nil {
		offset := (filtros.Pagina - 1) * filtros.Limite
		query += fmt.Sprintf(" ORDER BY updated_at DESC LIMIT %d OFFSET %d", filtros.Limite, offset)
	} else {
		query += " ORDER BY updated_at DESC LIMIT 20"
	}

	logger.Info("executando query no repository",
		slog.String("query", query),
		slog.Any("args", args))

	// Executar query
	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		logger.Error("erro ao executar query",
			slog.String("error", err.Error()),
			slog.String("query", query))
		return nil, 0, err
	}
	defer rows.Close()

	var proposicoes []domain.Proposicao
	for rows.Next() {
		var payload string
		if err := rows.Scan(&payload); err != nil {
			logger.Error("erro ao fazer scan do resultado",
				slog.String("error", err.Error()))
			continue
		}

		var proposicao domain.Proposicao
		if err := json.Unmarshal([]byte(payload), &proposicao); err != nil {
			logger.Error("erro ao deserializar proposição",
				slog.String("error", err.Error()),
				slog.String("payload", payload))
			continue
		}

		proposicoes = append(proposicoes, proposicao)
	}

	if err := rows.Err(); err != nil {
		logger.Error("erro ao iterar resultados",
			slog.String("error", err.Error()))
		return nil, 0, err
	}

	// Contar total (simplificado - em produção seria uma query separada)
	total := len(proposicoes)

	logger.Info("proposições encontradas no repository",
		slog.Int("total", total))

	return proposicoes, total, nil
}

// GetProposicaoPorID busca uma proposição específica por ID
func (r *ProposicaoRepository) GetProposicaoPorID(ctx context.Context, id int) (*domain.Proposicao, error) {
	logger := slog.Default()

	if r == nil || r.db == nil {
		logger.Warn("repository não inicializado")
		return nil, nil
	}

	query := "SELECT payload FROM proposicoes_cache WHERE id = $1"

	logger.Info("buscando proposição por ID no repository",
		slog.Int("id", id))

	var payload string
	rows, err := r.db.Query(ctx, query, id)
	if err != nil {
		logger.Error("erro ao executar query para buscar proposição",
			slog.Int("id", id),
			slog.String("error", err.Error()))
		return nil, err
	}
	defer rows.Close()

	if !rows.Next() {
		logger.Info("proposição não encontrada no repository",
			slog.Int("id", id))
		return nil, domain.ErrProposicaoNaoEncontrada
	}

	err = rows.Scan(&payload)
	if err != nil {
		logger.Error("erro ao fazer scan da proposição",
			slog.Int("id", id),
			slog.String("error", err.Error()))
		return nil, err
	}

	var proposicao domain.Proposicao
	if err := json.Unmarshal([]byte(payload), &proposicao); err != nil {
		logger.Error("erro ao deserializar proposição",
			slog.Int("id", id),
			slog.String("error", err.Error()),
			slog.String("payload", payload))
		return nil, err
	}

	logger.Info("proposição encontrada no repository",
		slog.Int("id", id),
		slog.String("identificacao", proposicao.GetIdentificacao()))

	return &proposicao, nil
}

// UpsertProposicoes insere ou atualiza proposições no cache
func (r *ProposicaoRepository) UpsertProposicoes(ctx context.Context, proposicoes []domain.Proposicao) error {
	logger := slog.Default()

	if r == nil || r.db == nil || len(proposicoes) == 0 {
		return nil
	}

	logger.Info("iniciando upsert de proposições",
		slog.Int("quantidade", len(proposicoes)))

	for _, proposicao := range proposicoes {
		payload, err := json.Marshal(proposicao)
		if err != nil {
			logger.Error("erro ao serializar proposição para upsert",
				slog.Int("id", proposicao.ID),
				slog.String("error", err.Error()))
			continue
		}

		query := `INSERT INTO proposicoes_cache (id, payload, updated_at)
                  VALUES ($1, $2, NOW())
                  ON CONFLICT (id) DO UPDATE SET 
                  payload = EXCLUDED.payload, 
                  updated_at = NOW()`

		_, err = r.db.Exec(ctx, query, proposicao.ID, string(payload))
		if err != nil {
			logger.Error("erro ao fazer upsert da proposição",
				slog.Int("id", proposicao.ID),
				slog.String("error", err.Error()))
			return err
		}
	}

	logger.Info("upsert de proposições concluído com sucesso",
		slog.Int("quantidade", len(proposicoes)))

	return nil
}

// buildWhereClause constrói a cláusula WHERE dinamicamente baseada nos filtros
func (r *ProposicaoRepository) buildWhereClause(filtros *domain.ProposicaoFilter) (string, []interface{}) {
	if filtros == nil {
		return "", nil
	}

	var conditions []string
	var args []interface{}
	argCount := 0

	// Filtro por SiglaTipo
	if filtros.SiglaTipo != "" {
		argCount++
		conditions = append(conditions, fmt.Sprintf("payload::jsonb->>'siglaTipo' = $%d", argCount))
		args = append(args, filtros.SiglaTipo)
	}

	// Filtro por Numero
	if filtros.Numero != nil && *filtros.Numero > 0 {
		argCount++
		conditions = append(conditions, fmt.Sprintf("(payload::jsonb->>'numero')::int = $%d", argCount))
		args = append(args, *filtros.Numero)
	}

	// Filtro por Ano
	if filtros.Ano != nil && *filtros.Ano > 0 {
		argCount++
		conditions = append(conditions, fmt.Sprintf("(payload::jsonb->>'ano')::int = $%d", argCount))
		args = append(args, *filtros.Ano)
	}

	// Filtro por SiglaUfAutor
	if filtros.SiglaUfAutor != "" {
		argCount++
		conditions = append(conditions, fmt.Sprintf("payload::jsonb->>'siglaUfAutor' = $%d", argCount))
		args = append(args, filtros.SiglaUfAutor)
	}

	// Filtro por SiglaPartidoAutor
	if filtros.SiglaPartidoAutor != "" {
		argCount++
		conditions = append(conditions, fmt.Sprintf("payload::jsonb->>'siglaPartidoAutor' = $%d", argCount))
		args = append(args, filtros.SiglaPartidoAutor)
	}

	// Filtro por NomeAutor (busca parcial)
	if filtros.NomeAutor != "" {
		argCount++
		conditions = append(conditions, fmt.Sprintf("payload::jsonb->>'nomeAutor' ILIKE $%d", argCount))
		args = append(args, "%"+filtros.NomeAutor+"%")
	}

	// Filtro por Keywords na ementa (busca parcial)
	if filtros.Keywords != "" {
		argCount++
		conditions = append(conditions, fmt.Sprintf("payload::jsonb->>'ementa' ILIKE $%d", argCount))
		args = append(args, "%"+filtros.Keywords+"%")
	}

	// Filtro por CodSituacao
	if filtros.CodSituacao != nil && *filtros.CodSituacao > 0 {
		argCount++
		conditions = append(conditions, fmt.Sprintf("(payload::jsonb->'statusProposicao'->>'codSituacao')::int = $%d", argCount))
		args = append(args, *filtros.CodSituacao)
	}

	whereClause := strings.Join(conditions, " AND ")
	return whereClause, args
}

// GetProposicoesCountByDeputadoAno retorna contagem de proposições por deputado para um ano
func (r *ProposicaoRepository) GetProposicoesCountByDeputadoAno(ctx context.Context, ano int) ([]domain.ProposicaoCount, error) {
	logger := slog.Default()

	if r == nil || r.db == nil {
		logger.Warn("repository não inicializado")
		return nil, nil
	}

	// As proposições estão armazenadas em payload JSON na tabela proposicoes_cache
	// Extrair autor/id do payload e agrupar por idAutor (campo esperado em payload)
	// Tentar extrair o autor/deputado a partir de diferentes formatos de payload:
	// - payload->'autores' array (usar primeiro autor)
	// - payload->>'idAutor' campo simples
	// - payload->'ultimoRelator'->>'id' como fallback
	query := `SELECT COALESCE(
								(payload::jsonb->'autores'->0->>'id')::int,
								(payload::jsonb->>'idAutor')::int,
								(payload::jsonb->'ultimoRelator'->>'id')::int
							) as id_deputado,
							COUNT(*) as cnt
							FROM proposicoes_cache
							WHERE (payload::jsonb->>'ano')::int = $1
							GROUP BY id_deputado
							HAVING COALESCE(
								(payload::jsonb->'autores'->0->>'id')::int,
								(payload::jsonb->>'idAutor')::int,
								(payload::jsonb->'ultimoRelator'->>'id')::int
							) IS NOT NULL
							ORDER BY cnt DESC`

	rows, err := r.db.Query(ctx, query, ano)
	if err != nil {
		logger.Error("erro ao executar query agregada de proposições",
			slog.String("error", err.Error()))
		return nil, err
	}
	defer rows.Close()

	var results []domain.ProposicaoCount
	for rows.Next() {
		var idDeputado int
		var cnt int
		if err := rows.Scan(&idDeputado, &cnt); err != nil {
			logger.Error("erro ao scan agregacao proposicoes",
				slog.String("error", err.Error()))
			continue
		}
		results = append(results, domain.ProposicaoCount{IDDeputado: idDeputado, Count: cnt})
	}

	if err := rows.Err(); err != nil {
		logger.Error("erro ao iterar resultados agregacao proposicoes",
			slog.String("error", err.Error()))
		return nil, err
	}

	return results, nil
}

package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"to-de-olho-backend/internal/domain"

	"github.com/jackc/pgx/v5"
)

type VotacaoRepository struct {
	db DB
}

func NewVotacaoRepository(db DB) *VotacaoRepository {
	return &VotacaoRepository{db: db}
}

// CreateVotacao cria uma nova votação no banco de dados
func (r *VotacaoRepository) CreateVotacao(ctx context.Context, votacao *domain.Votacao) error {
	query := `
		INSERT INTO votacoes (
			id_votacao_camara, titulo, ementa, data_votacao, aprovacao,
			placar_sim, placar_nao, placar_abstencao, placar_outros,
			id_proposicao_principal, tipo_proposicao, numero_proposicao, 
			ano_proposicao, relevancia, payload
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
		RETURNING id, created_at, updated_at`

	payloadJSON, err := json.Marshal(votacao.Payload)
	if err != nil {
		return fmt.Errorf("erro ao serializar payload: %w", err)
	}

	err = r.db.QueryRow(ctx, query,
		votacao.IDVotacaoCamara,
		votacao.Titulo,
		votacao.Ementa,
		votacao.DataVotacao,
		votacao.Aprovacao,
		votacao.PlacarSim,
		votacao.PlacarNao,
		votacao.PlacarAbstencao,
		votacao.PlacarOutros,
		votacao.IDProposicaoPrincipal,
		votacao.TipoProposicao,
		votacao.NumeroProposicao,
		votacao.AnoProposicao,
		votacao.Relevancia,
		string(payloadJSON)).Scan(&votacao.ID, &votacao.CreatedAt, &votacao.UpdatedAt)

	if err != nil {
		return fmt.Errorf("erro ao criar votação: %w", err)
	}

	return nil
}

// GetVotacaoByID busca uma votação pelo ID
func (r *VotacaoRepository) GetVotacaoByID(ctx context.Context, id int64) (*domain.Votacao, error) {
	query := `
		SELECT id, id_votacao_camara, titulo, ementa, data_votacao, aprovacao,
			   placar_sim, placar_nao, placar_abstencao, placar_outros,
			   id_proposicao_principal, tipo_proposicao, numero_proposicao,
			   ano_proposicao, relevancia, payload, created_at, updated_at
		FROM votacoes 
		WHERE id = $1`

	votacao := &domain.Votacao{}
	var payloadStr string

	err := r.db.QueryRow(ctx, query, id).Scan(
		&votacao.ID,
		&votacao.IDVotacaoCamara,
		&votacao.Titulo,
		&votacao.Ementa,
		&votacao.DataVotacao,
		&votacao.Aprovacao,
		&votacao.PlacarSim,
		&votacao.PlacarNao,
		&votacao.PlacarAbstencao,
		&votacao.PlacarOutros,
		&votacao.IDProposicaoPrincipal,
		&votacao.TipoProposicao,
		&votacao.NumeroProposicao,
		&votacao.AnoProposicao,
		&votacao.Relevancia,
		&payloadStr,
		&votacao.CreatedAt,
		&votacao.UpdatedAt)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrVotacaoNaoEncontrada
		}
		return nil, fmt.Errorf("erro ao buscar votação: %w", err)
	}

	// Parse do payload JSON
	if err := json.Unmarshal([]byte(payloadStr), &votacao.Payload); err != nil {
		return nil, fmt.Errorf("erro ao deserializar payload: %w", err)
	}

	return votacao, nil
}

// ListVotacoes lista votações com filtros e paginação
func (r *VotacaoRepository) ListVotacoes(ctx context.Context, filtros domain.FiltrosVotacao, pag domain.Pagination) ([]*domain.Votacao, int, error) {
	// Construir WHERE clause dinamicamente
	whereConditions := []string{"1=1"}
	args := []interface{}{}
	argIndex := 1

	if filtros.Busca != "" {
		whereConditions = append(whereConditions, fmt.Sprintf("(titulo ILIKE $%d OR ementa ILIKE $%d)", argIndex, argIndex))
		args = append(args, "%"+filtros.Busca+"%")
		argIndex++
	}

	if filtros.Ano != 0 {
		whereConditions = append(whereConditions, fmt.Sprintf("EXTRACT(YEAR FROM data_votacao) = $%d", argIndex))
		args = append(args, filtros.Ano)
		argIndex++
	}

	if filtros.Aprovacao != "" {
		whereConditions = append(whereConditions, fmt.Sprintf("aprovacao = $%d", argIndex))
		args = append(args, filtros.Aprovacao)
		argIndex++
	}

	if filtros.Relevancia != "" {
		whereConditions = append(whereConditions, fmt.Sprintf("relevancia = $%d", argIndex))
		args = append(args, filtros.Relevancia)
		argIndex++
	}

	if filtros.TipoProposicao != "" {
		whereConditions = append(whereConditions, fmt.Sprintf("tipo_proposicao = $%d", argIndex))
		args = append(args, filtros.TipoProposicao)
		argIndex++
	}

	whereClause := strings.Join(whereConditions, " AND ")

	// Count total
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM votacoes WHERE %s", whereClause)
	var total int
	if err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("erro ao contar votações: %w", err)
	}

	// Query principal com paginação
	offset := (pag.Page - 1) * pag.Limit
	query := fmt.Sprintf(`
		SELECT id, id_votacao_camara, titulo, ementa, data_votacao, aprovacao,
			   placar_sim, placar_nao, placar_abstencao, placar_outros,
			   id_proposicao_principal, tipo_proposicao, numero_proposicao,
			   ano_proposicao, relevancia, payload, created_at, updated_at
		FROM votacoes 
		WHERE %s
		ORDER BY data_votacao DESC
		LIMIT $%d OFFSET $%d`, whereClause, argIndex, argIndex+1)

	args = append(args, pag.Limit, offset)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("erro ao listar votações: %w", err)
	}
	defer rows.Close()

	var votacoes []*domain.Votacao
	for rows.Next() {
		votacao := &domain.Votacao{}
		var payloadStr string

		err := rows.Scan(
			&votacao.ID,
			&votacao.IDVotacaoCamara,
			&votacao.Titulo,
			&votacao.Ementa,
			&votacao.DataVotacao,
			&votacao.Aprovacao,
			&votacao.PlacarSim,
			&votacao.PlacarNao,
			&votacao.PlacarAbstencao,
			&votacao.PlacarOutros,
			&votacao.IDProposicaoPrincipal,
			&votacao.TipoProposicao,
			&votacao.NumeroProposicao,
			&votacao.AnoProposicao,
			&votacao.Relevancia,
			&payloadStr,
			&votacao.CreatedAt,
			&votacao.UpdatedAt)

		if err != nil {
			return nil, 0, fmt.Errorf("erro ao escanear votação: %w", err)
		}

		// Parse do payload JSON
		if err := json.Unmarshal([]byte(payloadStr), &votacao.Payload); err != nil {
			return nil, 0, fmt.Errorf("erro ao deserializar payload: %w", err)
		}

		votacoes = append(votacoes, votacao)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("erro ao iterar votações: %w", err)
	}

	return votacoes, total, nil
}

// UpdateVotacao atualiza uma votação existente
func (r *VotacaoRepository) UpdateVotacao(ctx context.Context, votacao *domain.Votacao) error {
	query := `
		UPDATE votacoes SET
			titulo = $2,
			ementa = $3,
			data_votacao = $4,
			aprovacao = $5,
			placar_sim = $6,
			placar_nao = $7,
			placar_abstencao = $8,
			placar_outros = $9,
			id_proposicao_principal = $10,
			tipo_proposicao = $11,
			numero_proposicao = $12,
			ano_proposicao = $13,
			relevancia = $14,
			payload = $15,
			updated_at = NOW()
		WHERE id = $1`

	payloadJSON, err := json.Marshal(votacao.Payload)
	if err != nil {
		return fmt.Errorf("erro ao serializar payload: %w", err)
	}

	result, err := r.db.Exec(ctx, query,
		votacao.ID,
		votacao.Titulo,
		votacao.Ementa,
		votacao.DataVotacao,
		votacao.Aprovacao,
		votacao.PlacarSim,
		votacao.PlacarNao,
		votacao.PlacarAbstencao,
		votacao.PlacarOutros,
		votacao.IDProposicaoPrincipal,
		votacao.TipoProposicao,
		votacao.NumeroProposicao,
		votacao.AnoProposicao,
		votacao.Relevancia,
		string(payloadJSON))

	if err != nil {
		return fmt.Errorf("erro ao atualizar votação: %w", err)
	}

	if result.RowsAffected() == 0 {
		return domain.ErrVotacaoNaoEncontrada
	}

	return nil
}

// DeleteVotacao remove uma votação do banco de dados
func (r *VotacaoRepository) DeleteVotacao(ctx context.Context, id int64) error {
	query := `DELETE FROM votacoes WHERE id = $1`

	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("erro ao deletar votação: %w", err)
	}

	if result.RowsAffected() == 0 {
		return domain.ErrVotacaoNaoEncontrada
	}

	return nil
}

// UpsertVotacao insere ou atualiza uma votação
func (r *VotacaoRepository) UpsertVotacao(ctx context.Context, votacao *domain.Votacao) error {
	query := `
		INSERT INTO votacoes (
			id_votacao_camara, titulo, ementa, data_votacao, aprovacao,
			placar_sim, placar_nao, placar_abstencao, placar_outros,
			id_proposicao_principal, tipo_proposicao, numero_proposicao, 
			ano_proposicao, relevancia, payload
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
		ON CONFLICT (id_votacao_camara) 
		DO UPDATE SET
			titulo = EXCLUDED.titulo,
			ementa = EXCLUDED.ementa,
			data_votacao = EXCLUDED.data_votacao,
			aprovacao = EXCLUDED.aprovacao,
			placar_sim = EXCLUDED.placar_sim,
			placar_nao = EXCLUDED.placar_nao,
			placar_abstencao = EXCLUDED.placar_abstencao,
			placar_outros = EXCLUDED.placar_outros,
			id_proposicao_principal = EXCLUDED.id_proposicao_principal,
			tipo_proposicao = EXCLUDED.tipo_proposicao,
			numero_proposicao = EXCLUDED.numero_proposicao,
			ano_proposicao = EXCLUDED.ano_proposicao,
			relevancia = EXCLUDED.relevancia,
			payload = EXCLUDED.payload,
			updated_at = NOW()
		RETURNING id, created_at, updated_at`

	payloadJSON, err := json.Marshal(votacao.Payload)
	if err != nil {
		return fmt.Errorf("erro ao serializar payload: %w", err)
	}

	err = r.db.QueryRow(ctx, query,
		votacao.IDVotacaoCamara,
		votacao.Titulo,
		votacao.Ementa,
		votacao.DataVotacao,
		votacao.Aprovacao,
		votacao.PlacarSim,
		votacao.PlacarNao,
		votacao.PlacarAbstencao,
		votacao.PlacarOutros,
		votacao.IDProposicaoPrincipal,
		votacao.TipoProposicao,
		votacao.NumeroProposicao,
		votacao.AnoProposicao,
		votacao.Relevancia,
		string(payloadJSON)).Scan(&votacao.ID, &votacao.CreatedAt, &votacao.UpdatedAt)

	if err != nil {
		return fmt.Errorf("erro ao upsert votação: %w", err)
	}

	return nil
}

// Métodos para votos de deputados
func (r *VotacaoRepository) CreateVotoDeputado(ctx context.Context, voto *domain.VotoDeputado) error {
	query := `
		INSERT INTO votos_deputados (id_votacao, id_deputado, voto, justificativa, payload)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (id_votacao, id_deputado) DO NOTHING
		RETURNING id, created_at`

	var payloadStr string
	if voto.Payload != nil {
		payloadJSON, err := json.Marshal(voto.Payload)
		if err != nil {
			return fmt.Errorf("erro ao serializar payload do voto: %w", err)
		}
		payloadStr = string(payloadJSON)
	}

	err := r.db.QueryRow(ctx, query,
		voto.IDVotacao,
		voto.IDDeputado,
		voto.Voto,
		voto.Justificativa,
		payloadStr).Scan(&voto.ID, &voto.CreatedAt)

	if err != nil {
		if strings.Contains(err.Error(), "no rows in result set") {
			// Voto já existe, não é erro
			return nil
		}
		return fmt.Errorf("erro ao criar voto do deputado: %w", err)
	}

	return nil
}

func (r *VotacaoRepository) GetVotosPorVotacao(ctx context.Context, idVotacao int64) ([]*domain.VotoDeputado, error) {
	query := `
		SELECT id, id_votacao, id_deputado, voto, justificativa, payload, created_at
		FROM votos_deputados 
		WHERE id_votacao = $1
		ORDER BY id_deputado`

	rows, err := r.db.Query(ctx, query, idVotacao)
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar votos da votação: %w", err)
	}
	defer rows.Close()

	var votos []*domain.VotoDeputado
	for rows.Next() {
		voto := &domain.VotoDeputado{}
		var payloadStr *string

		err := rows.Scan(
			&voto.ID,
			&voto.IDVotacao,
			&voto.IDDeputado,
			&voto.Voto,
			&voto.Justificativa,
			&payloadStr,
			&voto.CreatedAt)

		if err != nil {
			return nil, fmt.Errorf("erro ao escanear voto: %w", err)
		}

		// Parse do payload JSON se existir
		if payloadStr != nil && *payloadStr != "" {
			if err := json.Unmarshal([]byte(*payloadStr), &voto.Payload); err != nil {
				return nil, fmt.Errorf("erro ao deserializar payload do voto: %w", err)
			}
		}

		votos = append(votos, voto)
	}

	return votos, nil
}

func (r *VotacaoRepository) GetVotoPorDeputado(ctx context.Context, idVotacao int64, idDeputado int) (*domain.VotoDeputado, error) {
	query := `
		SELECT id, id_votacao, id_deputado, voto, justificativa, payload, created_at
		FROM votos_deputados 
		WHERE id_votacao = $1 AND id_deputado = $2`

	voto := &domain.VotoDeputado{}
	var payloadStr *string

	err := r.db.QueryRow(ctx, query, idVotacao, idDeputado).Scan(
		&voto.ID,
		&voto.IDVotacao,
		&voto.IDDeputado,
		&voto.Voto,
		&voto.Justificativa,
		&payloadStr,
		&voto.CreatedAt)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrVotoDeputadoNaoEncontrado
		}
		return nil, fmt.Errorf("erro ao buscar voto do deputado: %w", err)
	}

	// Parse do payload JSON se existir
	if payloadStr != nil && *payloadStr != "" {
		if err := json.Unmarshal([]byte(*payloadStr), &voto.Payload); err != nil {
			return nil, fmt.Errorf("erro ao deserializar payload do voto: %w", err)
		}
	}

	return voto, nil
}

// Métodos para orientações partidárias
func (r *VotacaoRepository) CreateOrientacaoPartido(ctx context.Context, orientacao *domain.OrientacaoPartido) error {
	query := `
		INSERT INTO orientacoes_partidos (id_votacao, partido, orientacao)
		VALUES ($1, $2, $3)
		ON CONFLICT (id_votacao, partido) DO NOTHING
		RETURNING id, created_at`

	err := r.db.QueryRow(ctx, query,
		orientacao.IDVotacao,
		orientacao.Partido,
		orientacao.Orientacao).Scan(&orientacao.ID, &orientacao.CreatedAt)

	if err != nil {
		if strings.Contains(err.Error(), "no rows in result set") {
			// Orientação já existe, não é erro
			return nil
		}
		return fmt.Errorf("erro ao criar orientação do partido: %w", err)
	}

	return nil
}

func (r *VotacaoRepository) GetOrientacoesPorVotacao(ctx context.Context, idVotacao int64) ([]*domain.OrientacaoPartido, error) {
	query := `
		SELECT id, id_votacao, partido, orientacao, created_at
		FROM orientacoes_partidos 
		WHERE id_votacao = $1
		ORDER BY partido`

	rows, err := r.db.Query(ctx, query, idVotacao)
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar orientações da votação: %w", err)
	}
	defer rows.Close()

	var orientacoes []*domain.OrientacaoPartido
	for rows.Next() {
		orientacao := &domain.OrientacaoPartido{}

		err := rows.Scan(
			&orientacao.ID,
			&orientacao.IDVotacao,
			&orientacao.Partido,
			&orientacao.Orientacao,
			&orientacao.CreatedAt)

		if err != nil {
			return nil, fmt.Errorf("erro ao escanear orientação: %w", err)
		}

		orientacoes = append(orientacoes, orientacao)
	}

	return orientacoes, nil
}

// GetVotacaoDetalhada obtém votação completa com votos e orientações
func (r *VotacaoRepository) GetVotacaoDetalhada(ctx context.Context, id int64) (*domain.VotacaoDetalhada, error) {
	// Buscar a votação
	votacao, err := r.GetVotacaoByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Buscar votos dos deputados
	votos, err := r.GetVotosPorVotacao(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar votos da votação: %w", err)
	}

	// Buscar orientações partidárias
	orientacoes, err := r.GetOrientacoesPorVotacao(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar orientações da votação: %w", err)
	}

	return &domain.VotacaoDetalhada{
		Votacao:     *votacao,
		Votos:       votos,
		Orientacoes: orientacoes,
	}, nil
}

// GetPresencaPorDeputadoAno agrega a participação (votos registrados) por deputado em um ano
func (r *VotacaoRepository) GetPresencaPorDeputadoAno(ctx context.Context, ano int) ([]domain.PresencaCount, error) {
	// Contar votos registrados em votações do ano agrupados por id_deputado
	query := `SELECT v.id_deputado, COUNT(*) as participacoes
			  FROM votos_deputados v
			  JOIN votacoes vt ON vt.id = v.id_votacao
			  WHERE EXTRACT(YEAR FROM vt.data_votacao) = $1
			  GROUP BY v.id_deputado
			  ORDER BY participacoes DESC`

	rows, err := r.db.Query(ctx, query, ano)
	if err != nil {
		return nil, fmt.Errorf("erro ao agregar presenca: %w", err)
	}
	defer rows.Close()

	var results []domain.PresencaCount
	for rows.Next() {
		var idDeputado int
		var participacoes int
		if err := rows.Scan(&idDeputado, &participacoes); err != nil {
			return nil, fmt.Errorf("erro ao scan presenca: %w", err)
		}
		results = append(results, domain.PresencaCount{IDDeputado: idDeputado, Participacoes: participacoes})
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("erro ao iterar presenca: %w", err)
	}

	return results, nil
}

// GetRankingDeputadosAggregated agrega estatísticas de votos por deputado para o ano informado
func (r *VotacaoRepository) GetRankingDeputadosAggregated(ctx context.Context, ano int) ([]domain.RankingDeputadoVotacao, error) {
	query := `
		SELECT
			v.id_deputado,
			COUNT(*) AS total_votacoes,
			SUM(CASE WHEN v.voto = 'Sim' THEN 1 ELSE 0 END) AS votos_favoraveis,
			SUM(CASE WHEN v.voto = 'Não' THEN 1 ELSE 0 END) AS votos_contrarios,
			SUM(CASE WHEN v.voto NOT IN ('Sim','Não') THEN 1 ELSE 0 END) AS abstencoes
		FROM votos_deputados v
		JOIN votacoes vt ON vt.id = v.id_votacao
		WHERE EXTRACT(YEAR FROM vt.data_votacao) = $1
		GROUP BY v.id_deputado`

	rows, err := r.db.Query(ctx, query, ano)
	if err != nil {
		return nil, fmt.Errorf("erro ao agregar ranking de deputados: %w", err)
	}
	defer rows.Close()

	var results []domain.RankingDeputadoVotacao
	for rows.Next() {
		var row domain.RankingDeputadoVotacao
		if err := rows.Scan(&row.IDDeputado, &row.TotalVotacoes, &row.VotosFavoraveis, &row.VotosContrarios, &row.Abstencoes); err != nil {
			return nil, fmt.Errorf("erro ao escanear ranking de deputados: %w", err)
		}
		results = append(results, row)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("erro ao iterar ranking de deputados: %w", err)
	}

	return results, nil
}

// GetDisciplinaPartidosAggregated agrega votos e orientações por partido para cálculo de disciplina
func (r *VotacaoRepository) GetDisciplinaPartidosAggregated(ctx context.Context, ano int) ([]domain.VotacaoPartido, error) {
	query := `
	WITH votos_por_partido AS (
		SELECT 
			COALESCE(dc.payload->>'siglaPartido', 'SEM_PARTIDO') AS partido,
			SUM(CASE WHEN v.voto = 'Sim' THEN 1 ELSE 0 END) AS favor,
			SUM(CASE WHEN v.voto = 'Não' THEN 1 ELSE 0 END) AS contra,
			SUM(CASE WHEN v.voto NOT IN ('Sim','Não') THEN 1 ELSE 0 END) AS abst
		FROM votos_deputados v
		JOIN votacoes vt ON vt.id = v.id_votacao
		LEFT JOIN deputados_cache dc ON dc.id = v.id_deputado
		WHERE EXTRACT(YEAR FROM vt.data_votacao) = $1
		GROUP BY partido
	), orientacoes_recente AS (
		SELECT DISTINCT ON (op.partido)
			op.partido,
			op.orientacao
		FROM orientacoes_partidos op
		JOIN votacoes vt ON vt.id = op.id_votacao
		WHERE EXTRACT(YEAR FROM vt.data_votacao) = $1
		ORDER BY op.partido, vt.data_votacao DESC, op.created_at DESC
	), membros AS (
		SELECT 
			COALESCE(payload->>'siglaPartido', 'SEM_PARTIDO') AS partido,
			COUNT(*) AS total_membros
		FROM deputados_cache
		GROUP BY partido
	)
	SELECT 
		vp.partido,
		COALESCE(or_rec.orientacao, '') AS orientacao,
		vp.favor,
		vp.contra,
		vp.abst,
		COALESCE(m.total_membros, 0) AS total_membros
	FROM votos_por_partido vp
	LEFT JOIN orientacoes_recente or_rec ON or_rec.partido = vp.partido
	LEFT JOIN membros m ON m.partido = vp.partido`

	rows, err := r.db.Query(ctx, query, ano)
	if err != nil {
		return nil, fmt.Errorf("erro ao agregar disciplina partidária: %w", err)
	}
	defer rows.Close()

	var results []domain.VotacaoPartido
	for rows.Next() {
		var row domain.VotacaoPartido
		if err := rows.Scan(&row.Partido, &row.Orientacao, &row.VotaramFavor, &row.VotaramContra, &row.VotaramAbstencao, &row.TotalMembros); err != nil {
			return nil, fmt.Errorf("erro ao escanear disciplina partidária: %w", err)
		}
		results = append(results, row)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("erro ao iterar disciplina partidária: %w", err)
	}

	return results, nil
}

// GetVotacaoStatsAggregated retorna estatísticas consolidadas de votações para o ano
func (r *VotacaoRepository) GetVotacaoStatsAggregated(ctx context.Context, ano int) (*domain.VotacaoStats, error) {
	stats := &domain.VotacaoStats{
		VotacoesPorMes:        make([]int, 12),
		VotacoesPorRelevancia: map[string]int{},
	}

	if ano == 0 {
		return stats, nil
	}

	const totaisQuery = `
		SELECT 
			COUNT(*)::INT AS total,
			COALESCE(SUM(CASE WHEN aprovacao = 'Aprovada' THEN 1 ELSE 0 END), 0)::INT AS aprovadas
		FROM votacoes
		WHERE EXTRACT(YEAR FROM data_votacao) = $1`

	if err := r.db.QueryRow(ctx, totaisQuery, ano).Scan(&stats.TotalVotacoes, &stats.VotacoesAprovadas); err != nil {
		return nil, fmt.Errorf("erro ao obter totais de votações: %w", err)
	}
	stats.VotacoesRejeitadas = stats.TotalVotacoes - stats.VotacoesAprovadas

	mesesQuery := `
		SELECT EXTRACT(MONTH FROM data_votacao)::INT AS mes, COUNT(*)
		FROM votacoes
		WHERE EXTRACT(YEAR FROM data_votacao) = $1
		GROUP BY mes`
	mesRows, err := r.db.Query(ctx, mesesQuery, ano)
	if err != nil {
		return nil, fmt.Errorf("erro ao obter distribuição mensal de votações: %w", err)
	}
	for mesRows.Next() {
		var mes, count int
		if err := mesRows.Scan(&mes, &count); err != nil {
			mesRows.Close()
			return nil, fmt.Errorf("erro ao escanear distribuição mensal: %w", err)
		}
		if mes >= 1 && mes <= 12 {
			stats.VotacoesPorMes[mes-1] = count
		}
	}
	mesErr := mesRows.Err()
	mesRows.Close()
	if mesErr != nil {
		return nil, fmt.Errorf("erro ao iterar distribuição mensal: %w", mesErr)
	}

	relevanciaQuery := `
		SELECT relevancia, COUNT(*)
		FROM votacoes
		WHERE EXTRACT(YEAR FROM data_votacao) = $1
		GROUP BY relevancia`
	relRows, err := r.db.Query(ctx, relevanciaQuery, ano)
	if err != nil {
		return nil, fmt.Errorf("erro ao obter distribuição por relevância: %w", err)
	}
	for relRows.Next() {
		var relevancia string
		var count int
		if err := relRows.Scan(&relevancia, &count); err != nil {
			relRows.Close()
			return nil, fmt.Errorf("erro ao escanear distribuição por relevância: %w", err)
		}
		stats.VotacoesPorRelevancia[relevancia] = count
	}
	relErr := relRows.Err()
	relRows.Close()
	if relErr != nil {
		return nil, fmt.Errorf("erro ao iterar distribuição por relevância: %w", relErr)
	}

	const votosTotaisQuery = `
		SELECT COALESCE(COUNT(*), 0)
		FROM votos_deputados v
		JOIN votacoes vt ON vt.id = v.id_votacao
		WHERE EXTRACT(YEAR FROM vt.data_votacao) = $1`
	var votosTotais int
	if err := r.db.QueryRow(ctx, votosTotaisQuery, ano).Scan(&votosTotais); err != nil {
		return nil, fmt.Errorf("erro ao obter total de votos registrados: %w", err)
	}

	if stats.TotalVotacoes > 0 {
		stats.MediaParticipacao = float64(votosTotais) / float64(stats.TotalVotacoes)
	}

	return stats, nil
}

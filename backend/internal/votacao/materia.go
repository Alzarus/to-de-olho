package votacao

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"gorm.io/gorm"

	"github.com/Alzarus/to-de-olho/internal/utils"
)

// PreencherCodigoMateria completa codigo_materia nas votacoes gravadas antes
// de o sync passar a gravar o campo, no que da para derivar do banco: a
// identificacao exata ("PL 2338/2023") de uma proposicao ou de uma materia ja
// carregada, desde que ela aponte para um codigo so. Identificacoes com
// sufixo ("PLP 121/2024 (Substitutivo-CD)") sao outra materia e ficam para
// CompletarCodigoMateria. Idempotente: so toca linhas com o campo nulo.
// Roda no startup, depois do AutoMigrate (precisa das tabelas proposicoes e
// materias).
func PreencherCodigoMateria(db *gorm.DB) error {
	res := db.Exec(`
		UPDATE votacoes v
		SET codigo_materia = u.codigo
		FROM (
			SELECT identificacao, MIN(codigo) AS codigo
			FROM (
				SELECT descricao_identificacao AS identificacao, codigo_materia::int AS codigo
				FROM proposicoes WHERE codigo_materia ~ '^[0-9]{1,9}$'
				UNION
				SELECT identificacao, codigo_materia FROM materias
			) t
			GROUP BY identificacao
			HAVING COUNT(DISTINCT codigo) = 1
		) u
		WHERE v.codigo_materia IS NULL AND BTRIM(v.materia) = u.identificacao`)
	if res.Error != nil {
		return fmt.Errorf("preencher codigo_materia: %w", res.Error)
	}
	if res.RowsAffected > 0 {
		slog.Info("votacoes: codigo_materia preenchido pelo banco", "linhas", res.RowsAffected)
	}
	return nil
}

// ContarSemCodigoMateria conta as votacoes (nao os votos) do recorte ainda
// sem codigo_materia.
func (r *Repository) ContarSemCodigoMateria(desde time.Time) (int64, error) {
	var n int64
	err := r.db.Model(&Votacao{}).
		Where("codigo_materia IS NULL AND data >= ?", desde).
		Distinct("codigo_votacao").
		Count(&n).Error
	if err != nil {
		return 0, fmt.Errorf("contar votacoes sem codigo_materia: %w", err)
	}
	return n, nil
}

// CompletarCodigoMateria recarrega o recorte inteiro quando ainda ha votacoes
// sem codigo_materia (o historico gravado antes desta coluna). A recarga e a
// mesma do backfill: uma chamada por mes, upsert idempotente, e o
// DoUpdates grava codigo_materia e id_processo. Depois da primeira execucao
// completa nao ha mais pendencia e a funcao nao chama a API.
func (s *SyncService) CompletarCodigoMateria(ctx context.Context) error {
	inicio := utils.InicioRecorte()
	faltam, err := s.repo.ContarSemCodigoMateria(inicio)
	if err != nil {
		return err
	}
	if faltam == 0 {
		return nil
	}
	slog.Info("votacoes sem codigo_materia: recarregando o recorte", "votacoes", faltam)
	if _, err := s.SyncPeriodo(ctx, inicio, time.Now()); err != nil {
		return fmt.Errorf("recarga para codigo_materia: %w", err)
	}
	return nil
}

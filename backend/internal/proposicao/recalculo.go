package proposicao

import (
	"fmt"
	"log/slog"

	"gorm.io/gorm"
)

// RecalcularPontuacao reaplica CalcularPontuacao nas linhas gravadas. A
// pontuacao e guardada no sync; quando a regra muda (v2.2: pesos por sigla),
// as linhas antigas precisam ser refeitas. Idempotente: so grava o que mudou,
// e numa segunda execucao nao altera nada.
func RecalcularPontuacao(db *gorm.DB) (int, error) {
	alteradas := 0
	var lote []Proposicao
	err := db.Model(&Proposicao{}).Order("id").FindInBatches(&lote, 2000, func(tx *gorm.DB, _ int) error {
		for _, p := range lote {
			nova := p.CalcularPontuacao()
			if nova == p.Pontuacao {
				continue
			}
			if err := db.Model(&Proposicao{}).Where("id = ?", p.ID).UpdateColumn("pontuacao", nova).Error; err != nil {
				return fmt.Errorf("proposicao %d: %w", p.ID, err)
			}
			alteradas++
		}
		return nil
	}).Error
	if err != nil {
		return alteradas, fmt.Errorf("recalcular pontuacao: %w", err)
	}
	if alteradas > 0 {
		slog.Info("pontuacao das proposicoes recalculada", "alteradas", alteradas)
	}
	return alteradas, nil
}

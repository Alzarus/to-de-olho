package votacao

import (
	"fmt"
	"log/slog"

	"gorm.io/gorm"
)

// PreencherHistorico completa sigla_materia e secreta nas linhas gravadas
// antes de o sync passar a gravar esses campos. Roda depois do AutoMigrate.
//
// E idempotente: so toca linhas com o campo vazio/nulo, entao a segunda
// execucao nao altera nada e nunca sobrescreve o que veio da API.
//   - sigla_materia: primeira palavra de materia ("PLP 124/2022" -> "PLP"),
//     a mesma regra de siglaDaIdentificacao.
//   - secreta: true se algum voto da votacao e "Votou" (so aparece em
//     votacao secreta), senao false.
func PreencherHistorico(db *gorm.DB) error {
	res := db.Exec(`
		UPDATE votacoes
		SET sigla_materia = LEFT(UPPER(SPLIT_PART(BTRIM(materia), ' ', 1)), 20)
		WHERE COALESCE(sigla_materia, '') = ''
		  AND BTRIM(COALESCE(materia, '')) <> ''`)
	if res.Error != nil {
		return fmt.Errorf("preencher sigla_materia: %w", res.Error)
	}
	siglas := res.RowsAffected

	res = db.Exec(`
		UPDATE votacoes
		SET secreta = codigo_votacao IN (
			SELECT codigo_votacao FROM votacoes WHERE sigla_voto = 'Votou'
		)
		WHERE secreta IS NULL`)
	if res.Error != nil {
		return fmt.Errorf("preencher secreta: %w", res.Error)
	}
	if siglas > 0 || res.RowsAffected > 0 {
		slog.Info("votacoes: historico preenchido", "sigla_materia", siglas, "secreta", res.RowsAffected)
	}
	return nil
}

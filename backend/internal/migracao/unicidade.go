// Package migracao guarda ajustes de schema que o AutoMigrate não faz sozinho.
package migracao

import (
	"fmt"
	"strings"

	"gorm.io/gorm"
)

// unicidadeRedundante é uma constraint UNIQUE antiga que duplica um índice
// único criado depois pelo modelo (tag uniqueIndex) sobre a mesma coluna.
type unicidadeRedundante struct {
	tabela, constraint, indice string
}

// Constraints criadas pelo GORM antigo (tag `unique`) antes de os modelos
// passarem a declarar `uniqueIndex` com nome próprio. O banco de produção
// ficou com as duas coisas. O GORM 1.31 com o driver 1.6 passa a relatar a
// coluna como unique por causa da constraint e, como o modelo só declara o
// índice, tenta remover a constraint pelo nome que ele mesmo geraria
// (uni_<tabela>_<coluna>), que não existe: o AutoMigrate falha e a API não
// sobe. Foi o incidente de 25/09/2026 (Dependabot #50).
var unicidadesRedundantes = []unicidadeRedundante{
	{"despesas_ceaps", "despesas_ceaps_id_origem_key", "idx_despesa_origem"},
	{"materias", "idx_materias_id_processo", "idx_materia_id_processo"},
}

// removerSeCoberta apaga a constraint só se ela existir e se o índice único
// que a substitui já existir na mesma tabela: a coluna nunca fica sem
// unicidade, nem por um instante. Idempotente; tabela ausente é no-op.
const removerSeCoberta = `
DO $$
DECLARE
	tabela regclass := to_regclass(quote_ident(current_schema()) || '.' || quote_ident(@tabela));
BEGIN
	IF tabela IS NULL THEN
		RETURN;
	END IF;
	IF EXISTS (SELECT 1 FROM pg_constraint WHERE conrelid = tabela AND conname = @constraint AND contype = 'u')
	   AND EXISTS (
		SELECT 1 FROM pg_index i JOIN pg_class c ON c.oid = i.indexrelid
		WHERE i.indrelid = tabela AND c.relname = @indice AND i.indisunique
	   ) THEN
		EXECUTE format('ALTER TABLE %s DROP CONSTRAINT %I', tabela, @constraint);
	END IF;
END $$`

// substituir troca os marcadores @tabela, @constraint e @indice por literais
// SQL. Bloco DO não aceita parâmetros ligados; os valores são constantes deste
// arquivo, e o literal ainda dobra aspas simples por garantia.
func substituir(sql string, u unicidadeRedundante) string {
	return strings.NewReplacer(
		"@tabela", literal(u.tabela),
		"@constraint", literal(u.constraint),
		"@indice", literal(u.indice),
	).Replace(sql)
}

func literal(v string) string {
	return "'" + strings.ReplaceAll(v, "'", "''") + "'"
}

// RemoverUnicidadesRedundantes roda antes do AutoMigrate (cmd/api/main.go).
func RemoverUnicidadesRedundantes(db *gorm.DB) error {
	for _, u := range unicidadesRedundantes {
		sql := substituir(removerSeCoberta, u)
		if err := db.Exec(sql).Error; err != nil {
			return fmt.Errorf("remover constraint %s de %s: %w", u.constraint, u.tabela, err)
		}
	}
	return nil
}

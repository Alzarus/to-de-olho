package migracao

import (
	"testing"

	"github.com/Alzarus/to-de-olho/internal/testdb"
	"gorm.io/gorm"
)

// Reproduz o estado de produção de 25/09/2026: constraint antiga e índice
// único novo sobre a mesma coluna.
func criarLegado(t *testing.T, db *gorm.DB, comIndice bool) {
	t.Helper()
	sqls := []string{
		`CREATE TABLE despesas_ceaps (id bigserial PRIMARY KEY, id_origem bigint NOT NULL,
			CONSTRAINT despesas_ceaps_id_origem_key UNIQUE (id_origem))`,
		`CREATE TABLE materias (id bigserial PRIMARY KEY, id_processo bigint NOT NULL,
			CONSTRAINT idx_materias_id_processo UNIQUE (id_processo))`,
	}
	if comIndice {
		sqls = append(sqls,
			`CREATE UNIQUE INDEX idx_despesa_origem ON despesas_ceaps (id_origem)`,
			`CREATE UNIQUE INDEX idx_materia_id_processo ON materias (id_processo)`,
		)
	}
	for _, s := range sqls {
		if err := db.Exec(s).Error; err != nil {
			t.Fatal(err)
		}
	}
}

func constraintExiste(t *testing.T, db *gorm.DB, nome string) bool {
	t.Helper()
	var n int64
	err := db.Raw(`SELECT count(*) FROM pg_constraint c JOIN pg_namespace s ON s.oid = c.connamespace
		WHERE s.nspname = current_schema() AND c.conname = ?`, nome).Scan(&n).Error
	if err != nil {
		t.Fatal(err)
	}
	return n > 0
}

func TestRemoverUnicidadesRedundantesComIndice(t *testing.T) {
	db := testdb.Abrir(t)
	criarLegado(t, db, true)

	// Duas vezes: a segunda precisa ser no-op (roda a cada subida da API).
	for i := 0; i < 2; i++ {
		if err := RemoverUnicidadesRedundantes(db); err != nil {
			t.Fatalf("execução %d: %v", i+1, err)
		}
	}
	for _, u := range unicidadesRedundantes {
		if constraintExiste(t, db, u.constraint) {
			t.Errorf("constraint %s deveria ter sido removida", u.constraint)
		}
	}

	// A unicidade segue garantida pelo índice.
	if err := db.Exec(`INSERT INTO despesas_ceaps (id_origem) VALUES (7), (7)`).Error; err == nil {
		t.Error("id_origem duplicado foi aceito: a coluna ficou sem unicidade")
	}
}

func TestRemoverUnicidadesRedundantesSemIndiceMantemConstraint(t *testing.T) {
	db := testdb.Abrir(t)
	criarLegado(t, db, false)

	if err := RemoverUnicidadesRedundantes(db); err != nil {
		t.Fatal(err)
	}
	for _, u := range unicidadesRedundantes {
		if !constraintExiste(t, db, u.constraint) {
			t.Errorf("constraint %s removida sem o índice que a substitui", u.constraint)
		}
	}
}

func TestRemoverUnicidadesRedundantesBancoVazio(t *testing.T) {
	db := testdb.Abrir(t)
	if err := RemoverUnicidadesRedundantes(db); err != nil {
		t.Fatalf("banco sem as tabelas deveria ser no-op: %v", err)
	}
}

func TestLiteralDobraAspas(t *testing.T) {
	if got := literal("a'b"); got != "'a''b'" {
		t.Errorf("literal(a'b) = %s", got)
	}
}

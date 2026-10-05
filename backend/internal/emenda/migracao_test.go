package emenda_test

import (
	"testing"

	"github.com/Alzarus/to-de-olho/internal/emenda"
	"github.com/Alzarus/to-de-olho/internal/senador"
	"github.com/Alzarus/to-de-olho/internal/testdb"
)

// Linhas gravadas antes das colunas de datas continuam legiveis depois do
// AutoMigrate e entram como pendentes.
func TestMigracaoDatas_LinhasAntigas(t *testing.T) {
	db := testdb.Abrir(t, &senador.Senador{})
	if err := db.Create(&senador.Senador{ID: 1, CodigoParlamentar: 1, Nome: "S1"}).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Exec(`CREATE TABLE emendas (id bigserial PRIMARY KEY, senador_id bigint, ano bigint,
		numero text, tipo text, funcional_programatica text, localidade text,
		valor_empenhado numeric, valor_pago numeric, data_ultima_atualizacao timestamptz)`).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Exec(`INSERT INTO emendas (senador_id, ano, numero, valor_empenhado, valor_pago, data_ultima_atualizacao)
		VALUES (1, 2024, 'A', 10, 5, now())`).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&emenda.Emenda{}); err != nil {
		t.Fatal(err)
	}
	repo := emenda.NewRepository(db)
	if _, err := repo.ListBySenador(1, 0); err != nil {
		t.Fatalf("listar linha antiga: %v", err)
	}
	pend, err := repo.DatasPendentes(0)
	if err != nil || len(pend) != 1 {
		t.Fatalf("pendentes = %+v, %v", pend, err)
	}
}

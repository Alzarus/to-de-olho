package gabinete

import (
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Repository encapsula o acesso as tabelas de gabinete
type Repository struct {
	db *gorm.DB
}

// NewRepository cria um novo repository
func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// carimbo e o instante gravado em atualizado_em. O Postgres guarda
// microssegundos; truncar evita que a comparacao da limpeza falhe.
func carimbo(t time.Time) time.Time {
	return t.UTC().Truncate(time.Microsecond)
}

// SubstituirSenadorAno grava o retrato da fonte para (senador, ano): upsert
// idempotente na chave natural e, na mesma transacao, remove as linhas que a
// fonte deixou de trazer (ex.: um vinculo que zerou).
func (r *Repository) SubstituirSenadorAno(senadorID, ano int, recursos []Recurso, beneficios []Beneficio, agora time.Time) error {
	ts := carimbo(agora)
	for i := range recursos {
		recursos[i].ID = 0
		recursos[i].SenadorID, recursos[i].Ano, recursos[i].AtualizadoEm = senadorID, ano, ts
	}
	for i := range beneficios {
		beneficios[i].ID = 0
		beneficios[i].SenadorID, beneficios[i].Ano, beneficios[i].AtualizadoEm = senadorID, ano, ts
	}

	return r.db.Transaction(func(tx *gorm.DB) error {
		if len(recursos) > 0 {
			if err := tx.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "senador_id"}, {Name: "ano"}, {Name: "local"}, {Name: "vinculo"}},
				DoUpdates: clause.AssignmentColumns([]string{"quantidade", "atualizado_em"}),
			}).Create(&recursos).Error; err != nil {
				return fmt.Errorf("upsert de gabinete_recursos (%d/%d): %w", senadorID, ano, err)
			}
		}
		if err := tx.Where("senador_id = ? AND ano = ? AND atualizado_em <> ?", senadorID, ano, ts).
			Delete(&Recurso{}).Error; err != nil {
			return fmt.Errorf("limpeza de gabinete_recursos (%d/%d): %w", senadorID, ano, err)
		}

		if len(beneficios) > 0 {
			if err := tx.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "senador_id"}, {Name: "ano"}, {Name: "tipo"}},
				DoUpdates: clause.AssignmentColumns([]string{"utilizacao", "atualizado_em"}),
			}).Create(&beneficios).Error; err != nil {
				return fmt.Errorf("upsert de gabinete_beneficios (%d/%d): %w", senadorID, ano, err)
			}
		}
		if err := tx.Where("senador_id = ? AND ano = ? AND atualizado_em <> ?", senadorID, ano, ts).
			Delete(&Beneficio{}).Error; err != nil {
			return fmt.Errorf("limpeza de gabinete_beneficios (%d/%d): %w", senadorID, ano, err)
		}
		return nil
	})
}

// SubstituirMesa grava a composicao atual da Mesa. Lista vazia nao apaga nada
// (falha da fonte nao pode sumir com a nota da Presidencia).
func (r *Repository) SubstituirMesa(cargos []CargoMesa, agora time.Time) error {
	if len(cargos) == 0 {
		return errors.New("composicao da mesa vazia: nada gravado")
	}
	ts := carimbo(agora)
	for i := range cargos {
		cargos[i].AtualizadoEm = ts
	}
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "senador_id"}},
			DoUpdates: clause.AssignmentColumns([]string{"cargo", "atualizado_em"}),
		}).Create(&cargos).Error; err != nil {
			return fmt.Errorf("upsert de gabinete_mesa: %w", err)
		}
		if err := tx.Where("atualizado_em <> ?", ts).Delete(&CargoMesa{}).Error; err != nil {
			return fmt.Errorf("limpeza de gabinete_mesa: %w", err)
		}
		return nil
	})
}

// AnosDisponiveis lista os anos com dados de pessoal do senador, do mais recente ao mais antigo
func (r *Repository) AnosDisponiveis(senadorID int) ([]int, error) {
	var anos []int
	err := r.db.Model(&Recurso{}).Where("senador_id = ?", senadorID).
		Distinct("ano").Order("ano DESC").Pluck("ano", &anos).Error
	return anos, err
}

// Recursos devolve as linhas de pessoal de (senador, ano)
func (r *Repository) Recursos(senadorID, ano int) ([]Recurso, error) {
	var out []Recurso
	err := r.db.Where("senador_id = ? AND ano = ?", senadorID, ano).
		Order("local ASC, quantidade DESC, vinculo ASC").Find(&out).Error
	return out, err
}

// Beneficios devolve os beneficios de (senador, ano)
func (r *Repository) Beneficios(senadorID, ano int) ([]Beneficio, error) {
	var out []Beneficio
	err := r.db.Where("senador_id = ? AND ano = ?", senadorID, ano).Order("tipo ASC").Find(&out).Error
	return out, err
}

// CargoMesaAtual devolve o cargo do senador na Mesa ou nil
func (r *Repository) CargoMesaAtual(senadorID int) (*CargoMesa, error) {
	var c CargoMesa
	err := r.db.Where("senador_id = ?", senadorID).Limit(1).Find(&c).Error
	if err != nil {
		return nil, err
	}
	if c.SenadorID == 0 {
		return nil, nil
	}
	return &c, nil
}

// ContarAno conta as linhas de pessoal gravadas num ano (guarda do scheduler)
func (r *Repository) ContarAno(ano int) (int64, error) {
	var n int64
	err := r.db.Model(&Recurso{}).Where("ano = ?", ano).Count(&n).Error
	return n, err
}

// UltimaAtualizacao e o atualizado_em mais recente de um ano (guarda do scheduler)
func (r *Repository) UltimaAtualizacao(ano int) (*time.Time, error) {
	var t *time.Time
	err := r.db.Model(&Recurso{}).Where("ano = ?", ano).Select("MAX(atualizado_em)").Scan(&t).Error
	return t, err
}

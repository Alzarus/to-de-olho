package senador

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"gorm.io/gorm"

	"github.com/Alzarus/to-de-olho/pkg/retry"
	"github.com/Alzarus/to-de-olho/pkg/senado"
)

type senadoExercicio = senado.Exercicio

// Periodos de exercicio (item 8 da auditoria). Ficam na tabela mandatos, uma
// linha por exercicio: Inicio e Fim (inclusivo; nulo = em exercicio). O teto
// da cota e o piso de tempo do ranking sao calculados a partir deles.

// diasPorMes converte dias em meses (media do calendario gregoriano)
const diasPorMes = 365.2425 / 12

// SyncExercicios grava os periodos de exercicio dos senadores em exercicio.
// Cada senador tem ate 3 tentativas; falhas sao devolvidas juntas no fim.
func (s *SyncService) SyncExercicios(ctx context.Context) error {
	senadores, err := s.repo.FindAll(false)
	if err != nil {
		return err
	}
	var total int
	var falhas []string
	for _, sen := range senadores {
		var mandatos []Mandato
		err := retry.WithRetry(ctx, 3, "exercicios "+sen.Nome, func() error {
			exercicios, err := s.client.ListarExercicios(ctx, sen.CodigoParlamentar)
			if err != nil {
				return err
			}
			mandatos, err = converterExercicios(sen.ID, exercicios)
			return err
		})
		if err == nil && len(mandatos) == 0 {
			err = fmt.Errorf("API sem periodo de exercicio")
		}
		if err == nil {
			err = s.repo.SubstituirMandatos(sen.ID, mandatos)
		}
		if err != nil {
			slog.Error("exercicios do senador nao sincronizados", "senador", sen.Nome, "error", err)
			falhas = append(falhas, sen.Nome)
			continue
		}
		total += len(mandatos)
	}
	slog.Info("sync de exercicios concluido", "senadores", len(senadores)-len(falhas), "falhas", len(falhas), "periodos", total)
	if len(falhas) > 0 {
		return fmt.Errorf("exercicios de %d senadores nao sincronizados: %s", len(falhas), strings.Join(falhas, ", "))
	}
	return nil
}

func converterExercicios(senadorID int, exercicios []senadoExercicio) ([]Mandato, error) {
	var out []Mandato
	for _, e := range exercicios {
		inicio, err := time.Parse("2006-01-02", e.Inicio)
		if err != nil {
			return nil, fmt.Errorf("data de inicio invalida %q: %w", e.Inicio, err)
		}
		m := Mandato{SenadorID: senadorID, Legislatura: e.Legislatura, Inicio: inicio, Tipo: e.Participacao}
		if e.Fim != "" {
			fim, err := time.Parse("2006-01-02", e.Fim)
			if err != nil {
				return nil, fmt.Errorf("data de fim invalida %q: %w", e.Fim, err)
			}
			m.Fim = &fim
		}
		out = append(out, m)
	}
	return out, nil
}

// SubstituirMandatos troca os periodos de exercicio do senador numa transacao
func (r *Repository) SubstituirMandatos(senadorID int, mandatos []Mandato) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("senador_id = ?", senadorID).Delete(&Mandato{}).Error; err != nil {
			return err
		}
		if len(mandatos) == 0 {
			return nil
		}
		return tx.Create(&mandatos).Error
	})
}

// MesesEmExercicio soma, em meses, os dias em que o senador ocupou a cadeira
// dentro de [inicio, fim).
func (r *Repository) MesesEmExercicio(senadorID int, inicio, fim time.Time) (float64, error) {
	var mandatos []Mandato
	if err := r.db.Where("senador_id = ?", senadorID).Find(&mandatos).Error; err != nil {
		return 0, err
	}
	return mesesEmExercicio(mandatos, inicio, fim), nil
}

func mesesEmExercicio(mandatos []Mandato, inicio, fim time.Time) float64 {
	var dias float64
	for _, m := range mandatos {
		a := m.Inicio
		b := fim
		if m.Fim != nil {
			if f := m.Fim.AddDate(0, 0, 1); f.Before(b) { // Fim e inclusivo
				b = f
			}
		}
		if a.Before(inicio) {
			a = inicio
		}
		if b.After(a) {
			dias += b.Sub(a).Hours() / 24
		}
	}
	return dias / diasPorMes
}

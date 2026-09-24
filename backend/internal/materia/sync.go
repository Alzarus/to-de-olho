package materia

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/Alzarus/to-de-olho/internal/utils"
	"github.com/Alzarus/to-de-olho/pkg/retry"
	senadoapi "github.com/Alzarus/to-de-olho/pkg/senado"
)

// ClienteLegis e o que o sync usa da API do Senado (*senado.LegisClient;
// interface para testes)
type ClienteLegis interface {
	ObterProcesso(ctx context.Context, idProcesso int) (*senadoapi.ProcessoDetalhe, error)
	ObterIDProcesso(ctx context.Context, codigoMateria int) (int, error)
}

const (
	// emParalelo: com 4 em paralelo a API devolveu 429 em ~1% das chamadas
	// (commit fc5f134). Aqui nao ha pressa: 2 workers e uma pausa entre
	// chamadas de cada um.
	emParalelo = 2
	pausa      = 300 * time.Millisecond
	// idadeMaxima: o detalhe e rebuscado depois disso (o Senado pode atribuir
	// um apelido ou uma explicacao da ementa depois da votacao)
	idadeMaxima = 30 * 24 * time.Hour
	// PorRodadaDiaria limita as chamadas do sync diario; o backfill nao limita
	PorRodadaDiaria = 800
)

// siglasProposicao sao os tipos de proposicao cujo detalhe vale buscar:
// requerimentos, indicacoes e oficios nao tem apelido nem explicacao da
// ementa, e sao a maioria das linhas.
var siglasProposicao = []string{"PEC", "PLP", "PL", "PLS", "PDL", "PRS"}

// SyncService carrega a tabela materias a partir de /processo/{id}.
type SyncService struct {
	db     *gorm.DB
	client ClienteLegis

	paralelo int
	pausa    time.Duration
	idade    time.Duration
}

// NewSyncService cria o servico de sync das materias
func NewSyncService(db *gorm.DB, client ClienteLegis) *SyncService {
	return &SyncService{db: db, client: client, paralelo: emParalelo, pausa: pausa, idade: idadeMaxima}
}

// ComPausa troca a pausa entre chamadas de cada worker (testes usam 0).
func (s *SyncService) ComPausa(d time.Duration) *SyncService {
	s.pausa = d
	return s
}

// Pendente e uma materia sem detalhe (ou com detalhe antigo).
type Pendente struct {
	CodigoMateria int
	IdProcesso    *int // nulo em linhas antigas: resolvido por /processo?codigoMateria=
}

// Resumo descreve uma rodada do sync.
type Resumo struct {
	Pendentes      int // materias a buscar nesta rodada
	Gravadas       int
	ComApelido     int // com apelido oficial
	ComExplicacao  int
	NaoEncontradas int // 404 ou codigo sem processo
	Falhas         int
}

// Pendentes lista as materias referenciadas por votacoes (exceto MSF/OFS, cuja
// descricao da votacao ja nomeia a autoridade) e por proposicoes do recorte
// (so os tipos de siglasProposicao) sem linha em materias ou com linha mais
// antiga que a idade maxima. Votacoes vem primeiro. limite <= 0: todas.
func (s *SyncService) Pendentes(limite int) ([]Pendente, error) {
	sqlStr := `
		WITH codigos AS (
			SELECT codigo_materia, MAX(id_processo) AS id_processo, 0 AS prioridade, MAX(data) AS data
			FROM votacoes
			WHERE codigo_materia IS NOT NULL AND COALESCE(sigla_materia, '') NOT IN ('MSF', 'OFS')
			GROUP BY codigo_materia
			UNION ALL
			SELECT ` + CodigoTextoParaInt("codigo_materia") + `, MAX(id_processo), 1, MAX(data_apresentacao)
			FROM proposicoes
			WHERE sigla_subtipo_materia IN ? AND data_apresentacao >= ?
			GROUP BY codigo_materia
		)
		SELECT c.codigo_materia, MAX(c.id_processo) AS id_processo
		FROM codigos c
		LEFT JOIN materias m ON m.codigo_materia = c.codigo_materia
		WHERE c.codigo_materia IS NOT NULL AND (m.codigo_materia IS NULL OR m.updated_at < ?)
		GROUP BY c.codigo_materia
		ORDER BY MIN(c.prioridade), MAX(c.data) DESC NULLS LAST, c.codigo_materia DESC`
	args := []any{siglasProposicao, utils.InicioRecorte(), time.Now().Add(-s.idade)}
	if limite > 0 {
		sqlStr += " LIMIT ?"
		args = append(args, limite)
	}
	var out []Pendente
	if err := s.db.Raw(sqlStr, args...).Scan(&out).Error; err != nil {
		return nil, fmt.Errorf("listar materias pendentes: %w", err)
	}
	return out, nil
}

// Sync busca o detalhe das materias pendentes (ate `limite`; <= 0: todas) e
// grava em materias. Falha numa materia nao interrompe as outras; o erro so
// volta se nenhuma pendente foi gravada.
func (s *SyncService) Sync(ctx context.Context, limite int) (Resumo, error) {
	var resumo Resumo
	pendentes, err := s.Pendentes(limite)
	if err != nil {
		return resumo, err
	}
	resumo.Pendentes = len(pendentes)
	if len(pendentes) == 0 {
		slog.Info("materias: nada pendente")
		return resumo, nil
	}
	slog.Info("materias: buscando detalhes", "pendentes", len(pendentes))

	var mu sync.Mutex
	fila := make(chan Pendente)
	var wg sync.WaitGroup
	for w := 0; w < max(1, s.paralelo); w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for p := range fila {
				m, err := s.buscar(ctx, p)
				if err == nil && m != nil {
					err = s.Upsert(m)
				}
				mu.Lock()
				switch {
				case errors.Is(err, senadoapi.ErrNaoEncontrado) || (err == nil && m == nil):
					resumo.NaoEncontradas++
					slog.Warn("materia sem processo na API", "codigo_materia", p.CodigoMateria)
				case err != nil:
					resumo.Falhas++
					slog.Error("materia nao sincronizada", "codigo_materia", p.CodigoMateria, "error", err)
				default:
					resumo.Gravadas++
					if m.Apelido != nil {
						resumo.ComApelido++
					}
					if m.ExplicacaoEmenta != nil {
						resumo.ComExplicacao++
					}
				}
				mu.Unlock()
				select {
				case <-ctx.Done():
				case <-time.After(s.pausa):
				}
			}
		}()
	}
	for _, p := range pendentes {
		if ctx.Err() != nil {
			break
		}
		fila <- p
	}
	close(fila)
	wg.Wait()

	slog.Info("materias sincronizadas", "pendentes", resumo.Pendentes, "gravadas", resumo.Gravadas,
		"com_apelido", resumo.ComApelido, "com_explicacao", resumo.ComExplicacao,
		"nao_encontradas", resumo.NaoEncontradas, "falhas", resumo.Falhas)
	if err := ctx.Err(); err != nil {
		return resumo, fmt.Errorf("sync de materias interrompido: %w", err)
	}
	if resumo.Gravadas == 0 && resumo.Falhas > 0 {
		return resumo, fmt.Errorf("nenhuma das %d materias pendentes foi gravada (%d falhas)", resumo.Pendentes, resumo.Falhas)
	}
	return resumo, nil
}

// buscar resolve o id do processo (se preciso) e baixa o detalhe, com retry.
// 404 nao e repetido. Devolve nil, nil quando o codigo nao tem processo.
func (s *SyncService) buscar(ctx context.Context, p Pendente) (*Materia, error) {
	id := 0
	if p.IdProcesso != nil {
		id = *p.IdProcesso
	}
	if id == 0 {
		err := comRetry(ctx, fmt.Sprintf("id do processo da materia %d", p.CodigoMateria), func() error {
			var err error
			id, err = s.client.ObterIDProcesso(ctx, p.CodigoMateria)
			return err
		})
		if err != nil {
			return nil, err
		}
		if id == 0 {
			return nil, nil
		}
	}
	var det *senadoapi.ProcessoDetalhe
	err := comRetry(ctx, fmt.Sprintf("processo %d", id), func() error {
		var err error
		det, err = s.client.ObterProcesso(ctx, id)
		return err
	})
	if err != nil {
		return nil, err
	}
	if det == nil {
		return nil, nil
	}
	m := DoProcesso(det)
	if m.IdProcesso == 0 {
		m.IdProcesso = id
	}
	if m.CodigoMateria != p.CodigoMateria {
		slog.Warn("codigoMateria do processo difere do pedido; mantido o pedido",
			"pedido", p.CodigoMateria, "processo", m.CodigoMateria, "id_processo", id)
		m.CodigoMateria = p.CodigoMateria
	}
	return m, nil
}

// comRetry: 5 tentativas (backoff ate 8s), como no detalhe de proposicoes;
// 404 encerra na hora.
func comRetry(ctx context.Context, operacao string, fn func() error) error {
	var naoEncontrado error
	err := retry.WithRetry(ctx, 5, operacao, func() error {
		err := fn()
		if errors.Is(err, senadoapi.ErrNaoEncontrado) {
			naoEncontrado = err
			return nil
		}
		return err
	})
	if naoEncontrado != nil {
		return naoEncontrado
	}
	return err
}

// DoProcesso converte o detalhe de /processo/{id} numa linha de materias.
func DoProcesso(d *senadoapi.ProcessoDetalhe) *Materia {
	var temas Temas
	vistos := map[string]bool{}
	for _, c := range d.Classificacoes {
		t := strings.TrimSpace(c.Descricao)
		if t == "" || vistos[t] {
			continue
		}
		vistos[t] = true
		temas = append(temas, t)
	}
	sigla := strings.ToUpper(strings.TrimSpace(d.Sigla))
	if sigla == "" {
		if campos := strings.Fields(d.Identificacao); len(campos) > 0 {
			sigla = strings.ToUpper(campos[0])
		}
	}
	if len(sigla) > 20 {
		sigla = sigla[:20]
	}
	return &Materia{
		CodigoMateria:    d.CodigoMateria,
		IdProcesso:       d.ID,
		Identificacao:    strings.TrimSpace(d.Identificacao),
		Sigla:            sigla,
		Apelido:          NormalizarApelido(d.Apelido),
		Ementa:           strings.TrimSpace(d.Conteudo.Ementa),
		ExplicacaoEmenta: textoOuNulo(d.Conteudo.ExplicacaoEmenta),
		Temas:            temas,
		UrlDocumento:     strings.TrimSpace(d.Documento.URL),
	}
}

// Upsert grava a materia pela chave codigo_materia (idempotente).
func (s *SyncService) Upsert(m *Materia) error {
	m.UpdatedAt = time.Now()
	err := s.db.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "codigo_materia"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"id_processo", "identificacao", "sigla", "apelido", "ementa",
			"explicacao_ementa", "temas", "url_documento", "updated_at",
		}),
	}).Create(m).Error
	if err != nil {
		return fmt.Errorf("gravar materia %d: %w", m.CodigoMateria, err)
	}
	return nil
}

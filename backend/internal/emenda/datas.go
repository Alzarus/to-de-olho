package emenda

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/Alzarus/to-de-olho/pkg/transparencia"
)

const (
	// DatasPorRodada limita as emendas consultadas no sync diario (~2,5
	// paginas de documentos cada). O backfill pelo endpoint nao limita.
	DatasPorRodada = 600
	// pausaDocumentos fica abaixo do limite do Portal (90 req/min das 6h as 24h)
	pausaDocumentos = 700 * time.Millisecond
	// maxPaginasDocumentos: a maior emenda vista tinha 4 paginas (15 por pagina)
	maxPaginasDocumentos = 50
)

// Datas resume os documentos de execucao de uma emenda.
type Datas struct {
	PrimeiroEmpenho   *time.Time
	PrimeiroPagamento *time.Time
	UltimoPagamento   *time.Time
}

// ResumirDatas pega a data do primeiro empenho e a do primeiro e do ultimo
// pagamento. Liquidacao fica de fora: e etapa intermediaria entre as duas.
// Data em formato invalido e ignorada.
func ResumirDatas(docs []transparencia.DocumentoEmendaDTO) Datas {
	var d Datas
	for _, doc := range docs {
		t, err := time.Parse("02/01/2006", strings.TrimSpace(doc.Data))
		if err != nil {
			continue
		}
		switch strings.ToLower(strings.TrimSpace(doc.Fase)) {
		case "empenho":
			d.PrimeiroEmpenho = menor(d.PrimeiroEmpenho, t)
		case "pagamento":
			d.PrimeiroPagamento = menor(d.PrimeiroPagamento, t)
			d.UltimoPagamento = maior(d.UltimoPagamento, t)
		}
	}
	return d
}

func menor(atual *time.Time, t time.Time) *time.Time {
	if atual == nil || t.Before(*atual) {
		return &t
	}
	return atual
}

func maior(atual *time.Time, t time.Time) *time.Time {
	if atual == nil || t.After(*atual) {
		return &t
	}
	return atual
}

// emendaPendente e um codigo de emenda cujas datas faltam ou estao velhas.
type emendaPendente struct {
	Numero         string
	ValorEmpenhado float64
}

// ResumoDatas descreve uma rodada do sync de datas.
type ResumoDatas struct {
	Pendentes int `json:"pendentes"`
	Gravadas  int `json:"gravadas"`
	SemDados  int `json:"sem_dados"` // a fonte nao devolveu documento; fica pendente
	Falhas    int `json:"falhas"`
}

// SyncDatas busca as datas das emendas nunca consultadas ou cujo valor pago
// mudou desde a ultima consulta (ate `limite`; <= 0: todas), das mais
// recentes para as mais antigas. Falha numa emenda nao interrompe as outras.
func (s *SyncService) SyncDatas(ctx context.Context, limite int) (ResumoDatas, error) {
	var resumo ResumoDatas
	pendentes, err := s.repo.DatasPendentes(limite)
	if err != nil {
		return resumo, err
	}
	resumo.Pendentes = len(pendentes)
	if len(pendentes) == 0 {
		slog.Info("datas das emendas: nada pendente")
		return resumo, nil
	}
	slog.Info("datas das emendas: buscando", "pendentes", len(pendentes))

	for _, p := range pendentes {
		if err := ctx.Err(); err != nil {
			return resumo, err
		}
		docs, err := s.documentos(ctx, p.Numero)
		if err != nil {
			slog.Warn("falha ao buscar documentos da emenda", "numero", p.Numero, "erro", err)
			resumo.Falhas++
			continue
		}
		// Lista vazia com valor empenhado: a fonte as vezes devolve vazio por
		// instantes; marcar como consultada esconderia as datas ate o valor mudar
		if len(docs) == 0 && p.ValorEmpenhado > 0 {
			resumo.SemDados++
			continue
		}
		if err := s.repo.GravarDatas(p.Numero, ResumirDatas(docs), time.Now()); err != nil {
			slog.Warn("falha ao gravar datas da emenda", "numero", p.Numero, "erro", err)
			resumo.Falhas++
			continue
		}
		resumo.Gravadas++
	}

	slog.Info("datas das emendas: concluido",
		"pendentes", resumo.Pendentes,
		"gravadas", resumo.Gravadas,
		"sem_dados", resumo.SemDados,
		"falhas", resumo.Falhas,
	)
	if resumo.Gravadas == 0 && resumo.Falhas > 0 {
		return resumo, fmt.Errorf("nenhuma das %d emendas teve as datas gravadas", resumo.Pendentes)
	}
	return resumo, nil
}

// documentos le todas as paginas de documentos da emenda.
func (s *SyncService) documentos(ctx context.Context, numero string) ([]transparencia.DocumentoEmendaDTO, error) {
	var todos []transparencia.DocumentoEmendaDTO
	for pagina := 1; pagina <= maxPaginasDocumentos; pagina++ {
		docs, err := s.documentosAPI.GetDocumentosEmenda(ctx, numero, pagina)
		if err != nil {
			return nil, err
		}
		time.Sleep(s.pausaDocumentos)
		if len(docs) == 0 {
			break
		}
		todos = append(todos, docs...)
	}
	return todos, nil
}

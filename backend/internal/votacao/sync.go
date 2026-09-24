package votacao

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/Alzarus/to-de-olho/internal/senador"
	"github.com/Alzarus/to-de-olho/internal/utils"
	"github.com/Alzarus/to-de-olho/pkg/retry"
	senadoapi "github.com/Alzarus/to-de-olho/pkg/senado"
)

// clienteLegis e o que o sync usa da API do Senado (interface para testes)
type clienteLegis interface {
	ListarVotacoesPeriodo(ctx context.Context, inicio, fim time.Time) ([]senadoapi.VotacaoSessaoAPI, error)
}

// SyncService gerencia sincronizacao de votacoes
type SyncService struct {
	repo        *Repository
	senadorRepo *senador.Repository
	client      clienteLegis
}

// NewSyncService cria um novo servico de sincronizacao
func NewSyncService(repo *Repository, senadorRepo *senador.Repository, client *senadoapi.LegisClient) *SyncService {
	return &SyncService{
		repo:        repo,
		senadorRepo: senadorRepo,
		client:      client,
	}
}

// ResumoCarga descreve o que uma carga de votacoes gravou
type ResumoCarga struct {
	Votacoes  int // votacoes nominais recebidas da API
	Votos     int // linhas gravadas (um voto por senador conhecido)
	Ignorados int // votos de parlamentares fora da tabela senadores
}

// SyncFromAPI carrega os votos de todas as votacoes do recorte (posse da
// legislatura atual ate hoje).
func (s *SyncService) SyncFromAPI(ctx context.Context) error {
	_, err := s.SyncPeriodo(ctx, utils.InicioRecorte(), time.Now())
	return err
}

// SyncRecentes grava as votacoes dos ultimos `dias` dias (sync diario).
func (s *SyncService) SyncRecentes(ctx context.Context, dias int) (ResumoCarga, error) {
	fim := time.Now()
	return s.SyncPeriodo(ctx, fim.AddDate(0, 0, -dias), fim)
}

// SyncPeriodo grava os votos das votacoes nominais com sessao em [inicio, fim].
//
// Fonte: /votacao?dataInicio=&dataFim=, que traz todas as cadeiras de cada
// votacao. Substitui as 81 chamadas por senador (que falhavam por corte de
// resposta e deixavam senadores inteiros de fora, item 4). O periodo e
// quebrado por mes para manter as respostas pequenas; cada mes tem 3
// tentativas e qualquer falha interrompe a carga com erro.
func (s *SyncService) SyncPeriodo(ctx context.Context, inicio, fim time.Time) (ResumoCarga, error) {
	var resumo ResumoCarga

	senadorPorCodigo, err := s.mapaSenadores()
	if err != nil {
		return resumo, err
	}

	for _, janela := range janelasMensais(inicio, fim) {
		var votacoesAPI []senadoapi.VotacaoSessaoAPI
		err := retry.WithRetry(ctx, 3, "votacoes "+janela[0].Format("2006-01"), func() error {
			var err error
			votacoesAPI, err = s.client.ListarVotacoesPeriodo(ctx, janela[0], janela[1])
			return err
		})
		if err != nil {
			return resumo, err
		}

		var votos []Votacao
		for _, v := range votacoesAPI {
			convertidos, ignorados, err := converterVotacao(v, senadorPorCodigo)
			if err != nil {
				return resumo, err
			}
			votos = append(votos, convertidos...)
			resumo.Ignorados += ignorados
		}
		if err := s.repo.UpsertBatch(votos); err != nil {
			return resumo, fmt.Errorf("falha ao gravar votos de %s: %w", janela[0].Format("2006-01"), err)
		}
		resumo.Votacoes += len(votacoesAPI)
		resumo.Votos += len(votos)
	}

	slog.Info("votacoes sincronizadas",
		"inicio", inicio.Format("2006-01-02"), "fim", fim.Format("2006-01-02"),
		"votacoes", resumo.Votacoes, "votos", resumo.Votos, "ignorados", resumo.Ignorados)
	return resumo, nil
}

// SenadoresSemVotos devolve os senadores em exercicio sem nenhum voto desde o
// inicio do recorte. Lista nao vazia depois de uma carga indica carga
// incompleta (item 4).
func (s *SyncService) SenadoresSemVotos() ([]string, error) {
	return s.repo.SenadoresEmExercicioSemVotos(utils.InicioRecorte())
}

func (s *SyncService) mapaSenadores() (map[int]int, error) {
	senadores, err := s.senadorRepo.FindAll(true)
	if err != nil {
		return nil, err
	}
	m := make(map[int]int, len(senadores))
	for _, sen := range senadores {
		m[sen.CodigoParlamentar] = sen.ID
	}
	return m, nil
}

// janelasMensais quebra [inicio, fim] em intervalos de um mes-calendario.
func janelasMensais(inicio, fim time.Time) [][2]time.Time {
	var janelas [][2]time.Time
	inicio = time.Date(inicio.Year(), inicio.Month(), inicio.Day(), 0, 0, 0, 0, time.UTC)
	fim = time.Date(fim.Year(), fim.Month(), fim.Day(), 0, 0, 0, 0, time.UTC)
	for a := inicio; !a.After(fim); {
		b := time.Date(a.Year(), a.Month()+1, 1, 0, 0, 0, 0, time.UTC).AddDate(0, 0, -1)
		if b.After(fim) {
			b = fim
		}
		janelas = append(janelas, [2]time.Time{a, b})
		a = b.AddDate(0, 0, 1)
	}
	return janelas
}

// converterVotacao gera uma linha por voto de senador conhecido.
func converterVotacao(v senadoapi.VotacaoSessaoAPI, senadorPorCodigo map[int]int) ([]Votacao, int, error) {
	if v.CodigoSessaoVotacao == 0 {
		return nil, 0, fmt.Errorf("votacao sem codigoSessaoVotacao (sessao %d, %s)", v.CodigoSessao, v.Identificacao)
	}
	dia, err := time.Parse("2006-01-02", v.DataSessao)
	if err != nil {
		return nil, 0, fmt.Errorf("data invalida %q na votacao %d: %w", v.DataSessao, v.CodigoSessaoVotacao, err)
	}
	// Meio-dia UTC para a data nao mudar de dia em nenhum fuso
	data := time.Date(dia.Year(), dia.Month(), dia.Day(), 12, 0, 0, 0, time.UTC)
	codigoSessao := strconv.Itoa(v.CodigoSessao)
	sigla := strings.ToUpper(strings.TrimSpace(v.Sigla))
	if sigla == "" {
		sigla = siglaDaIdentificacao(v.Identificacao)
	}
	secreta := votacaoSecreta(v.VotacaoSecreta)

	var votos []Votacao
	var ignorados int
	for _, voto := range v.Votos {
		senadorID, ok := senadorPorCodigo[voto.CodigoParlamentar]
		if !ok {
			ignorados++
			slog.Debug("voto de parlamentar fora da tabela senadores", "codigo", voto.CodigoParlamentar, "nome", voto.NomeParlamentar)
			continue
		}
		if voto.SiglaVoto == "" {
			continue
		}
		votos = append(votos, Votacao{
			SenadorID:         senadorID,
			CodigoVotacao:     v.CodigoSessaoVotacao,
			SessaoID:          codigoSessao,
			CodigoSessao:      codigoSessao,
			SequencialVotacao: v.SequencialVotacao,
			Data:              data,
			SiglaVoto:         voto.SiglaVoto,
			Voto:              rotuloVoto(voto.SiglaVoto),
			DescricaoVotacao:  v.DescricaoVotacao,
			Materia:           v.Identificacao,
			Ementa:            v.Ementa,
			Resultado:         v.ResultadoVotacao,
			SiglaMateria:      sigla,
			Secreta:           secreta,
			CodigoMateria:     positivoOuNulo(v.CodigoMateria),
			IdProcesso:        positivoOuNulo(v.IdProcesso),
		})
	}
	return votos, ignorados, nil
}

// siglaDaIdentificacao extrai a sigla do tipo da materia da identificacao
// ("PLP 124/2022 (Substitutivo-CD)" -> "PLP"). Mesma regra do UPDATE de
// PreencherHistorico.
func siglaDaIdentificacao(identificacao string) string {
	campos := strings.Fields(identificacao)
	if len(campos) == 0 {
		return ""
	}
	sigla := strings.ToUpper(campos[0])
	if len(sigla) > 20 {
		sigla = sigla[:20]
	}
	return sigla
}

// votacaoSecreta traduz o campo votacaoSecreta da API ("S"/"N"). Valor
// desconhecido fica nulo e e resolvido por PreencherHistorico.
func votacaoSecreta(v string) *bool {
	switch strings.ToUpper(strings.TrimSpace(v)) {
	case "S", "SIM", "TRUE":
		b := true
		return &b
	case "N", "NAO", "NÃO", "FALSE":
		b := false
		return &b
	default:
		return nil
	}
}

// rotuloVoto normaliza o voto para exibicao e para os filtros do frontend.
// A sigla bruta fica em SiglaVoto; P-OD (presidente) nao vira mais Obstrucao.
func rotuloVoto(sigla string) string {
	switch sigla {
	case "Não", "Nao":
		return "Nao"
	case "Obstrução", "Obstrucao":
		return "Obstrucao"
	case "Abstenção", "Abstencao":
		return "Abstencao"
	default:
		return sigla
	}
}

// positivoOuNulo copia o ponteiro, tratando 0 como ausente
func positivoOuNulo(v *int) *int {
	if v == nil || *v <= 0 {
		return nil
	}
	c := *v
	return &c
}

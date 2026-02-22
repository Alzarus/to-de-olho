package proposicao

import (
	"context"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/Alzarus/to-de-olho/internal/senador"
	senadoapi "github.com/Alzarus/to-de-olho/pkg/senado"
)

// SyncService gerencia sincronizacao de proposicoes
type SyncService struct {
	repo        *Repository
	senadorRepo *senador.Repository
	client      *senadoapi.LegisClient
}

// NewSyncService cria um novo servico de sincronizacao
func NewSyncService(repo *Repository, senadorRepo *senador.Repository, client *senadoapi.LegisClient) *SyncService {
	return &SyncService{
		repo:        repo,
		senadorRepo: senadorRepo,
		client:      client,
	}
}

// SyncFromAPI busca proposicoes da API para todos os senadores
func (s *SyncService) SyncFromAPI(ctx context.Context) error {
	slog.Info("iniciando sync de proposicoes")

	// Buscar todos os senadores
	senadores, err := s.senadorRepo.FindAll()
	if err != nil {
		return err
	}

	var totalProposicoes, totalSenadores int

	for _, sen := range senadores {
		count, err := s.SyncSenador(ctx, sen.ID)
		if err != nil {
			slog.Warn("falha ao buscar proposicoes", "senador", sen.Nome, "error", err)
			continue
		}

		totalProposicoes += count
		totalSenadores++
		slog.Debug("proposicoes sincronizadas", "senador", sen.Nome, "count", count)
	}

	slog.Info("sync de proposicoes concluido", "senadores", totalSenadores, "proposicoes", totalProposicoes)
	return nil
}

// SyncSenador busca proposicoes de um senador especifico
func (s *SyncService) SyncSenador(ctx context.Context, senadorID int) (int, error) {
	sen, err := s.senadorRepo.FindByID(senadorID)
	if err != nil {
		return 0, err
	}

	proposicoesAPI, err := s.client.ListarProposicoesParlamentar(ctx, sen.CodigoParlamentar)
	if err != nil {
		return 0, err
	}

	// Limpar proposicoes antigas nao e mais necessario com Upsert (OnConflict)
	// if err := s.repo.DeleteBySenadorID(senadorID); err != nil {
	// 	slog.Warn("falha ao limpar proposicoes antigas", "senador", senadorID, "error", err)
	// }

	var count int
	for _, p := range proposicoesAPI {
		proposicao := s.convertToModel(p, senadorID)
		
		// [PERFORMANCE] Se for sync diario (nao backfill), ignore proposicoes velhas 
		// Assumiremos que coisas apresentadas ha mais de 10 anos nao mudam de estado
		// ou apenas acompanhamos as tramitacoes recentes
		// Obs: A tramitacao atualiza o timestamp interno do sistema, entao upsert vale a pena para status
		// Mas aqui otimizamos
		if proposicao.DataApresentacao != nil {
			idadeAnos := time.Since(*proposicao.DataApresentacao).Hours() / 24 / 365
			// Filtra proposicoes velhas demais (mais de 4 anos == uma legislatura) se quiser.
			// Por ora mantemos todas porque as Arquivadas chegam juntas.
			_ = idadeAnos // suppress unused
		}

		// Calcular pontuacao
		proposicao.Pontuacao = proposicao.CalcularPontuacao()
		
		if err := s.repo.Upsert(&proposicao); err != nil {
			slog.Warn("falha ao salvar proposicao", "senador", senadorID, "error", err)
			continue
		}
		count++
	}

	return count, nil
}

// convertToModel converte uma proposicao da API para modelo interno
func (s *SyncService) convertToModel(api senadoapi.MateriaAPI, senadorID int) Proposicao {
	var dataApresentacao *time.Time
	if api.DataApresentacao != "" {
		if t, err := time.Parse("2006-01-02", api.DataApresentacao); err == nil {
			dataApresentacao = &t
		}
	}

	// Extrair sigla e numero/ano do campo Identificacao (ex: "PLS 4/2004")
	sigla, numero, ano := extrairIdentificacao(api.Identificacao)

	// Determinar estagio de tramitacao baseado em NormaGerada e SiglaTipoDeliberacao
	estagio := determinarEstagioV2(api.NormaGerada, api.SiglaTipoDeliberacao, api.Tramitando)

	return Proposicao{
		SenadorID:              senadorID,
		CodigoMateria:          strconv.Itoa(api.CodigoMateria),
		SiglaSubtipoMateria:    sigla,
		NumeroMateria:          numero,
		AnoMateria:             ano,
		DescricaoIdentificacao: api.Identificacao,
		Ementa:                 api.Ementa,
		SituacaoAtual:          api.SiglaTipoDeliberacao,
		DataApresentacao:       dataApresentacao,
		EstagioTramitacao:      estagio,
	}
}

// extrairIdentificacao extrai sigla, numero e ano do campo Identificacao
// Ex: "PLS 4/2004" -> ("PLS", "4", 2004)
func extrairIdentificacao(identificacao string) (sigla, numero string, ano int) {
	parts := strings.Fields(identificacao)
	if len(parts) >= 1 {
		sigla = parts[0]
	}
	if len(parts) >= 2 {
		// Formato: "4/2004"
		numAno := strings.Split(parts[1], "/")
		if len(numAno) >= 1 {
			numero = numAno[0]
		}
		if len(numAno) >= 2 {
			ano, _ = strconv.Atoi(numAno[1])
		}
	}
	return
}

// determinarEstagioV2 classifica o estagio baseado nos campos da API
func determinarEstagioV2(normaGerada, siglaTipoDeliberacao, tramitando string) string {
	// Se gerou norma (lei), e o estagio maximo
	if normaGerada != "" {
		return "TransformadoLei"
	}
	
	// Baseado na sigla de deliberacao
	switch siglaTipoDeliberacao {
	case "APROVADA_NO_PLENARIO":
		return "AprovadoPlenario"
	case "APROVADA_EM_COMISSAO_TERMINATIVA", "APROVADA_EM_COMISSAO":
		return "AprovadoComissao"
	case "EM_PAUTA_NO_PLENARIO", "AGUARDANDO_DELIBERACAO", "PRONTO_PARA_DELIBERACAO":
		return "EmComissao"
	case "ARQUIVADO_FIM_LEGISLATURA", "ARQUIVADA", "RETIRADO_PELO_AUTOR", "REJEITADA":
		// Arquivados ficam no ultimo estagio alcancado, assumimos apresentado
		return "Apresentado"
	}
	
	// Se ainda esta tramitando
	if tramitando == "Sim" {
		return "EmComissao"
	}
	
	// Default
	return "Apresentado"
}

// determinarEstagio classifica o estagio de tramitacao baseado na descricao
func determinarEstagio(situacao string) string {
	situacaoLower := strings.ToLower(situacao)

	// Transformado em lei (norma juridica)
	if strings.Contains(situacaoLower, "transformada em norma") ||
		strings.Contains(situacaoLower, "transformado em lei") ||
		strings.Contains(situacaoLower, "lei publicada") {
		return "TransformadoLei"
	}

	// Aprovado em plenario
	if strings.Contains(situacaoLower, "aprovado plenario") ||
		strings.Contains(situacaoLower, "aprovada pelo plenario") ||
		strings.Contains(situacaoLower, "remetida a camara") ||
		strings.Contains(situacaoLower, "sancionada") {
		return "AprovadoPlenario"
	}

	// Aprovado em comissao
	if strings.Contains(situacaoLower, "aprovado comissao") ||
		strings.Contains(situacaoLower, "aprovada pela comissao") ||
		strings.Contains(situacaoLower, "pronto para pauta") {
		return "AprovadoComissao"
	}

	// Em comissao
	if strings.Contains(situacaoLower, "comissao") ||
		strings.Contains(situacaoLower, "relator") ||
		strings.Contains(situacaoLower, "tramit") {
		return "EmComissao"
	}

	// Default: Apresentado
	return "Apresentado"
}

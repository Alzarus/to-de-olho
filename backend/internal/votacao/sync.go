package votacao

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/Alzarus/to-de-olho/internal/senador"
	senadoapi "github.com/Alzarus/to-de-olho/pkg/senado"
)

// SyncService gerencia sincronizacao de votacoes
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

// SyncFromAPI busca votacoes da API para todos os senadores
func (s *SyncService) SyncFromAPI(ctx context.Context) error {
	slog.Info("iniciando sync de votacoes")

	// Buscar todos os senadores
	senadores, err := s.senadorRepo.FindAll(false)
	if err != nil {
		return err
	}

	var totalVotacoes, totalSenadores int

	for _, sen := range senadores {
		// A API retorna sessoes de votacao
		sessoes, err := s.client.ListarVotacoesParlamentar(ctx, sen.CodigoParlamentar)
		if err != nil {
			slog.Warn("falha ao buscar votacoes", "senador", sen.Nome, "error", err)
			continue
		}

		// Processar cada sessao
		for _, sessao := range sessoes {
			// Encontrar o voto do senador nesta sessao
			var siglaVoto string
			for _, voto := range sessao.Votos {
				if voto.CodigoParlamentar == sen.CodigoParlamentar {
					siglaVoto = voto.SiglaVoto
					break
				}
			}

			if siglaVoto == "" {
				continue // Senador nao votou nesta sessao
			}

			votacao := s.convertSessaoToVotacao(sessao, sen.ID, siglaVoto)
			if err := s.repo.Upsert(&votacao); err != nil {
				slog.Warn("falha ao salvar votacao", "senador", sen.ID, "error", err)
				continue
			}
			totalVotacoes++
		}

		totalSenadores++
		slog.Debug("votacoes sincronizadas", "senador", sen.Nome, "sessoes", len(sessoes))
	}

	slog.Info("sync de votacoes concluido", "senadores", totalSenadores, "votacoes", totalVotacoes)
	return nil
}

// SyncSenador busca votacoes de um senador especifico
func (s *SyncService) SyncSenador(ctx context.Context, senadorID int) (int, error) {
	sen, err := s.senadorRepo.FindByID(senadorID)
	if err != nil {
		return 0, err
	}

	sessoes, err := s.client.ListarVotacoesParlamentar(ctx, sen.CodigoParlamentar)
	if err != nil {
		return 0, err
	}

	anoAtual := time.Now().Year()

	var count int
	for _, sessao := range sessoes {
		// [PERFORMANCE] Evitar carregar sessoes antigas no sync diario/atualizacoes
		if sessao.Ano < anoAtual-2 {
			continue // Ja deve estar no banco e nao muda mais, se for preciso use backfill
		}

		var siglaVoto string
		for _, voto := range sessao.Votos {
			if voto.CodigoParlamentar == sen.CodigoParlamentar {
				siglaVoto = voto.SiglaVoto
				break
			}
		}

		if siglaVoto == "" {
			continue
		}

		votacao := s.convertSessaoToVotacao(sessao, sen.ID, siglaVoto)
		if err := s.repo.Upsert(&votacao); err != nil {
			continue
		}
		count++
	}

	return count, nil
}

// SyncMetadata busca dados ricos (Datas, Ementas) da lista master e atualiza o banco
func (s *SyncService) SyncMetadata(ctx context.Context, ano int) error {
	slog.Info("iniciando sync de metadados (batch)", "ano", ano)

	votacoesMaster, err := s.client.ListarVotacoesAno(ctx, ano)
	if err != nil {
		return err
	}
	slog.Info("sessoes encontradas na API", "total", len(votacoesMaster))

	count := 0
	for _, vMaster := range votacoesMaster {
		// Construir ID da Sessao (ex: "12345_2024") to match existing records
		sessaoID := strconv.Itoa(vMaster.CodigoSessao) + "_" + strconv.Itoa(vMaster.Ano)

	// Parse Date + Time
		var dataFinal time.Time
		
		// Try ISO 8601 first (API return for 2024 List)
		if t, err := time.Parse("2006-01-02T15:04:05", vMaster.DataSessao); err == nil {
			dataFinal = t
		} else if t, err := time.Parse("2006-01-02", vMaster.DataSessao); err == nil {
			// Fallback to noon UTC
			dataFinal = time.Date(t.Year(), t.Month(), t.Day(), 12, 0, 0, 0, time.UTC)
		}

		// Rich Description: prioritize ementa, fallback to identificacaoMateria
		materia := ""
		if vMaster.IdentificacaoMateria != "" {
			materia = vMaster.IdentificacaoMateria
		} else if vMaster.Materia.Sigla != "" {
			materia = fmt.Sprintf("%s %s/%s", vMaster.Materia.Sigla, vMaster.Materia.Numero, vMaster.Materia.Ano)
		}
		
		descricao := vMaster.DescricaoVotacao
		if vMaster.EmentaLegislativo != "" {
			descricao = vMaster.EmentaLegislativo
		}

		// Update all records with this SessaoID
		updates := map[string]interface{}{}
		if !dataFinal.IsZero() {
			updates["data"] = dataFinal
		}
		if materia != "" {
			updates["materia"] = materia
		}
		if descricao != "" {
			 updates["descricao_votacao"] = descricao
		}

		if len(updates) > 0 {
			if err := s.repo.UpdateMetadata(sessaoID, updates); err != nil {
				slog.Warn("falha ao atualizar metadata", "sessao", sessaoID, "err", err)
			} else {
				count++
			}
		}
	}
	
	s.normalizeVotesBatch()

	slog.Info("sync metadata concluido", "sessoes_atualizadas", count)
	return nil
}

func (s *SyncService) normalizeVotesBatch() {
	mappings := map[string]string{
		"Não":       "Nao",
		"Sim":       "Sim", 
		"Obstrução": "Obstrucao",
		"P-OD":      "Obstrucao",
		"MIS":       "Outros",
		"Lsp":       "Licenca", 
		"Abstenção": "Abstencao",
	}
	
	for old, new := range mappings {
		if err := s.repo.UpdateVoteBatch(old, new); err != nil {
			slog.Warn("falha ao normalizar voto", "old", old, "new", new, "error", err)
		}
	}
}


// convertSessaoToVotacao converte uma sessao de votacao para modelo interno
func (s *SyncService) convertSessaoToVotacao(sessao senadoapi.VotacaoSessaoAPI, senadorID int, siglaVoto string) Votacao {
	var data time.Time
	var parsed bool

	if sessao.DataSessao != "" {
		// Formato: YYYY-MM-DD ou DD/MM/YYYY
		if t, err := time.Parse("2006-01-02", sessao.DataSessao); err == nil {
			// Definir meio-dia para evitar problemas de fuso horario
			data = time.Date(t.Year(), t.Month(), t.Day(), 12, 0, 0, 0, time.UTC)
			parsed = true
		} else if t, err := time.Parse("02/01/2006", sessao.DataSessao); err == nil {
			data = time.Date(t.Year(), t.Month(), t.Day(), 12, 0, 0, 0, time.UTC)
			parsed = true
		}
	}

	// Fallback: se nao conseguiu parsear a data, usa o campo Ano da API
	// para criar uma data valida (1 de janeiro do ano)
	// Isso garante que filtros por ano funcionem corretamente
	anoFallback := sessao.Ano
	if anoFallback == 0 {
		// Extrair ano do sessao_id (formato: "123456_2023")
		parts := strings.Split(strconv.Itoa(sessao.CodigoSessao)+"_"+strconv.Itoa(sessao.Ano), "_")
		if len(parts) >= 2 {
			if parsed, err := strconv.Atoi(parts[len(parts)-1]); err == nil && parsed >= 1988 && parsed <= 2100 {
				anoFallback = parsed
			}
		}
	}
	if !parsed && anoFallback > 0 {
		// Meio-dia para evitar shift de timezone
		data = time.Date(anoFallback, time.January, 1, 12, 0, 0, 0, time.UTC)
	}

	return Votacao{
		SenadorID:        senadorID,
		SessaoID:         strconv.Itoa(sessao.CodigoSessao) + "_" + strconv.Itoa(sessao.Ano),
		CodigoSessao:     strconv.Itoa(sessao.CodigoSessao),
		Data:             data,
		Voto:             normalizeVoto(siglaVoto),
		DescricaoVotacao: sessao.DescricaoVotacao,
		Materia:          "",
	}
}

func normalizeVoto(voto string) string {
	switch voto {
	case "Não", "Nao":
		return "Nao"
	case "Sim":
		return "Sim"
	case "Obstrução", "P-OD":
		return "Obstrucao"
	case "Abstenção", "Abstencao":
		return "Abstencao"
	default:
		return voto
	}
}

package proposicao

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Alzarus/to-de-olho/internal/senador"
	"github.com/Alzarus/to-de-olho/pkg/retry"
	senadoapi "github.com/Alzarus/to-de-olho/pkg/senado"
)

// clienteLegis e o que o sync usa da API do Senado (interface para testes)
type clienteLegis interface {
	ListarProposicoesParlamentar(ctx context.Context, codigoParlamentar int) ([]senadoapi.MateriaAPI, error)
	ObterProcesso(ctx context.Context, idProcesso int) (*senadoapi.ProcessoDetalhe, error)
}

// SyncService gerencia sincronizacao de proposicoes
type SyncService struct {
	repo        *Repository
	senadorRepo *senador.Repository
	client      clienteLegis

	// detalhes de processo ja buscados nesta rodada: a mesma PEC com "e outros"
	// aparece na lista de ate 30 senadores
	detalhes map[int]*senadoapi.ProcessoDetalhe
}

// NewSyncService cria um novo servico de sincronizacao
func NewSyncService(repo *Repository, senadorRepo *senador.Repository, client *senadoapi.LegisClient) *SyncService {
	return &SyncService{
		repo:        repo,
		senadorRepo: senadorRepo,
		client:      client,
	}
}

// ResumoAutoria conta de onde veio a posicao de autoria de cada linha
type ResumoAutoria struct {
	Texto      int // extraida do texto da listagem
	Detalhe    int // extraida de /processo/{id}
	SemPosicao int // senador ausente de autoriaIniciativa
}

// SyncFromAPI busca proposicoes da API para todos os senadores em exercicio.
// Cada senador tem ate 3 tentativas (a API corta respostas grandes); se algum
// falhar em todas, devolve erro com os nomes, depois de processar os demais.
func (s *SyncService) SyncFromAPI(ctx context.Context) error {
	slog.Info("iniciando sync de proposicoes")

	senadores, err := s.senadorRepo.FindAll(false)
	if err != nil {
		return err
	}
	s.detalhes = make(map[int]*senadoapi.ProcessoDetalhe)

	var totalProposicoes int
	var resumo ResumoAutoria
	var falhas []string

	for _, sen := range senadores {
		var count int
		var r ResumoAutoria
		err := retry.WithRetry(ctx, 3, "proposicoes "+sen.Nome, func() error {
			var err error
			count, r, err = s.syncSenador(ctx, sen)
			return err
		})
		if err != nil {
			slog.Error("proposicoes do senador nao sincronizadas", "senador", sen.Nome, "error", err)
			falhas = append(falhas, sen.Nome)
			continue
		}
		totalProposicoes += count
		resumo.Texto += r.Texto
		resumo.Detalhe += r.Detalhe
		resumo.SemPosicao += r.SemPosicao
		slog.Debug("proposicoes sincronizadas", "senador", sen.Nome, "count", count)
	}

	slog.Info("sync de proposicoes concluido",
		"senadores", len(senadores)-len(falhas), "falhas", len(falhas), "proposicoes", totalProposicoes,
		"posicao_texto", resumo.Texto, "posicao_detalhe", resumo.Detalhe, "sem_posicao", resumo.SemPosicao)

	if len(falhas) > 0 {
		return fmt.Errorf("proposicoes de %d senadores nao sincronizadas: %s", len(falhas), strings.Join(falhas, ", "))
	}
	return nil
}

// SyncSenador busca proposicoes de um senador especifico
func (s *SyncService) SyncSenador(ctx context.Context, senadorID int) (int, error) {
	sen, err := s.senadorRepo.FindByID(senadorID)
	if err != nil {
		return 0, err
	}
	if s.detalhes == nil {
		s.detalhes = make(map[int]*senadoapi.ProcessoDetalhe)
	}
	count, _, err := s.syncSenador(ctx, *sen)
	return count, err
}

func (s *SyncService) syncSenador(ctx context.Context, sen senador.Senador) (int, ResumoAutoria, error) {
	proposicoesAPI, err := s.client.ListarProposicoesParlamentar(ctx, sen.CodigoParlamentar)
	if err != nil {
		return 0, ResumoAutoria{}, err
	}

	proposicoes, resumo, err := s.montarProposicoes(ctx, sen, proposicoesAPI)
	if err != nil {
		return 0, resumo, err
	}
	if err := s.repo.UpsertBatch(proposicoes); err != nil {
		return 0, resumo, fmt.Errorf("falha ao gravar proposicoes: %w", err)
	}
	return len(proposicoes), resumo, nil
}

// montarProposicoes converte a lista da API em linhas (uma por materia),
// com posicao de autoria e pontuacao
func (s *SyncService) montarProposicoes(ctx context.Context, sen senador.Senador, lista []senadoapi.MateriaAPI) ([]Proposicao, ResumoAutoria, error) {
	var resumo ResumoAutoria
	if err := s.buscarDetalhes(ctx, sen, lista); err != nil {
		return nil, resumo, err
	}

	vistas := make(map[int]bool, len(lista))
	proposicoes := make([]Proposicao, 0, len(lista))

	for _, api := range lista {
		if vistas[api.CodigoMateria] {
			continue // o upsert em lote falha com a mesma chave duas vezes
		}
		vistas[api.CodigoMateria] = true

		p := s.convertToModel(api, sen.ID)

		posicao, total, ok := PosicaoNoTexto(api.Autoria, sen.Nome)
		if ok {
			resumo.Texto++
		} else {
			var err error
			posicao, total, err = s.posicaoPeloDetalhe(api.ID, sen.CodigoParlamentar)
			if err != nil {
				return nil, resumo, fmt.Errorf("detalhe do processo %d: %w", api.ID, err)
			}
			if posicao == 0 {
				resumo.SemPosicao++
				slog.Warn("senador ausente da autoria do processo", "senador", sen.Nome, "processo", api.ID, "identificacao", api.Identificacao)
			} else {
				resumo.Detalhe++
			}
		}
		p.PosicaoAutoria, p.TotalAutores = intOuNulo(posicao), intOuNulo(total)
		p.Pontuacao = p.CalcularPontuacao()
		proposicoes = append(proposicoes, p)
	}
	return proposicoes, resumo, nil
}

// detalhesEmParalelo e quantos /processo/{id} sao buscados ao mesmo tempo.
// ~12% das materias caem no detalhe; em serie a carga completa levaria ~50 min.
const detalhesEmParalelo = 4

// buscarDetalhes busca, em paralelo, o detalhe das materias cuja posicao o
// texto nao resolve e que ainda nao estao no cache da rodada.
func (s *SyncService) buscarDetalhes(ctx context.Context, sen senador.Senador, lista []senadoapi.MateriaAPI) error {
	var pendentes []int
	pedidos := make(map[int]bool)
	for _, api := range lista {
		if _, _, ok := PosicaoNoTexto(api.Autoria, sen.Nome); ok {
			continue
		}
		if _, ok := s.detalhes[api.ID]; ok || pedidos[api.ID] {
			continue
		}
		pedidos[api.ID] = true
		pendentes = append(pendentes, api.ID)
	}
	if len(pendentes) == 0 {
		return nil
	}

	detalhes := make([]*senadoapi.ProcessoDetalhe, len(pendentes))
	erros := make([]error, len(pendentes))
	fila := make(chan int)
	var wg sync.WaitGroup
	for w := 0; w < detalhesEmParalelo; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := range fila {
				id := pendentes[i]
				// 5 tentativas (backoff ate 8s): com 4 em paralelo a API devolve 429
				// em ~1% das chamadas
				erros[i] = retry.WithRetry(ctx, 5, fmt.Sprintf("processo %d", id), func() error {
					var err error
					detalhes[i], err = s.client.ObterProcesso(ctx, id)
					return err
				})
			}
		}()
	}
	for i := range pendentes {
		fila <- i
	}
	close(fila)
	wg.Wait()

	for i, id := range pendentes {
		if erros[i] != nil {
			return fmt.Errorf("detalhe do processo %d: %w", id, erros[i])
		}
		s.detalhes[id] = detalhes[i]
	}
	return nil
}

// posicaoPeloDetalhe devolve a ordem do senador em autoriaIniciativa, a partir
// do detalhe ja buscado. posicao = 0 quando o senador nao esta na lista.
func (s *SyncService) posicaoPeloDetalhe(idProcesso, codigoParlamentar int) (posicao, total int, err error) {
	detalhe, ok := s.detalhes[idProcesso]
	if !ok || detalhe == nil {
		return 0, 0, fmt.Errorf("detalhe do processo %d nao carregado", idProcesso)
	}
	posicao, total = PosicaoNoDetalhe(detalhe.AutoriaIniciativa, codigoParlamentar)
	return posicao, total, nil
}

// PosicaoNoDetalhe devolve a ordem oficial do parlamentar e o total de autores.
func PosicaoNoDetalhe(autores []senadoapi.AutorIniciativa, codigoParlamentar int) (posicao, total int) {
	for _, a := range autores {
		if a.CodigoParlamentar != nil && *a.CodigoParlamentar == codigoParlamentar {
			posicao = a.Ordem
		}
	}
	return posicao, len(autores)
}

func intOuNulo(v int) *int {
	if v == 0 {
		return nil
	}
	return &v
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
		Autoria:                api.Autoria,
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

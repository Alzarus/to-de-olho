package emenda

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/Alzarus/to-de-olho/internal/senador"
	"github.com/Alzarus/to-de-olho/pkg/transparencia"
)

type SyncService struct {
	repo          *Repository
	senadorRepo   *senador.Repository
	transparencia *transparencia.Client
}

func NewSyncService(repo *Repository, senadorRepo *senador.Repository, apiKey string) *SyncService {
	return &SyncService{
		repo:          repo,
		senadorRepo:   senadorRepo,
		transparencia: transparencia.NewClient(apiKey),
	}
}

func (s *SyncService) SyncAll(ctx context.Context, ano int) error {
	senadores, err := s.senadorRepo.FindAll()
	if err != nil {
		return err
	}

	for _, sen := range senadores {
		if err := s.SyncSenador(ctx, sen, ano); err != nil {
			log.Printf("Erro sync emendas %s: %v", sen.Nome, err)
			continue
		}
		// Rate limit para não estourar a API (mesmo syncrono)
		time.Sleep(500 * time.Millisecond)
	}
	return nil
}

func (s *SyncService) SyncSenador(ctx context.Context, sen senador.Senador, ano int) error {
	totalImportado := 0

	nomesBusca := montarNomesBusca(sen)
	for _, nomeBusca := range nomesBusca {
		pagina := 1
		for {
			// Verifica cancelamento do contexto
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}

			emendasDTO, err := s.transparencia.GetEmendas(ano, nomeBusca, pagina)
			if err != nil {
				return fmt.Errorf("erro API transparencia: %w", err)
			}

			if len(emendasDTO) == 0 {
				break
			}

			for _, dto := range emendasDTO {
				valorEmp := transparencia.ParseMoney(dto.ValorEmpenhado)
				valorPago := transparencia.ParseMoney(dto.ValorPago) + transparencia.ParseMoney(dto.ValorRestoPago)

				emenda := Emenda{
					SenadorID:             uint(sen.ID),
					Ano:                   dto.Ano,
					Numero:                dto.CodigoEmenda,
					Tipo:                  dto.TipoEmenda,
					FuncionalProgramatica: fmt.Sprintf("%s - %s", dto.Funcao, dto.Subfuncao),
					Localidade:            dto.LocalidadeDoGasto,
					ValorEmpenhado:        valorEmp,
					ValorPago:             valorPago,
					DataUltimaAtualizacao: time.Now(),
				}

				if err := s.repo.Upsert(&emenda); err != nil {
					log.Printf("Erro ao salvar emenda %s: %v", emenda.Numero, err)
				} else {
					totalImportado++
				}
			}

			pagina++
			// Limite de segurança para loops infinitos (API mal comportada)
			if pagina > 100 {
				break
			}

			// Pequeno delay entre páginas
			time.Sleep(100 * time.Millisecond)
		}
	}

	if totalImportado > 0 {
		log.Printf("Sync Emendas: %s (%d) - %d importadas", sen.Nome, ano, totalImportado)
	}

	return nil
}

func montarNomesBusca(sen senador.Senador) []string {
	mapa := map[string]struct{}{}
	if sen.Nome != "" {
		mapa[normalizarNomeAutor(sen.Nome)] = struct{}{}
	}
	if sen.NomeCompleto != "" {
		mapa[normalizarNomeAutor(sen.NomeCompleto)] = struct{}{}
	}

	nomes := make([]string, 0, len(mapa))
	for nome := range mapa {
		nomes = append(nomes, nome)
	}
	return nomes
}

func normalizarNomeAutor(nome string) string {
	nome = strings.TrimSpace(nome)
	nome = strings.ToUpper(nome)
	replacer := strings.NewReplacer(
		"Á", "A", "À", "A", "Â", "A", "Ã", "A", "Ä", "A",
		"É", "E", "È", "E", "Ê", "E", "Ë", "E",
		"Í", "I", "Ì", "I", "Î", "I", "Ï", "I",
		"Ó", "O", "Ò", "O", "Ô", "O", "Õ", "O", "Ö", "O",
		"Ú", "U", "Ù", "U", "Û", "U", "Ü", "U",
		"Ç", "C",
	)
	return replacer.Replace(nome)
}

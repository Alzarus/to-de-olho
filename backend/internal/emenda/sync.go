package emenda

import (
	"context"
	"fmt"
	"log/slog"
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
	slog.Info("SyncAll Emendas: iniciando", "senadores", len(senadores), "ano", ano)

	sucessos := 0
	falhas := 0
	for _, sen := range senadores {
		if err := s.SyncSenador(ctx, sen, ano); err != nil {
			slog.Warn("falha sync emendas para senador",
				"senador", sen.Nome,
				"ano", ano,
				"erro", err,
			)
			falhas++
			continue
		}
		sucessos++
		// Rate limit para nao estourar a API
		time.Sleep(500 * time.Millisecond)
	}

	slog.Info("SyncAll Emendas: concluido",
		"ano", ano,
		"sucessos", sucessos,
		"falhas", falhas,
	)
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

			slog.Debug("consultando emendas", "autor", nomeBusca, "pagina", pagina)
			emendasDTO, err := s.transparencia.GetEmendasWithCtx(ctx, ano, nomeBusca, pagina)
			if err != nil {
				slog.Warn("erro API emendas apos retries",
					"senador", sen.Nome,
					"query", nomeBusca,
					"pagina", pagina,
					"erro", err,
				)
				return fmt.Errorf("erro API transparencia: %w", err)
			}

			slog.Debug("emendas recebidas",
				"senador", sen.Nome,
				"query", nomeBusca,
				"quantidade", len(emendasDTO),
			)

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
					slog.Warn("erro ao salvar emenda", "numero", emenda.Numero, "erro", err)
				} else {
					totalImportado++
				}
			}

			pagina++
			// Limite de seguranca para loops infinitos
			if pagina > 100 {
				break
			}

			// Pequeno delay entre paginas
			time.Sleep(100 * time.Millisecond)
		}
	}

	if totalImportado > 0 {
		slog.Info("emendas importadas", "senador", sen.Nome, "ano", ano, "total", totalImportado)
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

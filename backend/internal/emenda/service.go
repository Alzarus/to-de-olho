package emenda

import (
	"bufio"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/Alzarus/to-de-olho/internal/senador"
)

type Service struct {
	repo        *Repository
	senadorRepo *senador.Repository
}

func NewService(repo *Repository, senadorRepo *senador.Repository) *Service {
	return &Service{repo: repo, senadorRepo: senadorRepo}
}

func (s *Service) ListBySenador(senadorID uint, ano int) ([]Emenda, error) {
	return s.repo.ListBySenador(senadorID, ano)
}

func (s *Service) GetResumo(senadorID uint, ano int) (*ResumoEmendas, error) {
	return s.repo.GetResumo(senadorID, ano)
}

func (s *Service) ImportarCSV(caminho string) error {
	slog.Info("iniciando importacao de emendas", "arquivo", caminho)

	if strings.TrimSpace(caminho) == "" {
		return errors.New("caminho do CSV vazio")
	}

	arquivo, err := os.Open(caminho)
	if err != nil {
		return fmt.Errorf("falha ao abrir CSV: %w", err)
	}
	defer arquivo.Close()

	reader := bufio.NewReader(arquivo)
	primeiraLinha, err := reader.ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return fmt.Errorf("falha ao ler cabecalho do CSV: %w", err)
	}

	if strings.TrimSpace(primeiraLinha) == "" {
		return errors.New("arquivo CSV vazio")
	}

	delimitador := detectarDelimitador(primeiraLinha)
	csvReader := csv.NewReader(io.MultiReader(strings.NewReader(primeiraLinha), reader))
	csvReader.Comma = delimitador
	csvReader.FieldsPerRecord = -1
	csvReader.LazyQuotes = true

	cabecalho, err := csvReader.Read()
	if err != nil {
		return fmt.Errorf("falha ao ler cabecalho CSV: %w", err)
	}

	indices := map[string]int{}
	for i, col := range cabecalho {
		indices[normalizarChave(col)] = i
	}

	senadores, err := s.senadorRepo.FindAll()
	if err != nil {
		return fmt.Errorf("falha ao carregar senadores: %w", err)
	}

	mapaSenadores := map[string]uint{}
	for _, sen := range senadores {
		if sen.Nome != "" {
			mapaSenadores[normalizarChave(sen.Nome)] = uint(sen.ID)
		}
		if sen.NomeCompleto != "" {
			mapaSenadores[normalizarChave(sen.NomeCompleto)] = uint(sen.ID)
		}
	}

	importadas := 0
	ignoradas := 0

	for {
		linha, err := csvReader.Read()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return fmt.Errorf("erro ao ler CSV: %w", err)
		}

		nomeAutor := buscarValor(linha, indices, "nomeautor", "autor", "nome autor")
		if nomeAutor == "" {
			ignoradas++
			continue
		}
		senadorID, ok := mapaSenadores[normalizarChave(nomeAutor)]
		if !ok {
			ignoradas++
			continue
		}

		ano := parseAno(buscarValor(linha, indices, "ano"))
		if ano == 0 {
			ignoradas++
			continue
		}

		codigoEmenda := buscarValor(linha, indices, "codigoemenda", "codigo emenda", "codigo")
		numeroEmenda := buscarValor(linha, indices, "numeroemenda", "numero emenda", "numero")
		numero := codigoEmenda
		if numero == "" {
			numero = numeroEmenda
		}
		if numero == "" {
			ignoradas++
			continue
		}

		funcao := buscarValor(linha, indices, "funcao", "fun\u00e7\u00e3o")
		subfuncao := buscarValor(linha, indices, "subfuncao", "subfun\u00e7\u00e3o")
		funcional := strings.TrimSpace(strings.Trim(strings.Join([]string{funcao, subfuncao}, " - "), " - "))

		emenda := Emenda{
			SenadorID:             senadorID,
			Ano:                   ano,
			Numero:                numero,
			Tipo:                  buscarValor(linha, indices, "tipoemenda", "tipo emenda", "tipo"),
			FuncionalProgramatica: funcional,
			Localidade:            buscarValor(linha, indices, "localidadedogasto", "localidade do gasto", "localidade"),
			ValorEmpenhado:        parseMoeda(buscarValor(linha, indices, "valorempenhado", "valor empenhado")),
			ValorPago:             parseMoeda(buscarValor(linha, indices, "valorpago", "valor pago")) + parseMoeda(buscarValor(linha, indices, "valorrestopago", "valor resto pago")),
			DataUltimaAtualizacao: time.Now(),
		}

		if err := s.repo.Upsert(&emenda); err != nil {
			slog.Warn("falha ao salvar emenda", "numero", numero, "erro", err)
			ignoradas++
			continue
		}
		importadas++
	}

	slog.Info("importacao de emendas concluida", "importadas", importadas, "ignoradas", ignoradas)
	return nil
}

func detectarDelimitador(linha string) rune {
	if strings.Count(linha, ";") >= strings.Count(linha, ",") {
		return ';'
	}
	return ','
}

func normalizarChave(valor string) string {
	valor = strings.TrimSpace(strings.ToLower(valor))
	valor = substituirAcentos(valor)
	replacer := strings.NewReplacer(" ", "", "_", "", "-", "", ".", "", "\t", "")
	valor = replacer.Replace(valor)
	return valor
}

func substituirAcentos(texto string) string {
	replacer := strings.NewReplacer(
		"\u00e1", "a", "\u00e0", "a", "\u00e2", "a", "\u00e3", "a", "\u00e4", "a",
		"\u00e9", "e", "\u00e8", "e", "\u00ea", "e", "\u00eb", "e",
		"\u00ed", "i", "\u00ec", "i", "\u00ee", "i", "\u00ef", "i",
		"\u00f3", "o", "\u00f2", "o", "\u00f4", "o", "\u00f5", "o", "\u00f6", "o",
		"\u00fa", "u", "\u00f9", "u", "\u00fb", "u", "\u00fc", "u",
		"\u00e7", "c",
	)
	return replacer.Replace(texto)
}

func buscarValor(linha []string, indices map[string]int, chaves ...string) string {
	for _, chave := range chaves {
		if idx, ok := indices[normalizarChave(chave)]; ok {
			if idx >= 0 && idx < len(linha) {
				return strings.TrimSpace(linha[idx])
			}
		}
	}
	return ""
}

func parseAno(valor string) int {
	ano, err := strconv.Atoi(strings.TrimSpace(valor))
	if err != nil {
		return 0
	}
	return ano
}

func parseMoeda(valor string) float64 {
	if valor == "" {
		return 0
	}
	limpo := strings.ReplaceAll(valor, "R$", "")
	limpo = strings.ReplaceAll(limpo, " ", "")
	limpo = strings.ReplaceAll(limpo, ".", "")
	limpo = strings.ReplaceAll(limpo, ",", ".")

	parsed, err := strconv.ParseFloat(limpo, 64)
	if err != nil {
		return 0
	}
	return parsed
}

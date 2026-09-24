package gabinete

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/Alzarus/to-de-olho/internal/senador"
	"github.com/Alzarus/to-de-olho/internal/utils"
	"github.com/Alzarus/to-de-olho/pkg/senado"
)

// paralelismo limita as chamadas simultaneas a recursos-utilizados: a rota e
// lenta (ate dezenas de segundos) e nao convem sobrecarregar a fonte.
const paralelismo = 4

// FonteRecursos e a parte do AdmClient usada pelo sync (permite teste sem rede)
type FonteRecursos interface {
	RecursosUtilizados(ctx context.Context, codigo, ano int) (*senado.RecursosUtilizadosAPI, error)
}

// FonteMesa e a parte do LegisClient usada pelo sync
type FonteMesa interface {
	ComposicaoMesaSF(ctx context.Context) ([]senado.CargoMesaAPI, error)
}

// ListaSenadores devolve os senadores a sincronizar
type ListaSenadores interface {
	FindAll(includeInactive bool) ([]senador.Senador, error)
}

// SyncService carrega pessoal agregado e beneficios da API administrativa
type SyncService struct {
	repo      *Repository
	senadores ListaSenadores
	recursos  FonteRecursos
	mesa      FonteMesa
	agora     func() time.Time
}

// NewSyncService cria o servico de sincronizacao
func NewSyncService(repo *Repository, senadores ListaSenadores, recursos FonteRecursos, mesa FonteMesa) *SyncService {
	return &SyncService{repo: repo, senadores: senadores, recursos: recursos, mesa: mesa, agora: time.Now}
}

// ResultadoSync resume uma carga de um ano
type ResultadoSync struct {
	Ano        int
	Senadores  int
	ComPessoal int
	SemDados   int
	Falhas     int
	NaoAchados int
	Servidores int
	Duracao    time.Duration
}

// normalizarLocal traduz o local da fonte ("Gabinete", "Escritório(s) de Apoio")
func normalizarLocal(local string) (string, bool) {
	l := strings.ToLower(strings.TrimSpace(local))
	switch {
	case strings.Contains(l, "gabinete"):
		return LocalGabinete, true
	case strings.HasPrefix(l, "escrit"):
		return LocalEscritorio, true
	}
	return "", false
}

// vinculoDoProprioSenador identifica a linha que conta o proprio parlamentar
func vinculoDoProprioSenador(vinculo string) bool {
	return strings.EqualFold(strings.TrimSpace(vinculo), "parlamentar")
}

// converter transforma a resposta da API nas linhas gravadas. Soma vinculos
// repetidos no mesmo local, descarta o vinculo PARLAMENTAR (o proprio
// senador) e quantidades nao positivas.
func converter(r *senado.RecursosUtilizadosAPI) ([]Recurso, []Beneficio) {
	if r == nil {
		return nil, nil
	}
	type chave struct{ local, vinculo string }
	somas := map[chave]int{}
	var ordem []chave
	for _, p := range r.Pessoal {
		local, ok := normalizarLocal(p.Local)
		if !ok {
			slog.Warn("local de lotacao desconhecido na fonte, ignorado", "local", p.Local)
			continue
		}
		for _, v := range p.Vinculos {
			nome := strings.TrimSpace(v.Vinculo)
			if nome == "" || vinculoDoProprioSenador(nome) || v.Quantidade <= 0 {
				continue
			}
			k := chave{local, nome}
			if _, visto := somas[k]; !visto {
				ordem = append(ordem, k)
			}
			somas[k] += v.Quantidade
		}
	}
	recursos := make([]Recurso, 0, len(ordem))
	for _, k := range ordem {
		recursos = append(recursos, Recurso{Local: k.local, Vinculo: k.vinculo, Quantidade: somas[k]})
	}

	vistos := map[string]bool{}
	var beneficios []Beneficio
	for _, b := range r.Beneficios {
		tipo := strings.TrimSpace(b.Beneficio)
		if tipo == "" || vistos[tipo] {
			continue
		}
		vistos[tipo] = true
		beneficios = append(beneficios, Beneficio{Tipo: tipo, Utilizacao: strings.TrimSpace(b.Utilizacao)})
	}
	return recursos, beneficios
}

type coleta struct {
	senadorID  int
	recursos   []Recurso
	beneficios []Beneficio
}

// SyncAno carrega um ano para todos os senadores em exercicio. Busca tudo
// antes de gravar: se nenhum senador vier com pessoal, aborta sem apagar nada
// (mesmo cuidado do ceaps/sync.go com a API devolvendo vazio).
func (s *SyncService) SyncAno(ctx context.Context, ano int) (ResultadoSync, error) {
	inicio := s.agora()
	res := ResultadoSync{Ano: ano}

	lista, err := s.senadores.FindAll(false)
	if err != nil {
		return res, fmt.Errorf("listar senadores em exercicio: %w", err)
	}
	res.Senadores = len(lista)
	if len(lista) == 0 {
		return res, errors.New("nenhum senador em exercicio no banco")
	}

	var (
		mu      sync.Mutex
		wg      sync.WaitGroup
		sem     = make(chan struct{}, paralelismo)
		coletas []coleta
	)
	for _, sen := range lista {
		if ctx.Err() != nil {
			break
		}
		wg.Add(1)
		sem <- struct{}{}
		go func(sen senador.Senador) {
			defer wg.Done()
			defer func() { <-sem }()
			r, err := s.recursos.RecursosUtilizados(ctx, sen.CodigoParlamentar, ano)
			mu.Lock()
			defer mu.Unlock()
			switch {
			case errors.Is(err, senado.ErrSenadorNaoEncontrado):
				res.NaoAchados++
				return
			case err != nil:
				res.Falhas++
				slog.Warn("falha ao buscar gabinete", "codigo", sen.CodigoParlamentar, "ano", ano, "error", err)
				return
			}
			rec, ben := converter(r)
			if len(rec) == 0 && len(ben) == 0 {
				res.SemDados++
				return
			}
			if len(rec) > 0 {
				res.ComPessoal++
			}
			coletas = append(coletas, coleta{senadorID: sen.ID, recursos: rec, beneficios: ben})
		}(sen)
	}
	wg.Wait()
	if err := ctx.Err(); err != nil {
		return res, fmt.Errorf("sync de gabinete %d cancelado: %w", ano, err)
	}

	if res.ComPessoal == 0 {
		res.Duracao = s.agora().Sub(inicio)
		return res, fmt.Errorf("API devolveu pessoal vazio para todos os %d senadores em %d (falhas: %d): nada gravado",
			res.Senadores, ano, res.Falhas)
	}

	agora := s.agora()
	for _, c := range coletas {
		if err := s.repo.SubstituirSenadorAno(c.senadorID, ano, c.recursos, c.beneficios, agora); err != nil {
			return res, fmt.Errorf("gravar gabinete do senador %d em %d: %w", c.senadorID, ano, err)
		}
		for _, r := range c.recursos {
			res.Servidores += r.Quantidade
		}
	}
	res.Duracao = s.agora().Sub(inicio)
	slog.Info("sync de gabinete concluido", "ano", ano, "senadores", res.Senadores,
		"com_pessoal", res.ComPessoal, "sem_dados", res.SemDados, "nao_achados", res.NaoAchados,
		"falhas", res.Falhas, "servidores", res.Servidores, "duracao", res.Duracao.String())
	return res, nil
}

// SyncMesa grava a composicao atual da Mesa Diretora (so senadores do banco)
func (s *SyncService) SyncMesa(ctx context.Context) error {
	cargosAPI, err := s.mesa.ComposicaoMesaSF(ctx)
	if err != nil {
		return err
	}
	lista, err := s.senadores.FindAll(true)
	if err != nil {
		return fmt.Errorf("listar senadores: %w", err)
	}
	idPorCodigo := make(map[int]int, len(lista))
	for _, sen := range lista {
		idPorCodigo[sen.CodigoParlamentar] = sen.ID
	}
	var cargos []CargoMesa
	vistos := map[int]bool{}
	for _, c := range cargosAPI {
		id, ok := idPorCodigo[c.CodigoParlamentar]
		if !ok || vistos[id] {
			continue
		}
		vistos[id] = true
		cargos = append(cargos, CargoMesa{SenadorID: id, Cargo: c.Cargo})
	}
	return s.repo.SubstituirMesa(cargos, s.agora())
}

// AnosDoRecorte vai do ano de inicio do recorte do mandato ate o ano atual
func AnosDoRecorte(agora time.Time) []int {
	var anos []int
	for ano := utils.InicioRecorte().Year(); ano <= agora.Year(); ano++ {
		anos = append(anos, ano)
	}
	return anos
}

// SyncTodos carrega a Mesa e os anos informados (backfill). Erro num ano nao
// interrompe os demais; devolve o primeiro erro.
func (s *SyncService) SyncTodos(ctx context.Context, anos []int) error {
	var primeiro error
	if err := s.SyncMesa(ctx); err != nil {
		slog.Error("falha no sync da mesa diretora", "error", err)
		primeiro = err
	}
	for _, ano := range anos {
		if _, err := s.SyncAno(ctx, ano); err != nil {
			slog.Error("falha no sync de gabinete", "ano", ano, "error", err)
			if primeiro == nil {
				primeiro = err
			}
		}
	}
	return primeiro
}

// PrecisaAtualizar e a guarda do sync diario: roda se o ano nao tem dados ou
// se a ultima carga tem mais de intervalo (semanal na pratica).
func (s *SyncService) PrecisaAtualizar(ano int, intervalo time.Duration) (bool, error) {
	ultima, err := s.repo.UltimaAtualizacao(ano)
	if err != nil {
		return false, err
	}
	return ultima == nil || s.agora().Sub(*ultima) >= intervalo, nil
}

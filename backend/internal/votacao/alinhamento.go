package votacao

import (
	"errors"
	"fmt"
	"math"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// Alinhamento entre senadores: em quantas votacoes nominais abertas cada par
// votou igual. So contam votos efetivos (Sim, Nao, Abstencao) dos dois: fica
// de fora a votacao secreta (o voto individual nao e publicado) e o registro
// de presenca sem voto (P-NRV, presidencia, obstrucao, licencas, faltas).

// votosEfetivos sao as siglas da API que expressam um voto conhecido
var votosEfetivos = []string{"Sim", "Não", "Abstenção"}

const (
	// minAlinhamento e maxAlinhamento limitam o numero de senadores (o comparador aceita ate 5)
	minAlinhamento = 2
	maxAlinhamento = 5
	// maxDivergencias limita a lista de votacoes com divergencia (as mais recentes)
	maxDivergencias = 300
)

// ParAlinhamento e a concordancia de um par de senadores (SenadorA < SenadorB).
type ParAlinhamento struct {
	SenadorA       int `json:"senador_a"`
	SenadorB       int `json:"senador_b"`
	VotacoesComuns int `json:"votacoes_comuns"` // os dois com voto efetivo
	VotosIguais    int `json:"votos_iguais"`
	// Percentual = votos_iguais / votacoes_comuns * 100; nulo sem votacao em comum
	Percentual *float64 `json:"percentual"`
}

// VotoSenador e o voto efetivo de um senador numa votacao.
type VotoSenador struct {
	SenadorID int    `json:"senador_id"`
	Voto      string `json:"voto"` // Sim, Não ou Abstenção
}

// Divergencia e uma votacao aberta em que ao menos dois dos senadores votaram diferente.
type Divergencia struct {
	CodigoVotacao    int           `json:"codigo_votacao"`
	Data             time.Time     `json:"data"`
	Materia          string        `json:"materia"`
	SiglaMateria     string        `json:"sigla_materia"`
	DescricaoVotacao string        `json:"descricao_votacao"`
	Resultado        string        `json:"resultado"`
	Votos            []VotoSenador `json:"votos"`
}

// ResultadoAlinhamento e a resposta de GET /votacoes/alinhamento.
type ResultadoAlinhamento struct {
	Ano               int              `json:"ano"` // 0: todos os anos
	Tipos             []string         `json:"tipos"`
	Senadores         []int            `json:"senadores"`
	Pares             []ParAlinhamento `json:"pares"`
	Divergencias      []Divergencia    `json:"divergencias"`
	TotalDivergencias int              `json:"total_divergencias"` // pode passar de len(Divergencias)
}

// FiltroAlinhamento reune os parametros do calculo.
type FiltroAlinhamento struct {
	IDs   []int
	Ano   int      // 0: todos
	Tipos []string // siglas da materia (PEC, PL...); vazio: todas
}

// votosFiltrados monta o recorte comum: votos efetivos, abertos, dos senadores e filtros pedidos.
func (r *Repository) votosFiltrados(f FiltroAlinhamento) (string, []any) {
	where := []string{
		"senador_id IN ?",
		"sigla_voto IN ?",
		// secreta nula (linhas antigas) com Sim/Nao/Abstencao so pode ser aberta:
		// na secreta a sigla e "Votou"
		"secreta IS NOT TRUE",
	}
	args := []any{f.IDs, votosEfetivos}
	if f.Ano > 0 {
		where = append(where, "EXTRACT(YEAR FROM data) = ?")
		args = append(args, f.Ano)
	}
	if len(f.Tipos) > 0 {
		where = append(where, "sigla_materia IN ?")
		args = append(args, f.Tipos)
	}
	sql := `SELECT senador_id, codigo_votacao, sigla_voto, data,
		COALESCE(materia, '') AS materia, COALESCE(sigla_materia, '') AS sigla_materia,
		COALESCE(descricao_votacao, '') AS descricao_votacao, COALESCE(resultado, '') AS resultado
		FROM votacoes WHERE ` + strings.Join(where, " AND ")
	return sql, args
}

// Alinhamento calcula a concordancia de cada par e lista as votacoes com divergencia.
func (r *Repository) Alinhamento(f FiltroAlinhamento) (*ResultadoAlinhamento, error) {
	base, args := r.votosFiltrados(f)

	var contagens []ParAlinhamento
	err := r.db.Raw(`WITH v AS (`+base+`)
		SELECT a.senador_id AS senador_a, b.senador_id AS senador_b,
			COUNT(*) AS votacoes_comuns,
			COUNT(*) FILTER (WHERE a.sigla_voto = b.sigla_voto) AS votos_iguais
		FROM v a JOIN v b ON a.codigo_votacao = b.codigo_votacao AND a.senador_id < b.senador_id
		GROUP BY a.senador_id, b.senador_id`, args...).Scan(&contagens).Error
	if err != nil {
		return nil, fmt.Errorf("contar votos iguais por par: %w", err)
	}

	porPar := make(map[[2]int]ParAlinhamento, len(contagens))
	for _, p := range contagens {
		porPar[[2]int{p.SenadorA, p.SenadorB}] = p
	}
	ids := append([]int(nil), f.IDs...)
	sort.Ints(ids)
	pares := make([]ParAlinhamento, 0, len(ids)*(len(ids)-1)/2)
	for i := 0; i < len(ids); i++ {
		for j := i + 1; j < len(ids); j++ {
			p, ok := porPar[[2]int{ids[i], ids[j]}]
			if !ok {
				p = ParAlinhamento{SenadorA: ids[i], SenadorB: ids[j]}
			}
			if p.VotacoesComuns > 0 {
				pct := math.Round(float64(p.VotosIguais)/float64(p.VotacoesComuns)*1000) / 10
				p.Percentual = &pct
			}
			pares = append(pares, p)
		}
	}

	var total int64
	err = r.db.Raw(`WITH v AS (`+base+`)
		SELECT COUNT(*) FROM (
			SELECT codigo_votacao FROM v GROUP BY codigo_votacao
			HAVING COUNT(*) >= 2 AND COUNT(DISTINCT sigla_voto) > 1
		) d`, args...).Scan(&total).Error
	if err != nil {
		return nil, fmt.Errorf("contar votacoes com divergencia: %w", err)
	}

	type linha struct {
		SenadorID        int
		CodigoVotacao    int
		SiglaVoto        string
		Data             time.Time
		Materia          string
		SiglaMateria     string
		DescricaoVotacao string
		Resultado        string
	}
	var linhas []linha
	err = r.db.Raw(`WITH v AS (`+base+`),
		d AS (
			SELECT codigo_votacao, MAX(data) AS data FROM v GROUP BY codigo_votacao
			HAVING COUNT(*) >= 2 AND COUNT(DISTINCT sigla_voto) > 1
			ORDER BY MAX(data) DESC, codigo_votacao DESC
			LIMIT ?
		)
		SELECT v.* FROM v JOIN d ON d.codigo_votacao = v.codigo_votacao
		ORDER BY v.data DESC, v.codigo_votacao DESC, v.senador_id`,
		append(args, maxDivergencias)...).Scan(&linhas).Error
	if err != nil {
		return nil, fmt.Errorf("listar votacoes com divergencia: %w", err)
	}

	divergencias := []Divergencia{}
	for _, l := range linhas {
		n := len(divergencias)
		if n == 0 || divergencias[n-1].CodigoVotacao != l.CodigoVotacao {
			divergencias = append(divergencias, Divergencia{
				CodigoVotacao: l.CodigoVotacao, Data: l.Data, Materia: l.Materia,
				SiglaMateria: l.SiglaMateria, DescricaoVotacao: l.DescricaoVotacao,
				Resultado: l.Resultado, Votos: []VotoSenador{},
			})
			n++
		}
		divergencias[n-1].Votos = append(divergencias[n-1].Votos, VotoSenador{SenadorID: l.SenadorID, Voto: l.SiglaVoto})
	}

	tipos := f.Tipos
	if tipos == nil {
		tipos = []string{}
	}
	return &ResultadoAlinhamento{
		Ano: f.Ano, Tipos: tipos, Senadores: ids, Pares: pares,
		Divergencias: divergencias, TotalDivergencias: int(total),
	}, nil
}

// parseIDs le "1,2,3": inteiros positivos, sem repetidos, entre 2 e 5.
func parseIDs(valor string) ([]int, error) {
	var ids []int
	vistos := map[int]bool{}
	for _, item := range strings.Split(valor, ",") {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		id, err := strconv.Atoi(item)
		if err != nil || id <= 0 {
			return nil, fmt.Errorf("id de senador invalido %q", item)
		}
		if !vistos[id] {
			vistos[id] = true
			ids = append(ids, id)
		}
	}
	if len(ids) < minAlinhamento || len(ids) > maxAlinhamento {
		return nil, errors.New("informe de 2 a 5 ids de senadores distintos em ids=")
	}
	return ids, nil
}

// GetAlinhamento godoc
// @Summary Alinhamento de votos entre senadores (votacoes nominais abertas)
// @Description Para cada par, conta as votacoes abertas em que os dois votaram Sim, Nao ou Abstencao e quantas vezes votaram igual.
// @Tags votacoes
// @Produce json
// @Param ids query string true "IDs dos senadores separados por virgula (2 a 5)"
// @Param ano query int false "Ano (default: todos)"
// @Param tipo query string false "Siglas da materia separadas por virgula (PEC,PL...)"
// @Success 200 {object} ResultadoAlinhamento
// @Failure 400 {object} map[string]interface{}
// @Router /api/v1/votacoes/alinhamento [get]
func (h *Handler) GetAlinhamento(c *gin.Context) {
	ids, err := parseIDs(c.Query("ids"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ano := 0
	if s := c.Query("ano"); s != "" {
		ano, err = strconv.Atoi(s)
		if err != nil || ano < 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "ano invalido"})
			return
		}
	}

	res, err := h.repo.Alinhamento(FiltroAlinhamento{IDs: ids, Ano: ano, Tipos: parseLista(c.Query("tipo"))})
	if err != nil {
		_ = c.Error(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "erro ao calcular alinhamento"})
		return
	}
	c.JSON(http.StatusOK, res)
}

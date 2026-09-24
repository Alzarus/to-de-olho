package gabinete

import (
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"time"

	"github.com/Alzarus/to-de-olho/internal/senador"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Handler expoe GET /api/v1/senadores/:id/gabinete
type Handler struct {
	repo        *Repository
	senadorRepo *senador.Repository
}

// NewHandler cria o handler
func NewHandler(repo *Repository, senadorRepo *senador.Repository) *Handler {
	return &Handler{repo: repo, senadorRepo: senadorRepo}
}

// rotulos dos locais na resposta, em pt-BR
var rotulos = map[string]string{
	LocalGabinete:   "Gabinete",
	LocalEscritorio: "Escritórios de apoio",
}

// FonteURL e a pagina oficial com o detalhe (nomes e cargos) do pessoal
func FonteURL(codigoParlamentar, ano int) string {
	return fmt.Sprintf("https://www6g.senado.leg.br/transparencia/sen/%d/pessoal/?local=gabinete&ano=%d", codigoParlamentar, ano)
}

// montarResumo agrega as linhas por local e por vinculo
func montarResumo(senadorID, ano int, anos []int, recursos []Recurso, beneficios []Beneficio) Resumo {
	res := Resumo{
		SenadorID:       senadorID,
		Ano:             ano,
		AnosDisponiveis: anos,
		Locais:          []LocalResumo{},
		PorVinculo:      []VinculoResumo{},
		Beneficios:      []BeneficioResumo{},
	}
	if res.AnosDisponiveis == nil {
		res.AnosDisponiveis = []int{}
	}

	porLocal := map[string]*LocalResumo{}
	porVinculo := map[string]int{}
	var atualizado time.Time
	for _, r := range recursos {
		l, ok := porLocal[r.Local]
		if !ok {
			l = &LocalResumo{Local: r.Local, Rotulo: rotulos[r.Local], Vinculos: []VinculoResumo{}}
			if l.Rotulo == "" {
				l.Rotulo = r.Local
			}
			porLocal[r.Local] = l
		}
		l.Vinculos = append(l.Vinculos, VinculoResumo{Vinculo: r.Vinculo, Quantidade: r.Quantidade})
		l.Total += r.Quantidade
		porVinculo[r.Vinculo] += r.Quantidade
		res.Total += r.Quantidade
		if r.AtualizadoEm.After(atualizado) {
			atualizado = r.AtualizadoEm
		}
	}
	// gabinete primeiro, depois escritorios, depois qualquer outro
	for _, local := range []string{LocalGabinete, LocalEscritorio} {
		if l, ok := porLocal[local]; ok {
			res.Locais = append(res.Locais, *l)
			delete(porLocal, local)
		}
	}
	for _, l := range porLocal {
		res.Locais = append(res.Locais, *l)
	}
	for _, l := range res.Locais {
		sort.SliceStable(l.Vinculos, func(i, j int) bool {
			if l.Vinculos[i].Quantidade != l.Vinculos[j].Quantidade {
				return l.Vinculos[i].Quantidade > l.Vinculos[j].Quantidade
			}
			return l.Vinculos[i].Vinculo < l.Vinculos[j].Vinculo
		})
	}
	for v, q := range porVinculo {
		res.PorVinculo = append(res.PorVinculo, VinculoResumo{Vinculo: v, Quantidade: q})
	}
	sort.Slice(res.PorVinculo, func(i, j int) bool {
		if res.PorVinculo[i].Quantidade != res.PorVinculo[j].Quantidade {
			return res.PorVinculo[i].Quantidade > res.PorVinculo[j].Quantidade
		}
		return res.PorVinculo[i].Vinculo < res.PorVinculo[j].Vinculo
	})

	for _, b := range beneficios {
		res.Beneficios = append(res.Beneficios, BeneficioResumo{Tipo: b.Tipo, Utilizacao: b.Utilizacao})
		if b.AtualizadoEm.After(atualizado) {
			atualizado = b.AtualizadoEm
		}
	}
	if !atualizado.IsZero() {
		res.AtualizadoEm = &atualizado
	}
	return res
}

// GetBySenador godoc
// @Summary Estrutura de gabinete (numeros agregados) de um senador
// @Tags gabinete
// @Produce json
// @Param id path int true "ID do senador"
// @Param ano query int false "Ano (padrao: o mais recente com dados)"
// @Success 200 {object} Resumo
// @Router /api/v1/senadores/{id}/gabinete [get]
func (h *Handler) GetBySenador(c *gin.Context) {
	senadorID, err := strconv.Atoi(c.Param("id"))
	if err != nil || senadorID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID invalido"})
		return
	}
	ano := 0
	if s := c.Query("ano"); s != "" {
		v, err := strconv.Atoi(s)
		if err != nil || v < 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "ano invalido"})
			return
		}
		ano = v
	}

	sen, err := h.senadorRepo.FindByID(senadorID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "senador nao encontrado"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "falha ao buscar senador"})
		return
	}

	anos, err := h.repo.AnosDisponiveis(senadorID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "falha ao buscar gabinete"})
		return
	}
	// sem ano (ou 0 = mandato): o ano mais recente com dados
	if ano == 0 && len(anos) > 0 {
		ano = anos[0]
	}

	var recursos []Recurso
	var beneficios []Beneficio
	if ano > 0 {
		if recursos, err = h.repo.Recursos(senadorID, ano); err == nil {
			beneficios, err = h.repo.Beneficios(senadorID, ano)
		}
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "falha ao buscar gabinete"})
			return
		}
	}

	res := montarResumo(senadorID, ano, anos, recursos, beneficios)
	if ano > 0 {
		res.FonteURL = FonteURL(sen.CodigoParlamentar, ano)
	} else {
		res.FonteURL = FonteURL(sen.CodigoParlamentar, time.Now().Year())
	}
	if cargo, err := h.repo.CargoMesaAtual(senadorID); err == nil && cargo != nil {
		res.CargoMesa = &cargo.Cargo
	}
	c.JSON(http.StatusOK, res)
}

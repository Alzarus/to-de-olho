package votacao

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestGetByIDLegadoDevolveSessaoParaRedirecionar(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/votacoes/:id", NewHandler(nil).GetByID)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/votacoes/473484_2025", nil))

	var corpo map[string]string
	json.Unmarshal(w.Body.Bytes(), &corpo)
	if w.Code != http.StatusNotFound || corpo["sessao_id"] != "473484" {
		t.Errorf("id legado: status %d corpo %v", w.Code, corpo)
	}

	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/votacoes/abc", nil))
	if w.Code != http.StatusBadRequest {
		t.Errorf("id invalido: status %d", w.Code)
	}
}

func TestParseLista(t *testing.T) {
	casos := []struct {
		entrada  string
		esperado []string
	}{
		{"", nil},
		{"PEC", []string{"PEC"}},
		{"pec, msf ,PEC", []string{"PEC", "MSF"}},
		{",,PL,", []string{"PL"}},
		{"PL;DROP TABLE,A", []string{"A"}},
		{"R,A", []string{"R", "A"}},
	}
	for _, c := range casos {
		if obtido := parseLista(c.entrada); !reflect.DeepEqual(obtido, c.esperado) {
			t.Errorf("parseLista(%q) = %v; esperado %v", c.entrada, obtido, c.esperado)
		}
	}
}

func TestParseListaTemTeto(t *testing.T) {
	var itens []string
	for i := 0; i < 5000; i++ {
		itens = append(itens, fmt.Sprintf("S%d", i))
	}
	if obtido := parseLista(strings.Join(itens, ",")); len(obtido) != maxItensLista {
		t.Errorf("lista longa deve parar em %d itens; veio %d", maxItensLista, len(obtido))
	}
}

func TestParseSecreta(t *testing.T) {
	casos := []struct {
		entrada  string
		esperado *bool
		erro     bool
	}{
		{"", nil, false},
		{"true", boolPtr(true), false},
		{"false", boolPtr(false), false},
		{" 1 ", boolPtr(true), false},
		{"talvez", nil, true},
	}
	for _, c := range casos {
		obtido, err := parseSecreta(c.entrada)
		if (err != nil) != c.erro || !reflect.DeepEqual(obtido, c.esperado) {
			t.Errorf("parseSecreta(%q) = %v, %v; esperado %v (erro=%v)", c.entrada, obtido, err, c.esperado, c.erro)
		}
	}
}

func TestGetAllSecretaInvalidaDa400(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/votacoes", NewHandler(nil).GetAll)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/votacoes?secreta=talvez", nil))
	if w.Code != http.StatusBadRequest {
		t.Errorf("secreta invalida: status %d", w.Code)
	}
}

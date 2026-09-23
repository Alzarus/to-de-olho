package votacao

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
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

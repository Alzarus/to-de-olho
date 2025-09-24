package domain

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// PaginationRequest parâmetros de paginação
type PaginationRequest struct {
	Page   int    `form:"page" json:"page"`
	Limit  int    `form:"limit" json:"limit"`
	Cursor string `form:"cursor" json:"cursor"`
	SortBy string `form:"sort_by" json:"sort_by"`
	Order  string `form:"order" json:"order"`
}

// PaginationResponse resposta paginada
type PaginationResponse struct {
	Data       interface{} `json:"data"`
	Pagination Pagination  `json:"pagination"`
}

// Pagination metadados de paginação
type Pagination struct {
	Page       int    `json:"page"`
	Limit      int    `json:"limit"`
	Total      int64  `json:"total"`
	TotalPages int    `json:"total_pages"`
	HasNext    bool   `json:"has_next"`
	HasPrev    bool   `json:"has_prev"`
	NextCursor string `json:"next_cursor,omitempty"`
	PrevCursor string `json:"prev_cursor,omitempty"`
}

// CursorData dados do cursor para navegação
type CursorData struct {
	ID        string    `json:"id"`
	Timestamp time.Time `json:"timestamp"`
	SortValue string    `json:"sort_value"`
}

// AllowedSortColumns define as colunas permitidas para ordenação por tabela
var AllowedSortColumns = map[string][]string{
	"deputados": {
		"id", "nome", "cpf", "sigla_partido", "sigla_uf",
		"created_at", "updated_at", "url_foto",
	},
	"proposicoes": {
		"id", "numero", "ano", "tipo", "ementa", "keywords",
		"descricao_tramitacao", "descricao_situacao", "created_at", "updated_at",
	},
	"despesas": {
		"id", "deputado_id", "ano", "mes", "tipo_documento", "cod_documento",
		"data_documento", "num_documento", "valor_documento", "nome_fornecedor",
		"valor_liquido", "valor_bruto", "created_at", "updated_at",
	},
}

// ValidateSortColumn valida se a coluna de ordenação é permitida para a tabela
func ValidateSortColumn(table, column string) error {
	allowedColumns, exists := AllowedSortColumns[table]
	if !exists {
		return fmt.Errorf("tabela não suportada: %s", table)
	}

	// Converte para lowercase para comparação case-insensitive
	column = strings.ToLower(column)
	for _, allowed := range allowedColumns {
		if strings.ToLower(allowed) == column {
			return nil
		}
	}

	return fmt.Errorf("coluna de ordenação não permitida: %s para tabela %s", column, table)
}

// ValidateAndNormalize valida e normaliza parâmetros de paginação
func (pr *PaginationRequest) ValidateAndNormalize() {
	// Página mínima é 1
	if pr.Page < 1 {
		pr.Page = 1
	}

	// Limite padrão e máximo
	if pr.Limit <= 0 {
		pr.Limit = 20
	}
	if pr.Limit > 1000 {
		pr.Limit = 1000
	}

	// Ordem padrão
	if pr.Order != "asc" && pr.Order != "desc" {
		pr.Order = "desc"
	}

	// Campo de ordenação padrão
	if pr.SortBy == "" {
		pr.SortBy = "created_at"
	}
}

// GetOffset calcula offset para paginação por página
func (pr *PaginationRequest) GetOffset() int {
	return (pr.Page - 1) * pr.Limit
}

// DecodeCursor decodifica cursor base64
func (pr *PaginationRequest) DecodeCursor() (*CursorData, error) {
	if pr.Cursor == "" {
		return nil, nil
	}

	decoded, err := base64.URLEncoding.DecodeString(pr.Cursor)
	if err != nil {
		return nil, fmt.Errorf("cursor inválido: %w", err)
	}

	var cursorData CursorData
	if err := json.Unmarshal(decoded, &cursorData); err != nil {
		return nil, fmt.Errorf("formato cursor inválido: %w", err)
	}

	return &cursorData, nil
}

// EncodeCursor codifica dados para cursor base64
func EncodeCursor(id string, timestamp time.Time, sortValue string) string {
	cursorData := CursorData{
		ID:        id,
		Timestamp: timestamp,
		SortValue: sortValue,
	}

	data, err := json.Marshal(cursorData)
	if err != nil {
		return ""
	}

	return base64.URLEncoding.EncodeToString(data)
}

// BuildPagination constrói response de paginação
func BuildPagination(req *PaginationRequest, total int64, data interface{}) *PaginationResponse {
	totalPages := int((total + int64(req.Limit) - 1) / int64(req.Limit))

	pagination := Pagination{
		Page:       req.Page,
		Limit:      req.Limit,
		Total:      total,
		TotalPages: totalPages,
		HasNext:    req.Page < totalPages,
		HasPrev:    req.Page > 1,
	}

	return &PaginationResponse{
		Data:       data,
		Pagination: pagination,
	}
}

// BuildCursorPagination constrói paginação por cursor
func BuildCursorPagination(data interface{}, hasNext bool, nextCursor string) *PaginationResponse {
	pagination := Pagination{
		HasNext:    hasNext,
		NextCursor: nextCursor,
	}

	return &PaginationResponse{
		Data:       data,
		Pagination: pagination,
	}
}

// GetPaginationSQL gera SQL para paginação por página
func GetPaginationSQL(baseQuery string, req *PaginationRequest, table string) (string, error) {
	// Validar coluna de ordenação
	if err := ValidateSortColumn(table, req.SortBy); err != nil {
		return "", err
	}

	offset := req.GetOffset()

	query := fmt.Sprintf(`
		%s
		ORDER BY %s %s
		LIMIT %d OFFSET %d
	`, baseQuery, req.SortBy, req.Order, req.Limit, offset)

	return query, nil
}

// GetCursorSQL gera SQL para paginação por cursor
// IMPORTANT: Este método retorna apenas a query SQL com placeholders.
// Os valores cursorData.SortValue e cursorData.ID devem ser passados
// como argumentos para a query para evitar SQL injection.
func GetCursorSQL(baseQuery string, req *PaginationRequest, cursorData *CursorData, table string) (string, []interface{}, error) {
	// Validar coluna de ordenação
	if err := ValidateSortColumn(table, req.SortBy); err != nil {
		return "", nil, err
	}

	if cursorData == nil {
		query := fmt.Sprintf(`
			%s
			ORDER BY %s %s
			LIMIT %d
		`, baseQuery, req.SortBy, req.Order, req.Limit+1)
		return query, nil, nil
	}

	operator := ">"
	if req.Order == "desc" {
		operator = "<"
	}

	query := fmt.Sprintf(`
		%s
		AND (%s %s $1 OR (%s = $1 AND id > $2))
		ORDER BY %s %s
		LIMIT %d
	`, baseQuery, req.SortBy, operator, req.SortBy, req.SortBy, req.Order, req.Limit+1)

	args := []interface{}{cursorData.SortValue, cursorData.ID}
	return query, args, nil
}

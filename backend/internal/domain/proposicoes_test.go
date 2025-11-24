package domain

import (
	"testing"
	"time"
)

func TestProposicao_Validate(t *testing.T) {
	tests := []struct {
		name       string
		proposicao *Proposicao
		wantError  bool
		errorType  error
	}{
		{
			name: "proposição válida",
			proposicao: &Proposicao{
				ID:        1,
				SiglaTipo: "PL",
				Numero:    1234,
				Ano:       2024,
				Ementa:    "Dispõe sobre teste unitário",
			},
			wantError: false,
		},
		{
			name: "ID inválido - zero",
			proposicao: &Proposicao{
				ID:        0,
				SiglaTipo: "PL",
				Numero:    1234,
				Ano:       2024,
				Ementa:    "Teste",
			},
			wantError: true,
			errorType: ErrProposicaoIDInvalido,
		},
		{
			name: "ID inválido - negativo",
			proposicao: &Proposicao{
				ID:        -1,
				SiglaTipo: "PL",
				Numero:    1234,
				Ano:       2024,
				Ementa:    "Teste",
			},
			wantError: true,
			errorType: ErrProposicaoIDInvalido,
		},
		{
			name: "ementa vazia",
			proposicao: &Proposicao{
				ID:        1,
				SiglaTipo: "PL",
				Numero:    1234,
				Ano:       2024,
				Ementa:    "",
			},
			wantError: true,
			errorType: ErrProposicaoEmentaVazia,
		},
		{
			name: "ano inválido - muito antigo",
			proposicao: &Proposicao{
				ID:        1,
				SiglaTipo: "PL",
				Numero:    1234,
				Ano:       1987,
				Ementa:    "Teste",
			},
			wantError: true,
			errorType: ErrProposicaoAnoInvalido,
		},
		{
			name: "ano inválido - futuro",
			proposicao: &Proposicao{
				ID:        1,
				SiglaTipo: "PL",
				Numero:    1234,
				Ano:       time.Now().Year() + 2,
				Ementa:    "Teste",
			},
			wantError: true,
			errorType: ErrProposicaoAnoInvalido,
		},
		{
			name: "número inválido - zero",
			proposicao: &Proposicao{
				ID:        1,
				SiglaTipo: "PL",
				Numero:    0,
				Ano:       2024,
				Ementa:    "Teste",
			},
			wantError: true,
			errorType: ErrProposicaoNumeroInvalido,
		},
		{
			name: "número inválido - negativo",
			proposicao: &Proposicao{
				ID:        1,
				SiglaTipo: "PL",
				Numero:    -1,
				Ano:       2024,
				Ementa:    "Teste",
			},
			wantError: true,
			errorType: ErrProposicaoNumeroInvalido,
		},
		{
			name: "tipo inválido - vazio",
			proposicao: &Proposicao{
				ID:        1,
				SiglaTipo: "",
				Numero:    1234,
				Ano:       2024,
				Ementa:    "Teste",
			},
			wantError: true,
			errorType: ErrProposicaoTipoInvalido,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.proposicao.Validate()

			if tt.wantError && err == nil {
				t.Errorf("Expected error but got none")
				return
			}

			if !tt.wantError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
				return
			}

			if tt.wantError && tt.errorType != nil && err != tt.errorType {
				t.Errorf("Expected error %v but got %v", tt.errorType, err)
			}
		})
	}
}

func TestProposicao_GetIdentificacao(t *testing.T) {
	tests := []struct {
		name       string
		proposicao *Proposicao
		expected   string
	}{
		{
			name: "PL comum",
			proposicao: &Proposicao{
				SiglaTipo: "PL",
				Numero:    1234,
				Ano:       2024,
			},
			expected: "PL 1234/2024",
		},
		{
			name: "PEC",
			proposicao: &Proposicao{
				SiglaTipo: "PEC",
				Numero:    15,
				Ano:       2023,
			},
			expected: "PEC 15/2023",
		},
		{
			name: "MPV",
			proposicao: &Proposicao{
				SiglaTipo: "MPV",
				Numero:    1185,
				Ano:       2024,
			},
			expected: "MPV 1185/2024",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.proposicao.GetIdentificacao()
			if result != tt.expected {
				t.Errorf("Expected %s but got %s", tt.expected, result)
			}
		})
	}
}

func TestProposicao_TypeCheckers(t *testing.T) {
	tests := []struct {
		name      string
		siglaTipo string
		isEmenda  bool
		isProjeto bool
		isMedida  bool
	}{
		{
			name:      "PEC é emenda",
			siglaTipo: TipoProposicaoPEC,
			isEmenda:  true,
			isProjeto: false,
			isMedida:  false,
		},
		{
			name:      "PL é projeto",
			siglaTipo: TipoProposicaoPL,
			isEmenda:  false,
			isProjeto: true,
			isMedida:  false,
		},
		{
			name:      "PLP é projeto",
			siglaTipo: TipoProposicaoPLP,
			isEmenda:  false,
			isProjeto: true,
			isMedida:  false,
		},
		{
			name:      "MPV é medida provisória",
			siglaTipo: TipoProposicaoMPV,
			isEmenda:  false,
			isProjeto: false,
			isMedida:  true,
		},
		{
			name:      "PDC não é nenhum dos tipos específicos",
			siglaTipo: TipoProposicaoPDC,
			isEmenda:  false,
			isProjeto: false,
			isMedida:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Proposicao{SiglaTipo: tt.siglaTipo}

			if p.IsEmenda() != tt.isEmenda {
				t.Errorf("IsEmenda() = %v, want %v", p.IsEmenda(), tt.isEmenda)
			}

			if p.IsProjeto() != tt.isProjeto {
				t.Errorf("IsProjeto() = %v, want %v", p.IsProjeto(), tt.isProjeto)
			}

			if p.IsMedidaProvisoria() != tt.isMedida {
				t.Errorf("IsMedidaProvisoria() = %v, want %v", p.IsMedidaProvisoria(), tt.isMedida)
			}
		})
	}
}

func TestProposicao_GetDataApresentacaoTime(t *testing.T) {
	tests := []struct {
		name          string
		dataStr       string
		expectError   bool
		expectedYear  int
		expectedMonth int
		expectedDay   int
	}{
		{
			name:          "formato ISO completo",
			dataStr:       "2024-03-15T14:30:00",
			expectError:   false,
			expectedYear:  2024,
			expectedMonth: 3,
			expectedDay:   15,
		},
		{
			name:          "formato ISO sem segundos",
			dataStr:       "2024-03-15T14:30",
			expectError:   false,
			expectedYear:  2024,
			expectedMonth: 3,
			expectedDay:   15,
		},
		{
			name:          "formato ISO apenas data",
			dataStr:       "2024-03-15",
			expectError:   false,
			expectedYear:  2024,
			expectedMonth: 3,
			expectedDay:   15,
		},
		{
			name:          "formato brasileiro",
			dataStr:       "15/03/2024",
			expectError:   false,
			expectedYear:  2024,
			expectedMonth: 3,
			expectedDay:   15,
		},
		{
			name:        "data vazia",
			dataStr:     "",
			expectError: true,
		},
		{
			name:        "formato inválido",
			dataStr:     "15-03-2024",
			expectError: true,
		},
		{
			name:        "string inválida",
			dataStr:     "not-a-date",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Proposicao{DataApresentacao: tt.dataStr}

			result, err := p.GetDataApresentacaoTime()

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Expected no error but got: %v", err)
				return
			}

			if result.Year() != tt.expectedYear {
				t.Errorf("Year = %d, want %d", result.Year(), tt.expectedYear)
			}

			if int(result.Month()) != tt.expectedMonth {
				t.Errorf("Month = %d, want %d", int(result.Month()), tt.expectedMonth)
			}

			if result.Day() != tt.expectedDay {
				t.Errorf("Day = %d, want %d", result.Day(), tt.expectedDay)
			}
		})
	}
}

func TestProposicaoFilter_Validate(t *testing.T) {
	currentYear := time.Now().Year()

	tests := []struct {
		name      string
		filter    *ProposicaoFilter
		wantError bool
		errorType error
	}{
		{
			name: "filtro válido padrão",
			filter: &ProposicaoFilter{
				Limite: 20,
				Pagina: 1,
			},
			wantError: false,
		},
		{
			name: "limite zero deve usar padrão",
			filter: &ProposicaoFilter{
				Limite: 0,
				Pagina: 1,
			},
			wantError: false,
		},
		{
			name: "limite excedido",
			filter: &ProposicaoFilter{
				Limite: 150,
				Pagina: 1,
			},
			wantError: true,
			errorType: ErrProposicaoLimiteExcedido,
		},
		{
			name: "página zero deve usar padrão",
			filter: &ProposicaoFilter{
				Limite: 20,
				Pagina: 0,
			},
			wantError: false,
		},
		{
			name: "ano inválido - muito antigo",
			filter: &ProposicaoFilter{
				Ano:    &[]int{1987}[0],
				Limite: 20,
				Pagina: 1,
			},
			wantError: true,
			errorType: ErrProposicaoAnoInvalido,
		},
		{
			name: "ano inválido - futuro",
			filter: &ProposicaoFilter{
				Ano:    &[]int{currentYear + 2}[0],
				Limite: 20,
				Pagina: 1,
			},
			wantError: true,
			errorType: ErrProposicaoAnoInvalido,
		},
		{
			name: "número inválido",
			filter: &ProposicaoFilter{
				Numero: &[]int{-1}[0],
				Limite: 20,
				Pagina: 1,
			},
			wantError: true,
			errorType: ErrProposicaoNumeroInvalido,
		},
		{
			name: "ordem inválida",
			filter: &ProposicaoFilter{
				Ordem:  "INVALID",
				Limite: 20,
				Pagina: 1,
			},
			wantError: true,
			errorType: ErrProposicaoOrdemInvalida,
		},
		{
			name: "ordenar por inválido",
			filter: &ProposicaoFilter{
				OrdenarPor: "invalid_field",
				Limite:     20,
				Pagina:     1,
			},
			wantError: true,
			errorType: ErrProposicaoOrdenarPorInvalido,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.filter.Validate()

			if tt.wantError && err == nil {
				t.Errorf("Expected error but got none")
				return
			}

			if !tt.wantError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
				return
			}

			if tt.wantError && tt.errorType != nil && err != tt.errorType {
				t.Errorf("Expected error %v but got %v", tt.errorType, err)
			}

			// Verificar se valores padrão foram definidos
			if !tt.wantError {
				if tt.filter.Limite == 0 {
					t.Errorf("Limite should have been set to default value")
				}
				if tt.filter.Pagina == 0 {
					t.Errorf("Pagina should have been set to default value")
				}
			}
		})
	}
}

func TestProposicaoFilter_SetDefaults(t *testing.T) {
	tests := []struct {
		name     string
		input    *ProposicaoFilter
		expected *ProposicaoFilter
	}{
		{
			name:  "valores zero devem ser definidos para padrão",
			input: &ProposicaoFilter{},
			expected: &ProposicaoFilter{
				Limite:     20,
				Pagina:     1,
				Ordem:      "DESC",
				OrdenarPor: "id",
			},
		},
		{
			name: "valores existentes devem ser preservados",
			input: &ProposicaoFilter{
				Limite:     50,
				Pagina:     3,
				Ordem:      "ASC",
				OrdenarPor: "numero",
			},
			expected: &ProposicaoFilter{
				Limite:     50,
				Pagina:     3,
				Ordem:      "ASC",
				OrdenarPor: "numero",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.input.SetDefaults()

			if tt.input.Limite != tt.expected.Limite {
				t.Errorf("Limite = %d, want %d", tt.input.Limite, tt.expected.Limite)
			}

			if tt.input.Pagina != tt.expected.Pagina {
				t.Errorf("Pagina = %d, want %d", tt.input.Pagina, tt.expected.Pagina)
			}

			if tt.input.Ordem != tt.expected.Ordem {
				t.Errorf("Ordem = %s, want %s", tt.input.Ordem, tt.expected.Ordem)
			}

			if tt.input.OrdenarPor != tt.expected.OrdenarPor {
				t.Errorf("OrdenarPor = %s, want %s", tt.input.OrdenarPor, tt.expected.OrdenarPor)
			}
		})
	}
}

func TestProposicaoFilter_BuildAPIQueryParams(t *testing.T) {
	ano := 2024
	numero := 1234
	codSituacao := 100
	dataInicio := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	dataFim := time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name     string
		filter   *ProposicaoFilter
		expected map[string]string
	}{
		{
			name: "filtro básico",
			filter: &ProposicaoFilter{
				Pagina:     1,
				Limite:     20,
				Ordem:      "DESC",
				OrdenarPor: "id",
			},
			expected: map[string]string{
				"pagina":     "1",
				"itens":      "20",
				"ordem":      "DESC",
				"ordenarPor": "id",
			},
		},
		{
			name: "filtro completo",
			filter: &ProposicaoFilter{
				SiglaTipo:              "PL",
				Numero:                 &numero,
				Ano:                    &ano,
				DataApresentacaoInicio: &dataInicio,
				DataApresentacaoFim:    &dataFim,
				CodSituacao:            &codSituacao,
				SiglaUfAutor:           "SP",
				SiglaPartidoAutor:      "PT",
				NomeAutor:              "123",
				Tema:                   "saúde",
				Keywords:               "covid",
				Pagina:                 2,
				Limite:                 50,
				Ordem:                  "ASC",
				OrdenarPor:             "numero",
			},
			expected: map[string]string{
				"siglaTipo":              "PL",
				"numero":                 "1234",
				"ano":                    "2024",
				"dataApresentacaoInicio": "2024-01-01",
				"dataApresentacaoFim":    "2024-12-31",
				"codSituacao":            "100",
				"siglaUfAutor":           "SP",
				"siglaPartidoAutor":      "PT",
				"idAutor":                "123",
				"tema":                   "saúde",
				"keywords":               "covid",
				"pagina":                 "2",
				"itens":                  "50",
				"ordem":                  "ASC",
				"ordenarPor":             "numero",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.filter.BuildAPIQueryParams()

			// Verificar se todos os parâmetros esperados estão presentes
			for key, expectedValue := range tt.expected {
				if actualValue, exists := result[key]; !exists {
					t.Errorf("Missing parameter %s", key)
				} else if actualValue != expectedValue {
					t.Errorf("Parameter %s = %s, want %s", key, actualValue, expectedValue)
				}
			}

			// Verificar se não há parâmetros extras (apenas quando não há valores opcionais)
			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d parameters but got %d", len(tt.expected), len(result))
			}
		})
	}
}

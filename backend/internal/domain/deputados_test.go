package domain

import (
	"testing"
)

func TestDeputado_Validate(t *testing.T) {
	tests := []struct {
		name      string
		deputado  Deputado
		wantError bool
		errorMsg  string
	}{
		{
			name: "deputado válido completo",
			deputado: Deputado{
				ID:       123,
				Nome:     "João Silva",
				Partido:  "PT",
				UF:       "SP",
				URLFoto:  "https://example.com/foto.jpg",
				Situacao: "Exercício",
				Email:    "joao@camara.leg.br",
			},
			wantError: false,
		},
		{
			name: "deputado sem nome - inválido",
			deputado: Deputado{
				ID:      123,
				Nome:    "",
				Partido: "PT",
				UF:      "SP",
			},
			wantError: true,
			errorMsg:  "nome é obrigatório",
		},
		{
			name: "deputado sem partido - inválido",
			deputado: Deputado{
				ID:   123,
				Nome: "João Silva",
				UF:   "SP",
			},
			wantError: true,
			errorMsg:  "partido é obrigatório",
		},
		{
			name: "UF inválida",
			deputado: Deputado{
				ID:      123,
				Nome:    "João Silva",
				Partido: "PT",
				UF:      "XX",
			},
			wantError: true,
			errorMsg:  "UF inválida",
		},
		{
			name: "ID zero - inválido",
			deputado: Deputado{
				ID:      0,
				Nome:    "João Silva",
				Partido: "PT",
				UF:      "SP",
			},
			wantError: true,
			errorMsg:  "ID deve ser maior que zero",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.deputado.Validate()

			if tt.wantError {
				if err == nil {
					t.Errorf("esperava erro mas não recebeu nenhum")
					return
				}
				if tt.errorMsg != "" && err.Error() != tt.errorMsg {
					t.Errorf("erro esperado: %q, recebido: %q", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("não esperava erro, mas recebeu: %v", err)
				}
			}
		})
	}
}

func TestDeputado_GetNomeCompleto(t *testing.T) {
	tests := []struct {
		name     string
		deputado Deputado
		expected string
	}{
		{
			name: "nome com partido e UF",
			deputado: Deputado{
				Nome:    "João Silva",
				Partido: "PT",
				UF:      "SP",
			},
			expected: "João Silva (PT/SP)",
		},
		{
			name: "nome sem partido",
			deputado: Deputado{
				Nome: "Maria Santos",
				UF:   "RJ",
			},
			expected: "Maria Santos (/RJ)",
		},
		{
			name: "nome apenas",
			deputado: Deputado{
				Nome: "Pedro Oliveira",
			},
			expected: "Pedro Oliveira (/)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.deputado.GetNomeCompleto()
			if result != tt.expected {
				t.Errorf("esperado: %q, recebido: %q", tt.expected, result)
			}
		})
	}
}

func TestDespesa_Validate(t *testing.T) {
	tests := []struct {
		name      string
		despesa   Despesa
		wantError bool
		errorMsg  string
	}{
		{
			name: "despesa válida",
			despesa: Despesa{
				Ano:               2024,
				Mes:               6,
				TipoDespesa:       "COMBUSTÍVEIS E LUBRIFICANTES",
				CodDocumento:      12345,
				TipoDocumento:     "Nota Fiscal",
				DataDocumento:     "2024-06-15",
				NumDocumento:      "NF-001",
				ValorLiquido:      150.75,
				NomeFornecedor:    "Posto XYZ Ltda",
				CNPJCPFFornecedor: "12.345.678/0001-90",
			},
			wantError: false,
		},
		{
			name: "ano inválido - muito antigo",
			despesa: Despesa{
				Ano:          1990,
				Mes:          6,
				ValorLiquido: 100.0,
			},
			wantError: true,
			errorMsg:  "ano deve ser entre 2000 e ano atual",
		},
		{
			name: "mês inválido - zero",
			despesa: Despesa{
				Ano:          2024,
				Mes:          0,
				ValorLiquido: 100.0,
			},
			wantError: true,
			errorMsg:  "mês deve ser entre 1 e 12",
		},
		{
			name: "valor negativo permitido",
			despesa: Despesa{
				Ano:          2024,
				Mes:          6,
				ValorLiquido: -50.0,
			},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.despesa.Validate()

			if tt.wantError {
				if err == nil {
					t.Errorf("esperava erro mas não recebeu nenhum")
					return
				}
				if tt.errorMsg != "" && err.Error() != tt.errorMsg {
					t.Errorf("erro esperado: %q, recebido: %q", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("não esperava erro, mas recebeu: %v", err)
				}
			}
		})
	}
}

func TestDespesa_GetMesNome(t *testing.T) {
	tests := []struct {
		name     string
		mes      int
		expected string
	}{
		{name: "janeiro", mes: 1, expected: "Janeiro"},
		{name: "junho", mes: 6, expected: "Junho"},
		{name: "dezembro", mes: 12, expected: "Dezembro"},
		{name: "mês inválido", mes: 13, expected: "Mês Inválido"},
		{name: "mês zero", mes: 0, expected: "Mês Inválido"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			despesa := Despesa{Mes: tt.mes}
			result := despesa.GetMesNome()
			if result != tt.expected {
				t.Errorf("esperado: %q, recebido: %q", tt.expected, result)
			}
		})
	}
}

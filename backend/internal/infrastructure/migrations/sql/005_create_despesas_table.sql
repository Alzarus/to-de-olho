-- Migration: 005_create_despesas_table.sql
-- Descrição: Cria tabela para despesas dos deputados com otimizações para analytics

CREATE TABLE IF NOT EXISTS despesas (
    id BIGSERIAL PRIMARY KEY,
    deputado_id INTEGER NOT NULL,
    ano INTEGER NOT NULL,
    mes INTEGER NOT NULL,
    cod_documento INTEGER,
    tipo_despesa VARCHAR(100) NOT NULL,
    tipo_documento VARCHAR(50),
    data_documento DATE,
    num_documento VARCHAR(50),
    valor_liquido DECIMAL(15,2) NOT NULL DEFAULT 0,
    valor_bruto DECIMAL(15,2) DEFAULT 0,
    valor_glosa DECIMAL(15,2) DEFAULT 0,
    nome_fornecedor VARCHAR(255),
    cnpj_cpf_fornecedor VARCHAR(20),
    url_documento TEXT,
    num_ressarcimento VARCHAR(50),
    cod_lote INTEGER,
    parcela INTEGER,
    payload JSONB NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Índices otimizados para queries de analytics
CREATE INDEX IF NOT EXISTS idx_despesas_deputado_ano ON despesas(deputado_id, ano DESC);
CREATE INDEX IF NOT EXISTS idx_despesas_ano_mes ON despesas(ano DESC, mes DESC);
CREATE INDEX IF NOT EXISTS idx_despesas_tipo ON despesas(tipo_despesa);
CREATE INDEX IF NOT EXISTS idx_despesas_valor ON despesas(valor_liquido DESC);
CREATE INDEX IF NOT EXISTS idx_despesas_fornecedor ON despesas(nome_fornecedor);
CREATE INDEX IF NOT EXISTS idx_despesas_created_at ON despesas(created_at DESC);

-- Índice composto para rankings de gastos
CREATE INDEX IF NOT EXISTS idx_despesas_ranking ON despesas(ano DESC, valor_liquido DESC, deputado_id);

-- Constraint para garantir valores positivos
ALTER TABLE despesas ADD CONSTRAINT check_valor_liquido_positive 
    CHECK (valor_liquido >= 0);

-- Constraint para anos válidos
ALTER TABLE despesas ADD CONSTRAINT check_ano_valid 
    CHECK (ano >= 2000 AND ano <= EXTRACT(YEAR FROM CURRENT_DATE) + 1);

-- Constraint para meses válidos
ALTER TABLE despesas ADD CONSTRAINT check_mes_valid 
    CHECK (mes >= 1 AND mes <= 12);

-- Comentários para documentação
COMMENT ON TABLE despesas IS 'Despesas parlamentares dos deputados da Câmara';
COMMENT ON COLUMN despesas.deputado_id IS 'ID do deputado na API da Câmara';
COMMENT ON COLUMN despesas.valor_liquido IS 'Valor líquido da despesa em reais';
COMMENT ON COLUMN despesas.payload IS 'Dados completos da despesa em formato JSON original';
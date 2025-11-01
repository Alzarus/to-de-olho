-- Migration 014: adiciona colunas exigidas pelo pipeline de despesas
-- Garante compatibilidade com o domínio e o DespesaRepository

ALTER TABLE despesas
    ADD COLUMN IF NOT EXISTS cod_tipo_documento INTEGER;

ALTER TABLE despesas
    ADD COLUMN IF NOT EXISTS valor_documento DECIMAL(15,2);

UPDATE despesas
SET valor_documento = COALESCE(valor_documento, valor_liquido)
WHERE valor_documento IS NULL;

ALTER TABLE despesas
    ALTER COLUMN valor_documento SET DEFAULT 0;

ALTER TABLE despesas
    ALTER COLUMN valor_documento SET NOT NULL;

COMMENT ON COLUMN despesas.cod_tipo_documento IS 'Código do tipo de documento conforme API da Câmara';
COMMENT ON COLUMN despesas.valor_documento IS 'Valor total do documento original (sem descontos)';

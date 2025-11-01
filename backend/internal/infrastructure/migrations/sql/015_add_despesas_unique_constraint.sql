-- Migration 015: garante constraint única para despesas e remove índice parcial obsoleto

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'uq_despesas_deputado_ano_cod_documento'
    ) THEN
        ALTER TABLE despesas
            ADD CONSTRAINT uq_despesas_deputado_ano_cod_documento
            UNIQUE (deputado_id, ano, cod_documento);
    END IF;
END
$$;

DROP INDEX IF EXISTS ux_despesas_deputado_ano_cod_documento;

COMMENT ON CONSTRAINT uq_despesas_deputado_ano_cod_documento ON despesas IS 'Assegura unicidade de documentos por deputado/ano.';

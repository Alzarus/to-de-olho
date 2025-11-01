-- Migration 016: ajusta constraint de valor líquido para permitir estornos negativos da Câmara

DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'check_valor_liquido_positive'
    ) THEN
        ALTER TABLE despesas DROP CONSTRAINT check_valor_liquido_positive;
    END IF;
END
$$;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'check_valor_liquido_range'
    ) THEN
        ALTER TABLE despesas
            ADD CONSTRAINT check_valor_liquido_range
            CHECK (valor_liquido >= -1000000000);
    END IF;
END
$$;

COMMENT ON CONSTRAINT check_valor_liquido_range ON despesas IS 'Permite valores negativos em estornos, limitando a até -1 bilhão.';

-- SAFE COMMAND: operações idempotentes para manter compatibilidade com replays de migração.

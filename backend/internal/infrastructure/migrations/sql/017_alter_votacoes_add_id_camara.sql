-- Migration: 017_alter_votacoes_add_id_camara.sql
-- Descrição: adiciona coluna id_camara textual e torna id_votacao_camara opcional para suportar IDs alfanuméricos

BEGIN;

ALTER TABLE votacoes
    ADD COLUMN IF NOT EXISTS id_camara VARCHAR(64);

UPDATE votacoes
   SET id_camara = COALESCE(id_camara, id_votacao_camara::text)
 WHERE id_camara IS NULL;

ALTER TABLE votacoes
    ALTER COLUMN id_camara SET NOT NULL;

ALTER TABLE votacoes
    ALTER COLUMN id_votacao_camara DROP NOT NULL;

-- Remover unique constraint anterior baseada apenas no id numérico, se existir
DO $$
BEGIN
    IF EXISTS (
        SELECT 1
          FROM pg_constraint
         WHERE conname = 'votacoes_id_votacao_camara_key'
           AND conrelid = 'votacoes'::regclass
    ) THEN
        ALTER TABLE votacoes DROP CONSTRAINT votacoes_id_votacao_camara_key;
    END IF;
END $$;

CREATE UNIQUE INDEX IF NOT EXISTS idx_votacoes_id_camara
    ON votacoes (id_camara);

CREATE UNIQUE INDEX IF NOT EXISTS idx_votacoes_id_votacao_camara
    ON votacoes (id_votacao_camara)
    WHERE id_votacao_camara IS NOT NULL;

COMMIT;

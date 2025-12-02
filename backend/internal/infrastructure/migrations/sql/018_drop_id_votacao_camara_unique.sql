-- Migration: 018_drop_id_votacao_camara_unique.sql
-- Descrição: Remove índice único em id_votacao_camara que causa conflitos
-- Razão: Múltiplas votações podem ter o mesmo id_votacao_camara base (ex: 2345468-38 e 2345468-41 
--        compartilham o valor numérico 2345468). O identificador único correto é id_camara (string completa).

BEGIN;

DROP INDEX IF EXISTS idx_votacoes_id_votacao_camara;

-- Manter índice não-único para performance de consultas se necessário
CREATE INDEX IF NOT EXISTS idx_votacoes_id_votacao_camara_btree
    ON votacoes (id_votacao_camara)
    WHERE id_votacao_camara IS NOT NULL;

COMMIT;

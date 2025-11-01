-- Migration: 006_add_despesas_to_sync_metrics.sql
-- Descrição: Adiciona campo despesas_updated na tabela sync_metrics

ALTER TABLE sync_metrics ADD COLUMN IF NOT EXISTS despesas_updated INTEGER DEFAULT 0;

-- Comentário para documentação
COMMENT ON COLUMN sync_metrics.despesas_updated IS 'Número de despesas atualizadas durante a sincronização';
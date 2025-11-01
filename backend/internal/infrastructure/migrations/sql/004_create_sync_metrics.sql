-- Migration: 004_create_sync_metrics.sql
-- Descrição: Cria tabela para métricas de sincronização incremental

CREATE TABLE IF NOT EXISTS sync_metrics (
    id BIGSERIAL PRIMARY KEY,
    sync_type VARCHAR(20) NOT NULL,                     -- 'daily', 'quick'
    start_time TIMESTAMP WITH TIME ZONE NOT NULL,
    end_time TIMESTAMP WITH TIME ZONE NOT NULL,
    duration_ms INTEGER NOT NULL,
    deputados_updated INTEGER DEFAULT 0,
    proposicoes_updated INTEGER DEFAULT 0,
    errors_count INTEGER DEFAULT 0,
    errors TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Índices para performance
CREATE INDEX IF NOT EXISTS idx_sync_metrics_type_time ON sync_metrics(sync_type, start_time DESC);
CREATE INDEX IF NOT EXISTS idx_sync_metrics_start_time ON sync_metrics(start_time DESC);

-- Comentários para documentação
COMMENT ON TABLE sync_metrics IS 'Métricas de sincronização incremental e diária';
COMMENT ON COLUMN sync_metrics.sync_type IS 'Tipo de sincronização: daily ou quick';
COMMENT ON COLUMN sync_metrics.duration_ms IS 'Duração da sincronização em milissegundos';
COMMENT ON COLUMN sync_metrics.errors IS 'Lista de erros ocorridos durante a sincronização';
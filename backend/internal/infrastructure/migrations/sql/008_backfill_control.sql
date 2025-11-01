-- Migration 008: Sistema de Controle de Backfill
-- ================================================

-- Tabela para controlar execuções de backfill
CREATE TABLE IF NOT EXISTS backfill_executions (
    id SERIAL PRIMARY KEY,
    execution_id UUID UNIQUE NOT NULL DEFAULT gen_random_uuid(),
    tipo VARCHAR(50) NOT NULL, -- 'historico', 'incremental', 'manual'
    ano_inicio INTEGER NOT NULL,
    ano_fim INTEGER NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'running', -- 'running', 'success', 'failed', 'partial'
    
    -- Métricas de execução
    deputados_processados INTEGER DEFAULT 0,
    proposicoes_processadas INTEGER DEFAULT 0,
    despesas_processadas INTEGER DEFAULT 0,
    votacoes_processadas INTEGER DEFAULT 0,
    
    -- Controle temporal
    started_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMPTZ,
    duration_seconds INTEGER,
    
    -- Metadados
    triggered_by VARCHAR(100) DEFAULT 'scheduler', -- 'scheduler', 'manual', 'deploy'
    error_message TEXT,
    config JSONB, -- Configurações específicas da execução
    
    -- Índices para consulta eficiente
    CONSTRAINT valid_status CHECK (status IN ('running', 'success', 'failed', 'partial')),
    CONSTRAINT valid_years CHECK (ano_inicio <= ano_fim AND ano_inicio >= 1988)
);

-- Índices para performance
CREATE INDEX idx_backfill_status_type ON backfill_executions(status, tipo);
CREATE INDEX idx_backfill_years ON backfill_executions(ano_inicio, ano_fim);
CREATE INDEX idx_backfill_started_at ON backfill_executions(started_at DESC);

-- Função para verificar se backfill histórico foi concluído
CREATE OR REPLACE FUNCTION has_successful_historical_backfill(start_year INTEGER, end_year INTEGER DEFAULT NULL)
RETURNS BOOLEAN AS $$
BEGIN
    IF end_year IS NULL THEN
        end_year := EXTRACT(YEAR FROM NOW());
    END IF;
    
    RETURN EXISTS (
        SELECT 1 FROM backfill_executions 
        WHERE tipo = 'historico' 
        AND status = 'success'
        AND ano_inicio <= start_year 
        AND ano_fim >= end_year
        AND completed_at IS NOT NULL
    );
END;
$$ LANGUAGE plpgsql;

-- Função para obter última execução por tipo
CREATE OR REPLACE FUNCTION get_last_backfill_execution(execution_type VARCHAR DEFAULT 'historico')
RETURNS TABLE (
    execution_id UUID,
    status VARCHAR,
    ano_inicio INTEGER,
    ano_fim INTEGER,
    started_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    duration_seconds INTEGER,
    deputados_processados INTEGER,
    proposicoes_processadas INTEGER,
    despesas_processadas INTEGER,
    votacoes_processadas INTEGER
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        be.execution_id,
        be.status,
        be.ano_inicio,
        be.ano_fim,
        be.started_at,
        be.completed_at,
        be.duration_seconds,
        be.deputados_processados,
        be.proposicoes_processadas,
        be.despesas_processadas,
        be.votacoes_processadas
    FROM backfill_executions be
    WHERE be.tipo = execution_type
    ORDER BY be.started_at DESC
    LIMIT 1;
END;
$$ LANGUAGE plpgsql;

-- Comentários para documentação
COMMENT ON TABLE backfill_executions IS 'Controla execuções de backfill histórico e incremental';
COMMENT ON FUNCTION has_successful_historical_backfill IS 'Verifica se backfill histórico foi concluído com sucesso para período específico';
COMMENT ON FUNCTION get_last_backfill_execution IS 'Retorna informações da última execução de backfill por tipo';
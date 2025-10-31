-- Migration: 009_scheduler_control
-- Autor: Sistema Inteligente de Monitoramento
-- Data: 2025-09-24
-- Descrição: Controla execuções de sincronizações incrementais do scheduler
-- NOTE: o migrator registra a versão automaticamente; não inserir na tabela schema_migrations

-- Necessário para gen_random_uuid()
CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- Tabela para rastrear execuções do scheduler
CREATE TABLE IF NOT EXISTS scheduler_executions (
    id SERIAL PRIMARY KEY,
    execution_id UUID NOT NULL DEFAULT gen_random_uuid(),
    tipo VARCHAR(20) NOT NULL CHECK (tipo IN ('diario','rapido','manual','inicial')),
    status VARCHAR(20) NOT NULL DEFAULT 'running' CHECK (status IN ('running','success','failed','partial')),

    -- Métricas de execução
    deputados_sincronizados INTEGER DEFAULT 0,
    proposicoes_sincronizadas INTEGER DEFAULT 0,
    despesas_sincronizadas INTEGER DEFAULT 0,
    votacoes_sincronizadas INTEGER DEFAULT 0,

    -- Controle temporal
    started_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMP WITH TIME ZONE,
    duration_seconds INTEGER,
    next_execution TIMESTAMP WITH TIME ZONE,

    -- Metadados e configuração
    triggered_by VARCHAR(50) DEFAULT 'cron',
    config JSONB DEFAULT '{}',
    error_message TEXT,

    -- Auditoria
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_scheduler_executions_status ON scheduler_executions(status);
CREATE INDEX IF NOT EXISTS idx_scheduler_executions_tipo ON scheduler_executions(tipo);
CREATE INDEX IF NOT EXISTS idx_scheduler_executions_started_at ON scheduler_executions(started_at DESC);
CREATE INDEX IF NOT EXISTS idx_scheduler_executions_execution_id ON scheduler_executions(execution_id);
CREATE INDEX IF NOT EXISTS idx_scheduler_executions_next ON scheduler_executions(next_execution) WHERE status = 'success';

-- Trigger para atualizar updated_at e calcular duration_seconds
CREATE OR REPLACE FUNCTION update_scheduler_executions_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();

    IF NEW.completed_at IS NOT NULL AND (OLD.completed_at IS NULL) THEN
        NEW.duration_seconds = EXTRACT(EPOCH FROM (NEW.completed_at - NEW.started_at))::INTEGER;
    END IF;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS scheduler_executions_updated_at_trigger ON scheduler_executions;
CREATE TRIGGER scheduler_executions_updated_at_trigger
    BEFORE UPDATE ON scheduler_executions
    FOR EACH ROW
    EXECUTE FUNCTION update_scheduler_executions_updated_at();

-- Função para obter última execução bem-sucedida
CREATE OR REPLACE FUNCTION get_last_successful_scheduler_execution(scheduler_tipo TEXT DEFAULT NULL)
RETURNS TABLE (
    execution_id UUID,
    tipo VARCHAR(20),
    completed_at TIMESTAMP WITH TIME ZONE,
    duration_seconds INTEGER,
    deputados_sincronizados INTEGER,
    proposicoes_sincronizadas INTEGER,
    despesas_sincronizadas INTEGER,
    votacoes_sincronizadas INTEGER
) AS $$
BEGIN
    RETURN QUERY
    SELECT
        s.execution_id,
        s.tipo,
        s.completed_at,
        s.duration_seconds,
        s.deputados_sincronizados,
        s.proposicoes_sincronizadas,
        s.despesas_sincronizadas,
        s.votacoes_sincronizadas
    FROM scheduler_executions s
    WHERE s.status = 'success'
      AND (scheduler_tipo IS NULL OR s.tipo = scheduler_tipo)
    ORDER BY s.completed_at DESC
    LIMIT 1;
END;
$$ LANGUAGE plpgsql;

-- Função que decide se o scheduler deve executar com base no intervalo mínimo
CREATE OR REPLACE FUNCTION should_scheduler_execute(
    scheduler_tipo TEXT,
    min_interval_hours INTEGER DEFAULT 1
) RETURNS TABLE (
    should_run BOOLEAN,
    reason TEXT,
    last_execution TIMESTAMP WITH TIME ZONE,
    hours_since_last NUMERIC
) AS $$
DECLARE
    last_exec TIMESTAMP WITH TIME ZONE;
    hours_diff NUMERIC;
BEGIN
    SELECT completed_at INTO last_exec
    FROM scheduler_executions
    WHERE tipo = scheduler_tipo AND status = 'success'
    ORDER BY completed_at DESC
    LIMIT 1;

    IF last_exec IS NULL THEN
        RETURN QUERY SELECT
            true::BOOLEAN,
            'Primeira execução do scheduler tipo: ' || scheduler_tipo,
            NULL::TIMESTAMP WITH TIME ZONE,
            NULL::NUMERIC;
        RETURN;
    END IF;

    hours_diff := EXTRACT(EPOCH FROM (NOW() - last_exec)) / 3600.0;

    IF hours_diff >= min_interval_hours THEN
        RETURN QUERY SELECT
            true::BOOLEAN,
            'Intervalo mínimo atingido (' || hours_diff::TEXT || 'h >= ' || min_interval_hours::TEXT || 'h)',
            last_exec,
            hours_diff;
    ELSE
        RETURN QUERY SELECT
            false::BOOLEAN,
            'Aguardando intervalo mínimo (' || hours_diff::TEXT || 'h < ' || min_interval_hours::TEXT || 'h)',
            last_exec,
            hours_diff;
    END IF;
END;
$$ LANGUAGE plpgsql;

-- Função de limpeza para execuções antigas
CREATE OR REPLACE FUNCTION cleanup_old_scheduler_executions()
RETURNS INTEGER AS $$
DECLARE
    deleted_count INTEGER;
BEGIN
    DELETE FROM scheduler_executions
    WHERE started_at < NOW() - INTERVAL '30 days';

    GET DIAGNOSTICS deleted_count = ROW_COUNT;

    RETURN deleted_count;
END;
$$ LANGUAGE plpgsql;

-- Comentários de documentação
COMMENT ON TABLE scheduler_executions IS 'Rastreabilidade completa de execuções de scheduler (sincronizações incrementais)';
COMMENT ON COLUMN scheduler_executions.execution_id IS 'UUID único da execução para rastreamento';
COMMENT ON COLUMN scheduler_executions.tipo IS 'Tipo de scheduler: diario, rapido, manual, inicial';
COMMENT ON COLUMN scheduler_executions.next_execution IS 'Próxima execução programada pelo cron';
COMMENT ON COLUMN scheduler_executions.triggered_by IS 'O que disparou a execução: cron, manual, api, startup';
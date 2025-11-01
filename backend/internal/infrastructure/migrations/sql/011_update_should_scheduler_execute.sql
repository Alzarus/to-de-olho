-- Migration: 011_update_should_scheduler_execute
-- Data: 2025-09-26
-- Descrição: Atualiza should_scheduler_execute para checar execuções em andamento (status = 'running')

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
    running_count INTEGER;
BEGIN
    -- Se houver execução em andamento para este tipo, não executar
    SELECT COUNT(*) INTO running_count
    FROM scheduler_executions
    WHERE tipo = scheduler_tipo AND status = 'running';

    IF running_count > 0 THEN
        RETURN QUERY SELECT
            false::BOOLEAN,
            'Há uma execução em andamento para este tipo: ' || scheduler_tipo,
            NULL::TIMESTAMP WITH TIME ZONE,
            NULL::NUMERIC;
        RETURN;
    END IF;

    -- Caso contrário, aplicar a lógica antiga baseada na última execução bem sucedida
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
-- Safe command: add a comment describing the function
COMMENT ON FUNCTION should_scheduler_execute(TEXT, INTEGER) IS 'Decide se o scheduler deve executar; evita iniciar quando há execução em andamento.';

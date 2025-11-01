-- Migration: 003_create_backfill_checkpoints.sql
-- Descrição: Cria tabela para checkpoints do processo de backfill histórico

CREATE TABLE IF NOT EXISTS backfill_checkpoints (
    id VARCHAR(255) PRIMARY KEY,
    type VARCHAR(50) NOT NULL,                              -- 'deputados', 'proposicoes', 'despesas'
    status VARCHAR(20) NOT NULL DEFAULT 'pending',          -- 'pending', 'in_progress', 'completed', 'failed'
    progress JSONB NOT NULL DEFAULT '{}',                   -- Progresso serializado
    metadata JSONB NOT NULL DEFAULT '{}',                   -- Metadados adicionais
    started_at TIMESTAMP WITH TIME ZONE,                    -- Quando foi iniciado
    completed_at TIMESTAMP WITH TIME ZONE,                  -- Quando foi completado
    error_message TEXT,                                      -- Mensagem de erro se falhou
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),      -- Quando foi criado
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()       -- Última atualização
);

-- Índices para performance
CREATE INDEX IF NOT EXISTS idx_backfill_checkpoints_type ON backfill_checkpoints(type);
CREATE INDEX IF NOT EXISTS idx_backfill_checkpoints_status ON backfill_checkpoints(status);
CREATE INDEX IF NOT EXISTS idx_backfill_checkpoints_created_at ON backfill_checkpoints(created_at);

-- Comentários para documentação
COMMENT ON TABLE backfill_checkpoints IS 'Checkpoints do processo de backfill histórico para resumabilidade';
COMMENT ON COLUMN backfill_checkpoints.id IS 'ID único do checkpoint (tipo_timestamp)';
COMMENT ON COLUMN backfill_checkpoints.type IS 'Tipo de dados sendo processados';
COMMENT ON COLUMN backfill_checkpoints.status IS 'Status atual do checkpoint';
COMMENT ON COLUMN backfill_checkpoints.progress IS 'Progresso detalhado em JSON';
COMMENT ON COLUMN backfill_checkpoints.metadata IS 'Metadados específicos do tipo de backfill';
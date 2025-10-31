-- Migration: 002_create_proposicoes_cache.sql
-- Descrição: Cria tabela para cache de proposições da Câmara dos Deputados

-- Criar extensão pg_trgm se não existir (para busca textual)
CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE TABLE IF NOT EXISTS proposicoes_cache (
    id INTEGER PRIMARY KEY,
    payload JSONB NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Índices para melhorar performance das queries
CREATE INDEX IF NOT EXISTS idx_proposicoes_cache_updated_at ON proposicoes_cache(updated_at DESC);

-- Índices BTREE para campos de texto/número (mais eficientes para igualdade e range)
CREATE INDEX IF NOT EXISTS idx_proposicoes_cache_sigla_tipo ON proposicoes_cache((payload->>'siglaTipo'));
CREATE INDEX IF NOT EXISTS idx_proposicoes_cache_numero ON proposicoes_cache(((payload->>'numero')::int));
CREATE INDEX IF NOT EXISTS idx_proposicoes_cache_ano ON proposicoes_cache(((payload->>'ano')::int));
CREATE INDEX IF NOT EXISTS idx_proposicoes_cache_uf_autor ON proposicoes_cache((payload->>'siglaUfAutor'));
CREATE INDEX IF NOT EXISTS idx_proposicoes_cache_partido_autor ON proposicoes_cache((payload->>'siglaPartidoAutor'));

-- Índice GIN apenas para busca textual na ementa
CREATE INDEX IF NOT EXISTS idx_proposicoes_cache_ementa ON proposicoes_cache USING GIN ((payload->>'ementa') gin_trgm_ops);

-- Comentários para documentação
COMMENT ON TABLE proposicoes_cache IS 'Cache de proposições da Câmara dos Deputados para melhorar performance';
COMMENT ON COLUMN proposicoes_cache.id IS 'ID único da proposição na API da Câmara';
COMMENT ON COLUMN proposicoes_cache.payload IS 'Dados completos da proposição em formato JSON';
COMMENT ON COLUMN proposicoes_cache.updated_at IS 'Data e hora da última atualização do cache';
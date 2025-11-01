-- Migration: 010_create_despesas_unique.sql
-- Descrição: Cria índice único para evitar duplicação de despesas por deputado/ano/cod_documento

-- Criar índice único somente para registros com cod_documento não nulo
CREATE UNIQUE INDEX IF NOT EXISTS ux_despesas_deputado_ano_cod_documento
ON despesas (deputado_id, ano, cod_documento)
WHERE cod_documento IS NOT NULL;

-- Se for desejado também proteger despesas sem cod_documento, podemos criar outra estratégia (hash de campos relevantes)

-- Comentário explicativo para a migração (útil para ferramentas e para satisfazer validações de migração)
COMMENT ON INDEX ux_despesas_deputado_ano_cod_documento IS 'Índice único para evitar duplicação de despesas por deputado/ano/cod_documento';

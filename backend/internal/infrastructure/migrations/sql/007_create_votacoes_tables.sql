-- Migration: 007_create_votacoes_tables.sql
-- Descrição: Cria tabelas para sistema de votações da Câmara dos Deputados

-- Tabela principal de votações
CREATE TABLE IF NOT EXISTS votacoes (
    id BIGSERIAL PRIMARY KEY,
    id_votacao_camara BIGINT UNIQUE NOT NULL,
    titulo VARCHAR(500) NOT NULL,
    ementa TEXT,
    data_votacao TIMESTAMP WITH TIME ZONE NOT NULL,
    aprovacao VARCHAR(50) NOT NULL, -- 'Aprovada', 'Rejeitada'
    placar_sim INTEGER DEFAULT 0,
    placar_nao INTEGER DEFAULT 0,
    placar_abstencao INTEGER DEFAULT 0,
    placar_outros INTEGER DEFAULT 0,
    id_proposicao_principal BIGINT,
    tipo_proposicao VARCHAR(20),
    numero_proposicao VARCHAR(20),
    ano_proposicao INTEGER,
    relevancia VARCHAR(10) DEFAULT 'baixa', -- 'alta', 'média', 'baixa'
    payload JSONB NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Tabela de votos individuais dos deputados
CREATE TABLE IF NOT EXISTS votos_deputados (
    id BIGSERIAL PRIMARY KEY,
    id_votacao BIGINT NOT NULL REFERENCES votacoes(id) ON DELETE CASCADE,
    id_deputado INTEGER NOT NULL,
    voto VARCHAR(20) NOT NULL, -- 'Sim', 'Não', 'Abstenção', 'Obstrução', 'Art17'
    justificativa TEXT,
    payload JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(id_votacao, id_deputado)
);

-- Tabela de orientações partidárias
CREATE TABLE IF NOT EXISTS orientacoes_partidos (
    id BIGSERIAL PRIMARY KEY,
    id_votacao BIGINT NOT NULL REFERENCES votacoes(id) ON DELETE CASCADE,
    partido VARCHAR(50) NOT NULL,
    orientacao VARCHAR(50) NOT NULL, -- 'Sim', 'Não', 'Liberado', 'Obstrução'
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(id_votacao, partido)
);

-- Índices para otimização de consultas
CREATE INDEX IF NOT EXISTS idx_votacoes_data ON votacoes(data_votacao DESC);
CREATE INDEX IF NOT EXISTS idx_votacoes_aprovacao ON votacoes(aprovacao);
CREATE INDEX IF NOT EXISTS idx_votacoes_proposicao ON votacoes(id_proposicao_principal);
CREATE INDEX IF NOT EXISTS idx_votacoes_relevancia ON votacoes(relevancia);
CREATE INDEX IF NOT EXISTS idx_votacoes_tipo ON votacoes(tipo_proposicao);
CREATE INDEX IF NOT EXISTS idx_votacoes_ano ON votacoes(ano_proposicao DESC);

CREATE INDEX IF NOT EXISTS idx_votos_deputados_deputado ON votos_deputados(id_deputado);
CREATE INDEX IF NOT EXISTS idx_votos_deputados_voto ON votos_deputados(voto);
CREATE INDEX IF NOT EXISTS idx_votos_deputados_votacao ON votos_deputados(id_votacao);

CREATE INDEX IF NOT EXISTS idx_orientacoes_partido ON orientacoes_partidos(partido);
CREATE INDEX IF NOT EXISTS idx_orientacoes_votacao ON orientacoes_partidos(id_votacao);

-- Trigger para atualizar updated_at automaticamente
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

DROP TRIGGER IF EXISTS update_votacoes_updated_at ON votacoes;
CREATE TRIGGER update_votacoes_updated_at 
    BEFORE UPDATE ON votacoes 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Constraints para validação de dados
DO $$
BEGIN
    -- Adicionar constraint check_placar_positive se não existir
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint 
        WHERE conname = 'check_placar_positive' AND conrelid = 'votacoes'::regclass
    ) THEN
        ALTER TABLE votacoes ADD CONSTRAINT check_placar_positive 
            CHECK (placar_sim >= 0 AND placar_nao >= 0 AND placar_abstencao >= 0 AND placar_outros >= 0);
    END IF;

    -- Adicionar constraint check_ano_valid se não existir
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint 
        WHERE conname = 'check_ano_valid' AND conrelid = 'votacoes'::regclass
    ) THEN
        ALTER TABLE votacoes ADD CONSTRAINT check_ano_valid 
            CHECK (ano_proposicao IS NULL OR (ano_proposicao >= 2000 AND ano_proposicao <= EXTRACT(YEAR FROM CURRENT_DATE) + 1));
    END IF;

    -- Adicionar constraint check_relevancia_valid se não existir
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint 
        WHERE conname = 'check_relevancia_valid' AND conrelid = 'votacoes'::regclass
    ) THEN
        ALTER TABLE votacoes ADD CONSTRAINT check_relevancia_valid 
            CHECK (relevancia IN ('alta', 'média', 'baixa'));
    END IF;
END $$;

-- Comentários para documentação
COMMENT ON TABLE votacoes IS 'Votações da Câmara dos Deputados com dados de transparência';
COMMENT ON COLUMN votacoes.id_votacao_camara IS 'ID da votação na API da Câmara dos Deputados';
COMMENT ON COLUMN votacoes.payload IS 'Dados completos da votação em formato JSON original';
COMMENT ON COLUMN votacoes.relevancia IS 'Classificação de relevância: alta, média ou baixa';

COMMENT ON TABLE votos_deputados IS 'Votos individuais dos deputados em cada votação';
COMMENT ON COLUMN votos_deputados.voto IS 'Voto do deputado: Sim, Não, Abstenção, Obstrução, Art17';

COMMENT ON TABLE orientacoes_partidos IS 'Orientações oficiais dos partidos para cada votação';
COMMENT ON COLUMN orientacoes_partidos.orientacao IS 'Orientação do partido: Sim, Não, Liberado, Obstrução';
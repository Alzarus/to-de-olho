-- Migration 012: create partidos table
CREATE TABLE IF NOT EXISTS partidos (
  id BIGINT PRIMARY KEY,
  sigla VARCHAR(20),
  nome VARCHAR(255),
  uri TEXT,
  payload JSONB,
  created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
  updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_partidos_sigla ON partidos(sigla);
CREATE INDEX IF NOT EXISTS idx_partidos_nome ON partidos(nome);

-- Migration: Create deputados_cache table
-- Version: 001
-- Description: Initial table for caching deputy data from API

CREATE TABLE IF NOT EXISTS deputados_cache (
    id INT PRIMARY KEY,
    payload JSONB NOT NULL,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create index for faster queries
CREATE INDEX IF NOT EXISTS idx_deputados_cache_updated_at ON deputados_cache(updated_at);
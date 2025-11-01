-- Migration 013: Add current_operation to backfill_executions
ALTER TABLE backfill_executions
ADD COLUMN IF NOT EXISTS current_operation TEXT;

-- Optional index for quick queries on current operation
CREATE INDEX IF NOT EXISTS idx_backfill_current_operation ON backfill_executions(current_operation);
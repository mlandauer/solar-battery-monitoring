ALTER TABLE measurements DROP COLUMN IF EXISTS regulator_state;
ALTER TABLE measurements ADD COLUMN regulator_state DOUBLE PRECISION NULL;

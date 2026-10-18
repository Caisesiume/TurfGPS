-- 0002_planted.sql — PLANTED FAILURE: a migration that does not apply (#153).
--
-- The column it constrains does not exist. The runner must report the apply
-- as failed, report 0002's verification as not run, and stop.

BEGIN;

ALTER TABLE public.planted_point ADD CONSTRAINT planted_never CHECK (no_such_column > 0);

COMMIT;

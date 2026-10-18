-- 0001_planted.down.sql — the rollback of 0001_planted.sql (#153).
--
-- It must exit 0 on both of the runner's runs, dropping the table and its key
-- on the first and nothing on the second: the observation the runner reports.

BEGIN;

DROP TABLE IF EXISTS public.planted_point;

COMMIT;

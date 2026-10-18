-- 0003_planted.sql — after a failed apply, never applied (#153).
--
-- It would apply on its own. Because 0002 fails first, the runner must report
-- it as not run rather than leave it out of the result.

SELECT 1;

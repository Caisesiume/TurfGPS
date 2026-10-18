-- 0001_planted.verify.sql — one part that passes, one that fails, and one that
-- cannot reach its condition (#153).
--
-- The runner must report A pass, B fail and C unreached, and exit 1. The job
-- that runs it fails if it reports anything else.

SET search_path = pg_catalog, public, pg_temp;

-- === Part A — passes: the table is there ====================================

DO $$
BEGIN
    IF to_regclass('public.planted_point') IS NULL THEN
        RAISE EXCEPTION 'public.planted_point is absent.';
    END IF;
END
$$;

-- === Part B — PLANTED FAILURE: a crossed axis ===============================
--
-- The known distance of ../pass/0001_fixture.verify.sql part B, with each
-- point's latitude and longitude written into each other's column. The pair
-- then lies one degree apart along a meridian, 110574.4 m, not along the
-- equator, so the assertion must raise.

CREATE TEMP TABLE planted_guard (LIKE public.planted_point INCLUDING GENERATED);

INSERT INTO pg_temp.planted_guard (id, latitude, longitude)
VALUES (1, 0, 0),
       (2, 1, 0);

DO $$
DECLARE
    measured  double precision;
    expected  constant double precision := 6378137 * pi() / 180;
    tolerance constant double precision := 0.001;
BEGIN
    SELECT ST_Distance(a.geom, b.geom) INTO measured
      FROM pg_temp.planted_guard a, pg_temp.planted_guard b
     WHERE a.id = 1 AND b.id = 2;

    IF measured IS NULL OR abs(measured - expected) > tolerance THEN
        RAISE EXCEPTION 'the equator pair measured % m, expected % m within % m.',
            coalesce(measured::text, 'nothing'), expected, tolerance;
    END IF;
END
$$;

-- === Part C — unreached: nothing to check on an empty table =================

DO $$
DECLARE
    loaded bigint;
BEGIN
    SELECT count(*) INTO loaded FROM public.planted_point;

    IF loaded = 0 THEN
        RAISE NOTICE 'part C: planted_point is empty, so nothing was checked. This is not a pass.';
        RETURN;
    END IF;
END
$$;

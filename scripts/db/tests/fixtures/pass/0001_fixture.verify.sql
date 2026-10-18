-- 0001_fixture.verify.sql — the assertions over 0001_fixture.sql (#153).
--
-- A fixture: see 0001_fixture.sql. Every part here passes against a correct
-- apply; the planted fixture beside this directory is the one that must not.
-- Each part runs in its own session after the preamble, so a part may not
-- depend on anything an earlier part left in its session.

SET search_path = pg_catalog, public, pg_temp;

-- === Part A — the column set, exactly =======================================

DO $$
DECLARE
    got text;
BEGIN
    SELECT string_agg(column_name::text, ',' ORDER BY column_name::text) INTO got
      FROM information_schema.columns
     WHERE table_schema = 'public' AND table_name = 'fixture_point';

    IF got IS DISTINCT FROM 'geom,id,latitude,longitude' THEN
        RAISE EXCEPTION 'fixture_point has columns [%], expected [geom,id,latitude,longitude].',
            coalesce(got, 'none — the table is absent');
    END IF;
END
$$;

-- === Part B — a known distance through the generated column =================
--
-- Two points on the equator one degree of longitude apart. The equator is a
-- geodesic of the WGS 84 ellipsoid, so their distance is the semi-major axis
-- times the angle: 6378137 m x pi/180 = 111319.4908 m. A crossed axis puts
-- them one degree apart along a meridian instead, 110574.4 m, which the
-- millimetre tolerance refuses. The clone carries the real table's generation
-- expression, so what is measured is the DDL that was applied.

CREATE TEMP TABLE fixture_guard (LIKE public.fixture_point INCLUDING GENERATED);

INSERT INTO pg_temp.fixture_guard (id, latitude, longitude)
VALUES (1, 0, 0),
       (2, 1, 0);

DO $$
DECLARE
    measured  double precision;
    expected  constant double precision := 6378137 * pi() / 180;
    tolerance constant double precision := 0.001;
BEGIN
    SELECT ST_Distance(a.geom, b.geom) INTO measured
      FROM pg_temp.fixture_guard a, pg_temp.fixture_guard b
     WHERE a.id = 1 AND b.id = 2;

    IF measured IS NULL OR abs(measured - expected) > tolerance THEN
        RAISE EXCEPTION 'the equator pair measured % m, expected % m within % m.',
            coalesce(measured::text, 'nothing'), expected, tolerance;
    END IF;
END
$$;

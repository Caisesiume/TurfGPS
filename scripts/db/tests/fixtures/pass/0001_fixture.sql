-- 0001_fixture.sql — a throwaway migration that scripts/db/migrate-verify.sh is
-- proven against (#153).
--
-- NOT A MIGRATION OF THIS PROJECT. It lives under scripts/db/tests/fixtures/ and
-- never under migrations/, and nothing applies it but the `migrations` job in
-- .github/workflows/gates.yml, against that job's throwaway PostGIS. Its table
-- borrows the zone store's shape — a geography point generated from latitude
-- and longitude — so the instance is exercised the way a real migration will
-- exercise it. It proves nothing about any real migration.

BEGIN;

CREATE EXTENSION IF NOT EXISTS postgis SCHEMA public;

CREATE TABLE IF NOT EXISTS public.fixture_point (
    id        integer          PRIMARY KEY,
    latitude  double precision NOT NULL,
    longitude double precision NOT NULL,
    geom      geography(Point, 4326)
              GENERATED ALWAYS AS (ST_SetSRID(ST_MakePoint(longitude, latitude), 4326)::geography) STORED
);

COMMIT;

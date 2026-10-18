-- 0001_planted.sql — the migration under the planted fixture (#153).
--
-- NOT A MIGRATION OF THIS PROJECT; see ../pass/0001_fixture.sql. This one
-- applies cleanly. The plants are in its verify file and in 0002, and the
-- `migrations` job in .github/workflows/gates.yml checks that the runner
-- reports each of them for what it is rather than as a pass.

BEGIN;

CREATE EXTENSION IF NOT EXISTS postgis SCHEMA public;

CREATE TABLE IF NOT EXISTS public.planted_point (
    id        integer          PRIMARY KEY,
    latitude  double precision NOT NULL,
    longitude double precision NOT NULL,
    geom      geography(Point, 4326)
              GENERATED ALWAYS AS (ST_SetSRID(ST_MakePoint(longitude, latitude), 4326)::geography) STORED
);

COMMIT;

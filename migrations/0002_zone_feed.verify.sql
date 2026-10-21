-- 0002_zone_feed.verify.sql — the assertions that make 0002 falsifiable
--
-- Run against a migrated copy, never against production:
--
--     psql -v ON_ERROR_STOP=1 -f migrations/0002_zone_feed.verify.sql
--
-- Every assertion raises with a message naming what it found; a silent run is a
-- pass. It writes nothing that survives: each probe inserts inside a
-- subtransaction it then discards, which consumes a few values from the
-- refusal table's id sequence and leaves no row behind.
--
-- EXECUTED ON THROWAWAY INSTANCES ONLY, part by part, by CI's `migrations` job
-- (`scripts/db/migrate-verify.sh`, #153), after 0001 and 0002 are applied.
--
-- Part A  the two column sets, in both directions
-- Part B  the latest feed time is one row
-- Part C  a refusal is recordable from what an item carried, and only in the vocabulary

-- === Part A — the two column sets, in both directions ========================
--
-- Both directions, for the reason part A of `0001_zone_store.verify.sql`
-- gives: a column ADDED is the event worth catching. Here the column to fear is
-- a raw-item or ownership column on the refusal record, which would carry a
-- feed item's `dateLastTaken` into the store `FR-026` keeps it out of.

DO $$
DECLARE
    want    record;
    actual  text[];
    extra   text[];
    missing text[];
BEGIN
    FOR want IN
        SELECT * FROM (VALUES
            ('zone_feed_state'::text,   ARRAY['latest_item_time', 'singleton']::text[]),
            ('zone_feed_refusal',       ARRAY['id', 'item_time', 'reason', 'recorded_at', 'subject', 'zone_id'])
        ) AS w(tbl, cols)
    LOOP
        SELECT array_agg(column_name::text ORDER BY column_name)
          INTO actual
          FROM information_schema.columns
         WHERE table_schema = 'public'
           AND table_name   = want.tbl;

        IF actual IS NULL THEN
            RAISE EXCEPTION 'table public.% does not exist — 0002 has not been applied here', want.tbl;
        END IF;

        SELECT array_agg(c ORDER BY c) INTO extra
          FROM unnest(actual) AS c WHERE c <> ALL (want.cols);

        SELECT array_agg(c ORDER BY c) INTO missing
          FROM unnest(want.cols) AS c WHERE c <> ALL (actual);

        IF extra IS NOT NULL THEN
            RAISE EXCEPTION
                '% has gained the column(s) %, which 0002 does not admit. A raw item or an ownership field here carries what FR-026 keeps out of the store.',
                want.tbl, extra;
        END IF;

        IF missing IS NOT NULL THEN
            RAISE EXCEPTION '% is missing the column(s) %, which the feed worker writes.', want.tbl, missing;
        END IF;
    END LOOP;
END
$$;

-- === Part B — the latest feed time is one row ================================
--
-- A second row would leave the next request's lower bound to whichever row a
-- reader happened to take. The primary key refuses a second `true`, and the
-- CHECK refuses the `false` that would otherwise slip past it.

DO $$
BEGIN
    BEGIN
        INSERT INTO public.zone_feed_state (latest_item_time) VALUES (now());

        BEGIN
            INSERT INTO public.zone_feed_state (latest_item_time) VALUES (now());
            RAISE EXCEPTION USING ERRCODE = 'ZZ011', MESSAGE = 'second_row_accepted';
        EXCEPTION
            WHEN unique_violation THEN
                NULL;   -- expected
            WHEN sqlstate 'ZZ011' THEN
                RAISE EXCEPTION 'zone_feed_state accepted a second row, so the latest feed time is no longer one value.';
        END;

        BEGIN
            INSERT INTO public.zone_feed_state (singleton, latest_item_time) VALUES (false, now());
            RAISE EXCEPTION USING ERRCODE = 'ZZ012', MESSAGE = 'false_row_accepted';
        EXCEPTION
            WHEN check_violation THEN
                NULL;   -- expected
            WHEN sqlstate 'ZZ012' THEN
                RAISE EXCEPTION 'zone_feed_state accepted a row keyed false beside the one keyed true, so the latest feed time is no longer one value.';
        END;

        RAISE EXCEPTION USING ERRCODE = 'ZZ013', MESSAGE = 'probe_rollback';
    EXCEPTION
        WHEN sqlstate 'ZZ013' THEN
            NULL;   -- intended; the probe's row is discarded
    END;
END
$$;

-- === Part C — a refusal is recordable from what an item carried ==============
--
-- `FR-122` records an item's zone id "where present", so a refusal carrying
-- neither an id nor a feed time must be recordable: an item missing both is the
-- one most in need of a record. And the reason is held to its vocabulary, as
-- `sync_run.outcome` is, because a value nobody anticipated is a refusal
-- nobody counts.

DO $$
DECLARE
    v text;
BEGIN
    FOREACH v IN ARRAY ARRAY['missing_field', 'unreadable_field', 'assertion']
    LOOP
        BEGIN
            INSERT INTO public.zone_feed_refusal (recorded_at, reason, subject) VALUES (now(), v, 'zone.id');
            RAISE EXCEPTION USING ERRCODE = 'ZZ014', MESSAGE = 'probe_rollback';
        EXCEPTION
            WHEN sqlstate 'ZZ014' THEN
                NULL;   -- intended; the subtransaction is discarded
            WHEN others THEN
                RAISE EXCEPTION
                    'zone_feed_refusal cannot record an item that carried no zone id and no feed time, with reason=%. Underlying error: %',
                    v, SQLERRM;
        END;
    END LOOP;

    BEGIN
        INSERT INTO public.zone_feed_refusal (recorded_at, reason, subject) VALUES (now(), 'not_a_known_reason', 'zone.id');
        RAISE EXCEPTION USING ERRCODE = 'ZZ015', MESSAGE = 'vocabulary_not_enforced';
    EXCEPTION
        WHEN check_violation THEN
            NULL;   -- expected
        WHEN sqlstate 'ZZ015' THEN
            RAISE EXCEPTION 'zone_feed_refusal accepted a reason outside its vocabulary.';
    END;
END
$$;

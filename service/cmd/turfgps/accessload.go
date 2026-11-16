package main

import (
	"context"
	"errors"
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/Caisesiume/TurfGPS/service/internal/accessload"
	"github.com/Caisesiume/TurfGPS/service/internal/config"
	"github.com/Caisesiume/TurfGPS/service/internal/zonestore"
)

// loadAccessMarkingsCommand is the argument that runs the executable as the
// access-marking loader (#255, `Architecture.md § Loading from the extract the
// tiles are built from`) rather than as the service:
//
//	TURFGPS_DATABASE_URL=postgres://… turfgps load-access-markings -extract /data/extract.osm.pbf
//
// The file must be the one the routing tiles are built from, in the same act;
// the planner plans nothing while the two record different files. It writes
// the feature tables and the load record in one transaction, so a failed load
// leaves the previous one current, or none. It is a deliberate operator's act,
// never run by the service and never on a timer, as `DEPLOYMENT.md` describes
// the tile build; it applies no migration, and needs
// `migrations/0003_access_markings.sql` applied first.
//
// IT IS A SUBCOMMAND, NOT A SECOND EXECUTABLE, so that starting the service on
// a clean host still takes exactly one file (`NFR-003` AC1,
// `build_test.go`). It runs before anything the service starts, and returns.
const loadAccessMarkingsCommand = "load-access-markings"

// loadAccessMarkings runs the loader. It logs the extract's SHA-256 and the
// counts it wrote, and nothing about any feature.
func loadAccessMarkings(args []string, lookup config.LookupFunc) error {
	flags := flag.NewFlagSet(loadAccessMarkingsCommand, flag.ContinueOnError)
	extract := flags.String("extract", "", "the .osm.pbf file the routing tiles are built from")

	if err := flags.Parse(args); err != nil {
		return err
	}

	if *extract == "" || flags.NArg() != 0 {
		return errors.New("usage: turfgps " + loadAccessMarkingsCommand + " -extract FILE")
	}

	databaseURL, ok := lookup(config.EnvDatabaseURL)
	if !ok || databaseURL == "" {
		return errors.New(config.EnvDatabaseURL + " is not set, so there is no store to load into")
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	ex, err := accessload.ReadFile(*extract)
	if err != nil {
		return err
	}

	pool, err := zonestore.Open(ctx, databaseURL)
	if err != nil {
		return err
	}
	defer pool.Close()

	rec, err := accessload.Load(ctx, pool, ex)
	if err != nil {
		return err
	}

	slog.Info("the access markings are loaded and current",
		"extract_sha256", rec.SHA256, "load", rec.ID,
		"ways", rec.WayRows, "barriers", rec.BarrierRows, "areas", rec.AreaRows,
		"barrier_ways", rec.BarrierWayRows, "way_exits", rec.WayExitRows)

	return nil
}

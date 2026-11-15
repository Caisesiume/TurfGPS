package main

import (
	"strings"
	"testing"

	"github.com/Caisesiume/TurfGPS/service/internal/config"
)

// TestTheLoaderRefusesWithoutAnExtractOrAStore: the loader subcommand loads
// nothing without both an extract named and a store to load into, and says
// which is missing.
func TestTheLoaderRefusesWithoutAnExtractOrAStore(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		args []string
		env  map[string]string
		want string
	}{
		{args: nil, env: map[string]string{config.EnvDatabaseURL: "postgres://127.0.0.1:1/x"}, want: "usage"},
		{args: []string{"-extract", "x.osm.pbf", "extra"}, env: map[string]string{config.EnvDatabaseURL: "postgres://127.0.0.1:1/x"}, want: "usage"},
		{args: []string{"-extract", "x.osm.pbf"}, env: nil, want: config.EnvDatabaseURL},
		{args: []string{"-extract", "no-such-file.osm.pbf"}, env: map[string]string{config.EnvDatabaseURL: "postgres://127.0.0.1:1/x"}, want: "no-such-file"},
	} {
		err := loadAccessMarkings(tc.args, lookupOnly(tc.env))
		if err == nil || !strings.Contains(err.Error(), tc.want) {
			t.Errorf("loadAccessMarkings(%v) gave %v, want a refusal naming %q", tc.args, err, tc.want)
		}
	}
}

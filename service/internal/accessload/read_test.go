package accessload

import (
	"testing"

	"github.com/Caisesiume/TurfGPS/service/internal/osmpbf"
)

// extractPath is the committed extract the routing tiles are built from
// (`internal/valhalla/testdata/README.md`).
const extractPath = "../valhalla/testdata/goteborg-inom-vallgraven.osm.pbf"

func value(v Values, key string) (string, bool) {
	for i, k := range Keys {
		if k == key && v[i] != nil {
			return *v[i], true
		}
	}

	return "", false
}

// TestTheReaderTakesTheTwelveKeysVerbatim reads the committed extract and finds
// the features #255's criteria name, each with exactly the values the extract
// holds, and no `:conditional` form read as a key (the Owner's third tag
// decision on #258).
func TestTheReaderTakesTheTwelveKeysVerbatim(t *testing.T) {
	t.Parallel()

	ex, err := ReadFile(extractPath)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}

	if ex.WayFeatures != int64(len(ex.Ways)) || ex.BarrierFeatures != int64(len(ex.Barriers)) {
		t.Fatalf("read %d way features into %d rows and %d barrier features into %d rows, want one row each", ex.WayFeatures, len(ex.Ways), ex.BarrierFeatures, len(ex.Barriers))
	}

	ways := map[int64]Way{}
	for _, w := range ex.Ways {
		ways[w.ID] = w
	}

	barriers := map[int64]Barrier{}
	for _, b := range ex.Barriers {
		barriers[b.ID] = b
	}

	for _, tc := range []struct {
		way  int64
		tags map[string]string
	}{
		{25683224, map[string]string{"access": "military"}},
		{5179080, map[string]string{"motor_vehicle": "private", "access": "yes"}},
		{107135272, map[string]string{"access": "no", "motor_vehicle": "yes"}},
		{32709123, map[string]string{"vehicle:backward": "no"}},
	} {
		w, ok := ways[tc.way]
		if !ok {
			t.Errorf("way %d is not read, want it among the ways carrying an access key", tc.way)

			continue
		}

		for k, want := range tc.tags {
			if got, ok := value(w.Values, k); !ok || got != want {
				t.Errorf("way %d's %s is %q (%v), want %q", tc.way, k, got, ok, want)
			}
		}

		if len(w.Lons) != len(w.Nodes) || len(w.Lats) != len(w.Nodes) || len(w.Nodes) < 2 {
			t.Errorf("way %d has %d nodes and %d/%d coordinates", tc.way, len(w.Nodes), len(w.Lons), len(w.Lats))
		}
	}

	gate, ok := barriers[8691377189]
	if !ok || gate.Barrier == nil || *gate.Barrier != "lift_gate" {
		t.Fatalf("node 8691377189 is read as %+v, want the lift_gate", gate)
	}

	if got, _ := value(gate.Values, "access"); got != "residents" {
		t.Errorf("the lift_gate's access is %q, want residents", got)
	}

	// A node carrying the keys without a `barrier` key is a barrier too (the
	// Owner's first tag decision on #258).
	if entrance, ok := barriers[1705784512]; !ok || entrance.Barrier != nil {
		t.Errorf("node 1705784512 (access=private, entrance=home) is read as %+v, want a barrier with no barrier key", entrance)
	}

	// Östra Larmgatan carries only motor_vehicle:conditional.
	if w, ok := ways[175304439]; ok {
		t.Errorf("way 175304439 carries only a :conditional form and was read as %+v, want it not taken", w)
	}

	var paired bool

	for _, p := range ex.BarrierWays {
		if p.Node == 8691377189 && p.Way == 444469850 {
			paired = true
		}
	}

	if !paired {
		t.Error("the lift_gate is not paired with way 444469850, which references it")
	}

	if len(ex.SHA256) != 64 {
		t.Errorf("the SHA-256 is %q", ex.SHA256)
	}
}

// TestAConditionalFormIsNoKey holds the Owner's third tag decision on #258
// (`RD-064`) at the reader: a `:conditional` form of an access key, plain or
// directional, is read as none of the twelve keys. A `:backward:conditional`
// permission is therefore not a more specific directional form, and lifts
// nothing a `:backward` key marks (N16 on #259); a conditional permission
// lifts no plain marking either (`FR-125`'s fourteenth criterion).
func TestAConditionalFormIsNoKey(t *testing.T) {
	t.Parallel()

	const when = "yes @ (Mo-Fr 08:00-18:00)"

	for _, tc := range []struct {
		name string
		tags map[string]string
		want map[string]string
	}{
		{
			"a :backward:conditional permission beside a :backward marking",
			map[string]string{"motor_vehicle:backward": "no", "motorcar:backward:conditional": when},
			map[string]string{"motor_vehicle:backward": "no"},
		},
		{
			"a :forward:conditional permission beside a :forward marking",
			map[string]string{"access:forward": "private", "motorcar:forward:conditional": when, "motor_vehicle:forward:conditional": when},
			map[string]string{"access:forward": "private"},
		},
		{
			"a plain conditional permission beside a plain marking",
			map[string]string{"access": "private", "motor_vehicle:conditional": when},
			map[string]string{"access": "private"},
		},
	} {
		got, carrying := valuesOf(tc.tags)
		if !carrying {
			t.Errorf("%s: %v is read as carrying no access key", tc.name, tc.tags)

			continue
		}

		for i, k := range Keys {
			want, named := tc.want[k]

			switch {
			case named && (got[i] == nil || *got[i] != want):
				t.Errorf("%s: %s is read as %v, want %q", tc.name, k, deref(got[i]), want)
			case !named && got[i] != nil:
				t.Errorf("%s: %s is read as %q from %v, want it absent: a :conditional form is not that key", tc.name, k, *got[i], tc.tags)
			}
		}
	}

	if _, carrying := valuesOf(map[string]string{"motorcar:backward:conditional": when}); carrying {
		t.Error("a feature carrying only motorcar:backward:conditional is read as carrying an access key")
	}
}

func deref(s *string) any {
	if s == nil {
		return nil
	}

	return *s
}

// TestAnUnsortedFileIsRefused: both passes rely on every node preceding every
// way, so a file in another order is refused rather than read with barriers
// missing from their ways.
func TestAnUnsortedFileIsRefused(t *testing.T) {
	t.Parallel()

	for name, h := range map[string]osmpbf.Handler{
		"the ways pass":  &waysPass{nodesOf: map[int64]bool{}},
		"the nodes pass": &nodesPass{wayNodes: map[int64]bool{}, coords: map[int64][2]float64{}, barrierIDs: map[int64]bool{}},
	} {
		if err := h.Way(osmpbf.Way{ID: 1}); err != nil {
			t.Fatalf("%s: a first way: %v", name, err)
		}

		if err := h.Node(osmpbf.Node{ID: 2}); err == nil {
			t.Errorf("%s: a node after a way was accepted", name)
		}
	}
}

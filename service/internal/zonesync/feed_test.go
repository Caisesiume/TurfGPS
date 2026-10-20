package zonesync

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/Caisesiume/TurfGPS/service/internal/httpapi"
)

// Tests for the zone feed (#232): the per-item ingest decision, and the poller
// that requests the feed on its schedule. The store's half — what a batch
// leaves in PostgreSQL — is `internal/syncstore`'s feed_postgis_test.go.

// Three items as the live zone feed returned them on 20 October 2026, verbatim
// apart from the line breaks: one with no `type`, one with a type, and one whose
// region carries neither `country` nor `area`. Every one carries
// `dateLastTaken`, which the all-zones response never does.
const (
	liveItemUntyped = `{"zone":{"dateCreated":"2026-10-19T21:46:07+0000","dateLastTaken":"0002-12-02T00:00:00+0000",
  "latitude":44.753365,"name":"NotSaneNow","id":831667,"totalTakeovers":0,
  "region":{"area":{"name":"Pierce County","id":4720},"country":"us","name":"Wisconsin","id":412},
  "pointsPerHour":1,"longitude":-92.801862,"takeoverPoints":185},"time":"2026-10-19T21:46:07+0000","type":"zone"}`

	liveItemTyped = `{"zone":{"dateCreated":"2026-10-18T20:00:00+0000","dateLastTaken":"0002-12-02T00:00:00+0000",
  "latitude":55.942005,"name":"StAnnesHoly","id":830395,"totalTakeovers":0,
  "region":{"area":{"name":"City of Edinburgh","id":2780},"country":"gb","name":"Scotland","id":200},
  "type":{"name":"Holy","id":9},"pointsPerHour":3,"longitude":-3.273378,"takeoverPoints":155},
  "time":"2026-10-18T20:00:00+0000","type":"zone"}`

	liveItemNoCountry = `{"zone":{"dateCreated":"2026-10-17T16:46:29+0000","dateLastTaken":"0002-12-02T00:00:00+0000",
  "latitude":53.273498,"name":"IKnowICeann","id":829796,"totalTakeovers":0,"region":{"name":"Ireland","id":266},
  "type":{"name":"Train Station","id":13},"pointsPerHour":1,"longitude":-9.045525,"takeoverPoints":185},
  "time":"2026-10-17T16:46:29+0000","type":"zone"}`
)

// feedStart is a poll's start instant after every live item above.
var feedStart = time.Date(2026, 10, 20, 0, 0, 0, 0, time.UTC)

func feedOf(items ...string) []byte { return []byte("[" + strings.Join(items, ",") + "]") }

func ptr[T any](v T) *T { return &v }

func at(raw string) time.Time {
	t, err := time.Parse("2006-01-02T15:04:05Z0700", raw)
	if err != nil {
		panic(err)
	}

	return t.UTC()
}

// TestACompleteFeedItemMapsToItsZone is AC3's mapping half: a complete live
// item becomes the zone it carries, every value where the copy keeps it. The
// expected zones are written out from the items by hand, not derived by the
// mapping under test.
func TestACompleteFeedItemMapsToItsZone(t *testing.T) {
	t.Parallel()

	batch, err := ParseZoneFeed(feedOf(liveItemUntyped, liveItemTyped, liveItemNoCountry), feedStart)
	if err != nil {
		t.Fatalf("the live feed was refused: %v", err)
	}

	want := []Zone{
		{
			ID: 831667, Name: "NotSaneNow", Latitude: 44.753365, Longitude: -92.801862,
			DateCreated: at("2026-10-19T21:46:07+0000"), TotalTakeovers: 0, TakeoverPoints: 185, PointsPerHour: 1,
			RegionID: 412, RegionName: "Wisconsin", RegionCountry: ptr("us"), AreaID: ptr[int32](4720), AreaName: ptr("Pierce County"),
		},
		{
			ID: 830395, Name: "StAnnesHoly", Latitude: 55.942005, Longitude: -3.273378,
			DateCreated: at("2026-10-18T20:00:00+0000"), TotalTakeovers: 0, TakeoverPoints: 155, PointsPerHour: 3,
			TypeID: ptr[int16](9), RegionID: 200, RegionName: "Scotland", RegionCountry: ptr("gb"),
			AreaID: ptr[int32](2780), AreaName: ptr("City of Edinburgh"),
		},
		{
			ID: 829796, Name: "IKnowICeann", Latitude: 53.273498, Longitude: -9.045525,
			DateCreated: at("2026-10-17T16:46:29+0000"), TotalTakeovers: 0, TakeoverPoints: 185, PointsPerHour: 1,
			TypeID: ptr[int16](13), RegionID: 266, RegionName: "Ireland",
		},
	}

	if !reflect.DeepEqual(batch.Zones, want) {
		t.Errorf("the live feed mapped to\n%+v\nwant\n%+v", batch.Zones, want)
	}

	if len(batch.Refusals) != 0 {
		t.Errorf("three complete live items were refused %d time(s): %+v", len(batch.Refusals), batch.Refusals)
	}

	if batch.Latest == nil || !batch.Latest.Equal(at("2026-10-19T21:46:07+0000")) {
		t.Errorf("the latest feed time is %v, want the newest item's 2026-10-19T21:46:07Z", batch.Latest)
	}
}

// TestAFeedItemTheCopyCannotStoreIsRecordedAndNeverGuessed is AC4: an item
// lacking a required field, or carrying a value a per-row assertion refuses,
// yields no zone and a refusal naming its zone id where present, its feed time
// and the field or assertion; an item lacking only optional fields is stored
// with nothing recorded against it.
func TestAFeedItemTheCopyCannotStoreIsRecordedAndNeverGuessed(t *testing.T) {
	t.Parallel()

	const itemTime = "2026-10-19T12:00:00+0000"

	item := func(zone string) string {
		return `{"type":"zone","time":"` + itemTime + `","zone":{` + zone + `}}`
	}

	complete := `"id":42,"name":"Alpha","latitude":59.3,"longitude":18.06,"dateCreated":"2024-04-28T10:00:00+0000",` +
		`"totalTakeovers":7,"takeoverPoints":80,"pointsPerHour":3`

	for _, tc := range []struct {
		name   string
		item   string
		stored bool
		want   []FeedRefusal
	}{
		{
			name:   "only optional fields absent",
			item:   item(complete + `,"region":{"id":266,"name":"Ireland"}`),
			stored: true,
		},
		{
			name: "no latitude",
			item: item(strings.Replace(complete, `"latitude":59.3,`, "", 1) + `,"region":{"id":141,"name":"Stockholm"}`),
			want: []FeedRefusal{{ZoneID: ptr[int32](42), ItemTime: ptr(at(itemTime)), Reason: FeedMissingField, Subject: "zone.latitude"}},
		},
		{
			name: "a null region id and no name",
			item: item(strings.Replace(complete, `"name":"Alpha",`, "", 1) + `,"region":{"id":null,"name":"Stockholm"}`),
			want: []FeedRefusal{
				{ZoneID: ptr[int32](42), ItemTime: ptr(at(itemTime)), Reason: FeedMissingField, Subject: "zone.name"},
				{ZoneID: ptr[int32](42), ItemTime: ptr(at(itemTime)), Reason: FeedMissingField, Subject: "zone.region.id"},
			},
		},
		{
			name: "no id at all",
			item: item(strings.Replace(complete, `"id":42,`, "", 1) + `,"region":{"id":141,"name":"Stockholm"}`),
			want: []FeedRefusal{{ItemTime: ptr(at(itemTime)), Reason: FeedMissingField, Subject: "zone.id"}},
		},
		{
			name: "no zone",
			item: `{"type":"zone","time":"` + itemTime + `"}`,
			want: []FeedRefusal{{ItemTime: ptr(at(itemTime)), Reason: FeedMissingField, Subject: "zone"}},
		},
		{
			name: "a latitude outside its range",
			item: item(strings.Replace(complete, `"latitude":59.3`, `"latitude":91`, 1) + `,"region":{"id":141,"name":"Stockholm"}`),
			want: []FeedRefusal{{ZoneID: ptr[int32](42), ItemTime: ptr(at(itemTime)), Reason: FeedAssertion, Subject: AssertLatitudeRange}},
		},
		{
			name: "a longitude outside its range",
			item: item(strings.Replace(complete, `"longitude":18.06`, `"longitude":-180.5`, 1) + `,"region":{"id":141,"name":"Stockholm"}`),
			want: []FeedRefusal{{ZoneID: ptr[int32](42), ItemTime: ptr(at(itemTime)), Reason: FeedAssertion, Subject: AssertLongitudeRange}},
		},
		{
			name: "a dateCreated later than the poll's start",
			item: item(strings.Replace(complete, `"dateCreated":"2024-04-28T10:00:00+0000"`, `"dateCreated":"2026-10-21T00:00:00+0000"`, 1) + `,"region":{"id":141,"name":"Stockholm"}`),
			want: []FeedRefusal{{ZoneID: ptr[int32](42), ItemTime: ptr(at(itemTime)), Reason: FeedAssertion, Subject: AssertCreatedNotAfterRun}},
		},
		{
			name: "negative takeovers",
			item: item(strings.Replace(complete, `"totalTakeovers":7`, `"totalTakeovers":-1`, 1) + `,"region":{"id":141,"name":"Stockholm"}`),
			want: []FeedRefusal{{ZoneID: ptr[int32](42), ItemTime: ptr(at(itemTime)), Reason: FeedAssertion, Subject: AssertTakeoversNonNeg}},
		},
		{
			name: "takeover points wider than their column",
			item: item(strings.Replace(complete, `"takeoverPoints":80`, `"takeoverPoints":40000`, 1) + `,"region":{"id":141,"name":"Stockholm"}`),
			want: []FeedRefusal{{ZoneID: ptr[int32](42), ItemTime: ptr(at(itemTime)), Reason: FeedUnreadableField, Subject: "zone.takeoverPoints"}},
		},
		{
			name: "a dateCreated in no accepted form",
			item: item(strings.Replace(complete, `"2024-04-28T10:00:00+0000"`, `"last spring"`, 1) + `,"region":{"id":141,"name":"Stockholm"}`),
			want: []FeedRefusal{{ZoneID: ptr[int32](42), ItemTime: ptr(at(itemTime)), Reason: FeedUnreadableField, Subject: "zone.dateCreated"}},
		},
		{
			name: "a name carrying a NUL byte",
			item: item(strings.Replace(complete, `"Alpha"`, `"Al\u0000pha"`, 1) + `,"region":{"id":141,"name":"Stockholm"}`),
			want: []FeedRefusal{{ZoneID: ptr[int32](42), ItemTime: ptr(at(itemTime)), Reason: FeedUnreadableField, Subject: "zone.name"}},
		},
		{
			name: "an item that is not a zone item",
			item: `{"type":"takeover","time":"` + itemTime + `","zone":{` + complete + `,"region":{"id":141,"name":"Stockholm"}}}`,
			want: []FeedRefusal{{ItemTime: ptr(at(itemTime)), Reason: FeedUnreadableField, Subject: "type"}},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			batch, err := ParseZoneFeed(feedOf(tc.item), feedStart)
			if err != nil {
				t.Fatalf("the feed holding one item was refused whole: %v — one item must be decided on its own", err)
			}

			if tc.stored {
				if len(batch.Zones) != 1 || len(batch.Refusals) != 0 {
					t.Errorf("an item lacking only optional fields gave %d zone(s) and refusals %+v, want it stored with nothing recorded",
						len(batch.Zones), batch.Refusals)
				}

				return
			}

			if len(batch.Zones) != 0 {
				t.Errorf("an item with %s was stored as %+v — a zone guessed from an item the copy cannot store", tc.name, batch.Zones)
			}

			if !reflect.DeepEqual(batch.Refusals, tc.want) {
				t.Errorf("an item with %s was recorded as\n%s\nwant\n%s", tc.name, describe(batch.Refusals), describe(tc.want))
			}

			// RD-043: the latest feed time moves past a recorded item.
			if batch.Latest == nil || !batch.Latest.Equal(at(itemTime)) {
				t.Errorf("the latest feed time after a refused item is %v, want its own %s", batch.Latest, itemTime)
			}
		})
	}
}

func describe(refusals []FeedRefusal) string {
	var b strings.Builder

	for _, r := range refusals {
		id, when := "none", "none"
		if r.ZoneID != nil {
			id = fmt.Sprint(*r.ZoneID)
		}

		if r.ItemTime != nil {
			when = r.ItemTime.Format(time.RFC3339)
		}

		fmt.Fprintf(&b, "  {zone %s, time %s, %s, %s}\n", id, when, r.Reason, r.Subject)
	}

	return b.String()
}

// TestOneBadItemDoesNotRefuseTheBatch: a feed is decided item by item, so a
// refused item leaves its neighbours stored.
func TestOneBadItemDoesNotRefuseTheBatch(t *testing.T) {
	t.Parallel()

	bad := `{"type":"zone","time":"2026-10-19T22:00:00+0000","zone":{"id":7}}`

	batch, err := ParseZoneFeed(feedOf(bad, liveItemUntyped, liveItemTyped), feedStart)
	if err != nil {
		t.Fatalf("a feed with one incomplete item was refused whole: %v", err)
	}

	if len(batch.Zones) != 2 {
		t.Errorf("a feed of one incomplete item and two complete ones stored %d zone(s), want the two", len(batch.Zones))
	}

	if len(batch.Refusals) == 0 || batch.Refusals[0].ZoneID == nil || *batch.Refusals[0].ZoneID != 7 {
		t.Errorf("the incomplete item was recorded as %+v, want refusals naming zone 7", batch.Refusals)
	}

	if batch.Latest == nil || !batch.Latest.Equal(at("2026-10-19T22:00:00+0000")) {
		t.Errorf("the latest feed time is %v, want the refused item's, which is the newest", batch.Latest)
	}
}

// TestOneIDNamedTwiceIsStoredOnceFromTheLaterItem is RD-042's last sentence,
// and it is load-bearing below this package: the merge is one INSERT … ON
// CONFLICT, which PostgreSQL refuses when two staged rows name one id.
func TestOneIDNamedTwiceIsStoredOnceFromTheLaterItem(t *testing.T) {
	t.Parallel()

	older := strings.Replace(strings.Replace(liveItemUntyped, `"NotSaneNow"`, `"OlderName"`, 1),
		`"time":"2026-10-19T21:46:07+0000"`, `"time":"2026-10-19T10:00:00+0000"`, 1)

	for _, order := range [][]string{{liveItemUntyped, older}, {older, liveItemUntyped}} {
		batch, err := ParseZoneFeed(feedOf(order...), feedStart)
		if err != nil {
			t.Fatalf("refused: %v", err)
		}

		if len(batch.Zones) != 1 || batch.Zones[0].Name != "NotSaneNow" {
			t.Errorf("one id named by two items gave %+v, want one zone, from the item with the later feed time", batch.Zones)
		}
	}
}

// TestAFeedTimeAheadOfThePollDoesNotMoveTheLowerBound: a time later than the
// poll's own start is not taken as the latest feed time, so it cannot hide
// every zone created before it from the next request.
func TestAFeedTimeAheadOfThePollDoesNotMoveTheLowerBound(t *testing.T) {
	t.Parallel()

	ahead := strings.Replace(liveItemTyped, `"time":"2026-10-18T20:00:00+0000"`, `"time":"2027-01-01T00:00:00+0000"`, 1)

	batch, err := ParseZoneFeed(feedOf(ahead, liveItemNoCountry), feedStart)
	if err != nil {
		t.Fatalf("refused: %v", err)
	}

	if batch.Latest == nil || !batch.Latest.Equal(at("2026-10-17T16:46:29+0000")) {
		t.Errorf("the latest feed time is %v, want the newest one not ahead of the poll's start", batch.Latest)
	}
}

// TestAResponseThatIsNotAFeedIsRefusedWhole: past this, items are decided one
// by one; a body that is not an array of them is decided once.
func TestAResponseThatIsNotAFeedIsRefusedWhole(t *testing.T) {
	t.Parallel()

	for _, body := range []string{``, `{}`, `[{"type":"zone"`, `"zone"`} {
		if _, err := ParseZoneFeed([]byte(body), feedStart); err == nil {
			t.Errorf("the body %q was accepted as a zone feed", body)
		}
	}

	batch, err := ParseZoneFeed([]byte(`[]`), feedStart)
	if err != nil || !batch.empty() {
		t.Errorf("an empty feed gave %+v, %v; want an empty batch and no error — a feed may carry no items", batch, err)
	}
}

// TestTheFeedMappingCarriesNoOwnership is AC5 (`FR-026`): the feed's mapping
// is the all-zones mapping, whose Zone has no owner and no last-taken field,
// so an item carrying `currentOwner` and `dateLastTaken` maps to exactly the
// zone the same item carries without them.
func TestTheFeedMappingCarriesNoOwnership(t *testing.T) {
	t.Parallel()

	zoneType := reflect.TypeFor[Zone]()

	for i := range zoneType.NumField() {
		name := strings.ToLower(zoneType.Field(i).Name)
		if strings.Contains(name, "owner") || strings.Contains(name, "taken") {
			t.Errorf("Zone has the field %s, so the ingest mapping can populate an ownership column FR-026 forbids", zoneType.Field(i).Name)
		}
	}

	owned := strings.Replace(liveItemUntyped, `"dateLastTaken":"0002-12-02T00:00:00+0000"`,
		`"dateLastTaken":"2026-10-19T22:00:00+0000","currentOwner":{"id":330825,"name":"Zinky"}`, 1)
	bare := strings.Replace(liveItemUntyped, `"dateLastTaken":"0002-12-02T00:00:00+0000",`, ``, 1)

	withOwner, err := ParseZoneFeed(feedOf(owned), feedStart)
	if err != nil {
		t.Fatalf("refused: %v", err)
	}

	without, err := ParseZoneFeed(feedOf(bare), feedStart)
	if err != nil {
		t.Fatalf("refused: %v", err)
	}

	if !reflect.DeepEqual(withOwner.Zones, without.Zones) || len(withOwner.Zones) != 1 {
		t.Errorf("an item carrying an owner and a last-taken date mapped to %+v, and the same item without them to %+v — the ownership fields reached the mapping",
			withOwner.Zones, without.Zones)
	}
}

// --- the poller --------------------------------------------------------------

// fakeFeedStore holds the latest feed time the way the database does: moved
// forward by an applied batch, never back, and read by every poller built
// over it — which is what a process restart amounts to.
type fakeFeedStore struct {
	mu sync.Mutex

	latest    *time.Time
	now       time.Time
	cursorErr error
	applyErr  error
	applied   []FeedBatch

	// held, when set, holds ApplyFeed until it is closed, after signalling
	// entered.
	held    chan struct{}
	entered chan struct{}
}

func (s *fakeFeedStore) FeedCursor(context.Context) (*time.Time, time.Time, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.cursorErr != nil {
		return nil, time.Time{}, s.cursorErr
	}

	return s.latest, s.now, nil
}

func (s *fakeFeedStore) ApplyFeed(_ context.Context, batch FeedBatch) (FeedApplied, error) {
	if s.held != nil {
		select {
		case s.entered <- struct{}{}:
		default:
		}

		<-s.held
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.applyErr != nil {
		return FeedApplied{}, s.applyErr
	}

	s.applied = append(s.applied, batch)

	if batch.Latest != nil && (s.latest == nil || batch.Latest.After(*s.latest)) {
		latest := *batch.Latest
		s.latest = &latest
	}

	return FeedApplied{Inserted: len(batch.Zones)}, nil
}

func (s *fakeFeedStore) batches() int {
	s.mu.Lock()
	defer s.mu.Unlock()

	return len(s.applied)
}

// fakeFeedFetch answers every request with one body and records the lower
// bound each one carried.
type fakeFeedFetch struct {
	mu     sync.Mutex
	body   []byte
	err    error
	afters []*time.Time
}

func (f *fakeFeedFetch) fetch(_ context.Context, after *time.Time) ([]byte, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.afters = append(f.afters, after)

	return f.body, f.err
}

func (f *fakeFeedFetch) requests() []*time.Time {
	f.mu.Lock()
	defer f.mu.Unlock()

	return append([]*time.Time(nil), f.afters...)
}

// feedTestInterval is this file's own and stands for no schedule: the feed's
// interval has its home in the Owner's decision on #20 and reaches the service
// from configuration only.
const feedTestInterval = 7 * time.Minute

func newTestFeedPoller(t *testing.T, store FeedStore, fetch FeedFetchFunc) *FeedPoller {
	t.Helper()

	p, err := NewFeedPoller(FeedConfig{
		Fetch: fetch, Store: store, Interval: feedTestInterval, FetchTimeout: time.Second, DatabaseTimeout: time.Second,
	})
	if err != nil {
		t.Fatalf("building the feed poller: %v", err)
	}

	return p
}

// firings drives p.Run through n firings of its schedule without a clock, and
// returns the waits it asked for between them.
func firings(t *testing.T, p *FeedPoller, n int) []time.Duration {
	t.Helper()

	var (
		clock = feedStart
		waits []time.Duration
	)

	p.now = func() time.Time { return clock }
	p.sleep = func(_ context.Context, d time.Duration) bool {
		waits = append(waits, d)
		clock = clock.Add(d)

		return len(waits) < n
	}

	if err := p.Run(t.Context()); err != nil {
		t.Fatalf("Run returned %v, want nil", err)
	}

	return waits
}

// TestEachFiringIssuesOneFeedRequest is AC1's first half (`FR-119`): when the
// schedule fires, the job issues one zone-feed request — not none, not two —
// and the next firing is an interval later.
func TestEachFiringIssuesOneFeedRequest(t *testing.T) {
	t.Parallel()

	for _, n := range []int{1, 3} {
		store := &fakeFeedStore{now: feedStart}
		fetch := &fakeFeedFetch{body: feedOf(liveItemUntyped)}

		waits := firings(t, newTestFeedPoller(t, store, fetch.fetch), n)

		if got := len(fetch.requests()); got != n {
			t.Errorf("%d firing(s) of the feed schedule issued %d zone-feed request(s), want exactly one per firing", n, got)
		}

		for i, w := range waits {
			if w != feedTestInterval {
				t.Errorf("wait %d before the next firing was %s, want the configured interval %s", i, w, feedTestInterval)
			}
		}
	}
}

// TestTheLowerBoundIsTheLatestFeedTimeRecorded is AC2 (`FR-120`): a request
// carries no lower bound while no feed time is recorded, and the recorded one
// once there is — including the first request of a new process, which reads it
// from the store and never from the process that recorded it.
func TestTheLowerBoundIsTheLatestFeedTimeRecorded(t *testing.T) {
	t.Parallel()

	store := &fakeFeedStore{now: feedStart}
	fetch := &fakeFeedFetch{body: feedOf(liveItemUntyped, liveItemTyped)}

	firings(t, newTestFeedPoller(t, store, fetch.fetch), 2)

	// A second process over the same store: nothing carries over but the store.
	restarted := &fakeFeedFetch{body: feedOf()}
	firings(t, newTestFeedPoller(t, store, restarted.fetch), 1)

	newest := at("2026-10-19T21:46:07+0000")

	got := append(fetch.requests(), restarted.requests()...)
	if len(got) != 3 {
		t.Fatalf("three firings issued %d requests", len(got))
	}

	if got[0] != nil {
		t.Errorf("the first request, with no feed time recorded, carried the lower bound %s, want none", got[0])
	}

	for i, after := range got[1:] {
		if after == nil || !after.Equal(newest) {
			t.Errorf("request %d carried the lower bound %v, want the latest feed time recorded, %s", i+2, after, newest)
		}
	}
}

// TestAFailedPollRecordsNothingAndDoesNotStop: a poll that cannot read the
// latest feed time issues no request; one whose batch is not written leaves
// the time where it was; and neither stops the loop.
func TestAFailedPollRecordsNothingAndDoesNotStop(t *testing.T) {
	t.Parallel()

	unread := &fakeFeedStore{now: feedStart, cursorErr: errStore}
	fetch := &fakeFeedFetch{body: feedOf(liveItemUntyped)}

	if waits := firings(t, newTestFeedPoller(t, unread, fetch.fetch), 2); len(waits) != 2 {
		t.Errorf("the loop stopped after %d wait(s) on a store it could not read", len(waits))
	}

	if n := len(fetch.requests()); n != 0 {
		t.Errorf("%d zone-feed request(s) were issued with no latest feed time read, want none", n)
	}

	for _, applyErr := range []error{errStore, ErrSyncLockHeld} {
		unwritten := &fakeFeedStore{now: feedStart, applyErr: applyErr}
		firings(t, newTestFeedPoller(t, unwritten, (&fakeFeedFetch{body: feedOf(liveItemUntyped)}).fetch), 2)

		if unwritten.latest != nil {
			t.Errorf("a batch refused with %v moved the latest feed time to %s", applyErr, unwritten.latest)
		}
	}
}

// TestARequestIsAnsweredWhileAFeedPollIsHeldInItsWrite is AC1's second half
// (`FR-119`): a request served while a feed poll is in flight issues no
// zone-feed request and does not wait on the poll. The import half is
// `offrequestpath_test.go`, which ranges over this package and so over the
// feed; this is the runtime half, built as
// TestARequestIsAnsweredWhileARefreshIsHeldInItsMerge is and for the reasons
// that file gives.
func TestARequestIsAnsweredWhileAFeedPollIsHeldInItsWrite(t *testing.T) {
	t.Parallel()

	store := &fakeFeedStore{now: feedStart, held: make(chan struct{}), entered: make(chan struct{}, 1)}
	release := sync.OnceFunc(func() { close(store.held) })
	t.Cleanup(release)

	fetch := &fakeFeedFetch{body: feedOf(liveItemUntyped)}
	poller := newTestFeedPoller(t, store, fetch.fetch)

	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	stopped := make(chan error, 1)

	go func() { stopped <- poller.Run(ctx) }()

	select {
	case <-store.entered:
	case <-time.After(harnessBudget):
		t.Fatalf("no feed poll reached its write within %s, so nothing was in flight", harnessBudget)
	}

	served := make(chan int, 1)

	go func() {
		rec := httptest.NewRecorder()
		httpapi.NewMux().ServeHTTP(rec, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/", nil))
		served <- rec.Code
	}()

	select {
	case code := <-served:
		if code == http.StatusNotFound {
			t.Fatalf("the request reached no handler (status %d), so nothing on the request surface was exercised", code)
		}

		if code != http.StatusOK {
			t.Errorf("the request answered %d while a feed poll was in flight, want %d", code, http.StatusOK)
		}
	case <-time.After(harnessBudget):
		t.Fatalf("the request was unanswered %s after a feed poll entered its write, which is still held: the request is waiting on the feed job", harnessBudget)
	}

	if n := len(fetch.requests()); n != 1 {
		t.Errorf("%d zone-feed request(s) existed when the request was answered, want exactly the poll's one", n)
	}

	if n := store.batches(); n != 0 {
		t.Errorf("the held poll wrote %d batch(es) before the request was answered, so it was not in flight", n)
	}

	release()
	cancel()

	select {
	case err := <-stopped:
		if err != nil {
			t.Errorf("the poller returned %v, want nil", err)
		}
	case <-time.After(harnessBudget):
		t.Errorf("the poller did not return within %s of cancellation", harnessBudget)
	}
}

func TestTheFeedPollerRefusesAnIncompleteConfiguration(t *testing.T) {
	t.Parallel()

	ok := FeedConfig{Fetch: (&fakeFeedFetch{}).fetch, Store: &fakeFeedStore{}, Interval: time.Minute, FetchTimeout: time.Second, DatabaseTimeout: time.Second}

	for name, mutate := range map[string]func(*FeedConfig){
		"no fetch":            func(c *FeedConfig) { c.Fetch = nil },
		"no store":            func(c *FeedConfig) { c.Store = nil },
		"no interval":         func(c *FeedConfig) { c.Interval = 0 },
		"no fetch timeout":    func(c *FeedConfig) { c.FetchTimeout = 0 },
		"no database timeout": func(c *FeedConfig) { c.DatabaseTimeout = -time.Second },
	} {
		cfg := ok
		mutate(&cfg)

		if _, err := NewFeedPoller(cfg); err == nil {
			t.Errorf("a feed poller with %s was built", name)
		}
	}

	if _, err := NewFeedPoller(ok); err != nil {
		t.Errorf("a complete configuration was refused: %v", err)
	}
}

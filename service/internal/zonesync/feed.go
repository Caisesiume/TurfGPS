package zonesync

// The zone feed (#232): new zones reach the local copy between full refreshes,
// from the zone feed the Owner's decision on #20 (`#issuecomment-6279336314`)
// adds, on that feed's own schedule. The records are `FR-119` … `FR-122`, with
// `FR-026` binding the mapping, and they stay `draft` until
// `Architecture.md § Retrieving zones` carries the feed.
//
// IT IS THE SAME WORKER PACKAGE ON PURPOSE. Everything said in this package's
// doc about staying off the request path holds for the feed because the feed
// is here: FeedPoller.Run is a loop the process owns, nothing exported runs one
// poll, and `offrequestpath_test.go` keeps every request-path package from
// reaching this one. It is a separate loop from Scheduler, with its own
// configuration, because the two answer to different schedules.
//
// IT WRITES THROUGH THE SAME MAPPING, AND ONLY NEW IDS. A feed item's zone is
// decoded by wireZone and mapped by its zone method — the all-zones mapping,
// which has no field for an owner or a last-taken date, so the feed's
// `dateLastTaken` is dropped where every other unknown key is (`FR-026`) — and
// `internal/syncstore` inserts it through the INSERT of
// `Architecture.md § The sync write path`, built from the one column binding
// the all-zones path uses. An id the copy already holds is left untouched: a
// feed item is the zone as it was created, and the copy may hold newer values,
// coordinates included, from a full sync (#238; this supersedes `RD-041`).
//
// WHAT IS NEW IS DECIDING PER ITEM. A full response is refused whole; a feed
// batch is not, because a feed returns "from 0 to x" items and can be reset
// (the live API index, read 20 October 2026), so one bad item must neither
// block the rest nor be guessed into a row. An item the copy cannot store is
// recorded instead (`FR-122`, `RD-042`), and the latest feed time advances past
// it (`RD-043`). The whole-response assertions — the row-count floor above all —
// are the all-zones path's and do not apply here.
//
// ABSENCE MEANS NOTHING HERE. A zone missing from a feed batch is a zone the
// feed did not repeat, and nothing in this file reads it as anything else.

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"
)

// FeedFetchFunc requests the zone feed. after is the lower bound — only items
// newer than it are asked for — and nil asks for none (`FR-120`).
//
// A func, as FetchFunc is, so the adapter that implements it needs no
// knowledge of this package.
type FeedFetchFunc func(ctx context.Context, after *time.Time) ([]byte, error)

// FeedRefusalReason is how an item the copy cannot store is classified, and
// what `zone_feed_refusal.reason` carries. The CHECK in
// `migrations/0002_zone_feed.sql` holds the column to these three.
type FeedRefusalReason string

const (
	// FeedMissingField: a field the copy requires is absent or null.
	FeedMissingField FeedRefusalReason = "missing_field"

	// FeedUnreadableField: a field is present in a form its column cannot
	// hold — a string where a number belongs, a number its column's width
	// cannot hold, a date in no accepted form, text carrying a NUL byte.
	FeedUnreadableField FeedRefusalReason = "unreadable_field"

	// FeedAssertion: every field is present and readable, and a value is one a
	// per-row assertion refuses.
	FeedAssertion FeedRefusalReason = "assertion"
)

// The per-row assertions a feed item is held to, by the name a refusal records.
//
// The first two and the last are the per-row assertions of `Architecture.md §
// The sync write path`; the first three are also the CHECK constraints on
// `zone` and carry their names. zone_takeovers_nonneg is there for the reason
// the others are: a value the table refuses aborts the transaction that
// carries it, which would refuse every item in the batch and hold the latest
// feed time back for as long as the feed repeats the item. Refusing the one
// item is what keeps the rest storable.
const (
	AssertLatitudeRange      = "zone_lat_range"
	AssertLongitudeRange     = "zone_lon_range"
	AssertTakeoversNonNeg    = "zone_takeovers_nonneg"
	AssertCreatedNotAfterRun = "date_created_not_after_run_start"
)

// The item-level subjects a refusal can name, beside the zone's own fields.
const (
	feedItemType              = "zone"
	feedSubjectItem           = "item"
	feedSubjectType           = "type"
	feedSubjectZone           = "zone"
	feedSubjectZonePathPrefix = "zone."
)

// FeedRefusal is one recorded reason an item was not stored. An item refused
// for two reasons is two refusals, so a record names one field or one
// assertion and is queryable by it.
type FeedRefusal struct {
	// ZoneID is the item's zone id, nil where the item carried none.
	ZoneID *int32

	// ItemTime is the item's feed time, nil where it carried none this file
	// can read.
	ItemTime *time.Time

	Reason FeedRefusalReason

	// Subject names the field, as its path in the item, or the assertion.
	Subject string
}

// FeedBatch is what one feed response leaves to be written, in one
// transaction.
type FeedBatch struct {
	// Zones are the storable zones, one per id.
	Zones []Zone

	// Refusals are the records of every item that was not stored.
	Refusals []FeedRefusal

	// Latest is the feed time the latest-processed record moves to, never
	// back: nil leaves it where it is.
	Latest *time.Time
}

// empty reports whether the batch would write nothing at all.
func (b FeedBatch) empty() bool {
	return len(b.Zones) == 0 && len(b.Refusals) == 0 && b.Latest == nil
}

// FeedApplied is what one batch's write did with its zones: Inserted were new
// to the copy, and Held were ids it already held, left as they were.
type FeedApplied struct {
	Inserted int
	Held     int
}

// ErrSyncLockHeld is what ApplyFeed returns when the sync's lock is held — by
// a full refresh in progress, or a migration — and nothing was written. The
// batch is left for the next poll, whose request asks for it again because the
// latest feed time did not move.
var ErrSyncLockHeld = errors.New("the zone sync's lock is held, so nothing from this feed batch was written")

// FeedStore is what a feed poll needs of the zone store. Declared here, by the
// consumer, as Store is.
type FeedStore interface {
	// FeedCursor returns the latest feed time recorded, nil when none is, and
	// the server's current instant. That instant is the poll's start: the
	// bound date_created is asserted against, read from the clock every other
	// instant in the store is read from.
	FeedCursor(ctx context.Context) (latest *time.Time, now time.Time, err error)

	// ApplyFeed writes the batch in ONE transaction, under the sync's lock:
	// the zones whose id the copy lacks, the refusals, and the latest
	// feed time — so a time is never recorded past items that were not
	// written. It returns ErrSyncLockHeld, having written nothing, when the
	// lock is held elsewhere.
	ApplyFeed(ctx context.Context, batch FeedBatch) (FeedApplied, error)
}

// FeedConfig is everything a FeedPoller needs. Every field but Logger is
// required, and there are no defaults here: the interval's home is the Owner's
// decision cited at the top of this file, and it reaches this package through
// `internal/config`, which gives it none either.
type FeedConfig struct {
	Fetch  FeedFetchFunc
	Store  FeedStore
	Logger *slog.Logger

	Interval        time.Duration
	FetchTimeout    time.Duration
	DatabaseTimeout time.Duration
}

// FeedPoller is the zone feed's background job (`FR-119`). Run is its only
// exported entry, and it is a loop: there is nothing here for a request to
// call, and nothing a request could wait on.
type FeedPoller struct {
	fetch FeedFetchFunc
	store FeedStore
	log   *slog.Logger

	interval        time.Duration
	fetchTimeout    time.Duration
	databaseTimeout time.Duration

	// now and sleep are fields so this package's tests can drive the loop
	// without a clock, as Scheduler's are.
	now   func() time.Time
	sleep func(ctx context.Context, d time.Duration) bool
}

// NewFeedPoller builds the job.
func NewFeedPoller(cfg FeedConfig) (*FeedPoller, error) {
	if cfg.Fetch == nil {
		return nil, errors.New("no zone-feed fetch")
	}

	if cfg.Store == nil {
		return nil, errors.New("no zone-feed store")
	}

	for _, b := range []struct {
		name string
		d    time.Duration
	}{
		{"the zone-feed interval", cfg.Interval},
		{"the zone-feed fetch timeout", cfg.FetchTimeout},
		{"the zone-feed database timeout", cfg.DatabaseTimeout},
	} {
		if b.d <= 0 {
			return nil, fmt.Errorf("%s is not positive", b.name)
		}
	}

	return &FeedPoller{
		fetch:           cfg.Fetch,
		store:           cfg.Store,
		log:             Config{Logger: cfg.Logger}.logger(),
		interval:        cfg.Interval,
		fetchTimeout:    cfg.FetchTimeout,
		databaseTimeout: cfg.DatabaseTimeout,
		now:             time.Now,
		sleep:           sleep,
	}, nil
}

// Run polls the feed on its schedule until ctx is cancelled, and then returns
// nil. The first poll is at once, so a process that starts picks up what was
// created while it was down; each later one an interval after the previous one
// began.
//
// IT NEVER RETURNS ON A FAILURE, for the reason Scheduler.Run gives: each
// failure costs one interval of the feed and nothing else, and the next poll
// asks again from the same latest feed time.
//
// THE SCHEDULE IS THIS PROCESS'S AND DOES NOT SURVIVE A RESTART, unlike the
// all-zones gate. That gate exists because the all-zones endpoint meters one
// request per its own interval; the feed answers only to the API's general
// limit, which `internal/turf`'s budget holds for every request this process
// sends. A restart costs one feed request early, and the latest feed time —
// which does survive it — keeps that request to items not yet processed.
func (p *FeedPoller) Run(ctx context.Context) error {
	for {
		if ctx.Err() != nil {
			return nil
		}

		began := p.now()

		p.poll(ctx)

		if wait := p.interval - p.now().Sub(began); wait > 0 {
			if !p.sleep(ctx, wait) {
				return nil
			}
		}
	}
}

// poll is one firing: read the latest feed time, request the feed once, and
// write what came back.
func (p *FeedPoller) poll(ctx context.Context) {
	readCtx, cancelRead := context.WithTimeout(ctx, p.databaseTimeout)
	latest, startedAt, err := p.store.FeedCursor(readCtx)

	cancelRead()

	if err != nil {
		// No request without the lower bound: one sent without it would re-read
		// what is already stored, and one sent with a guessed bound could skip
		// what is not.
		p.log.Error("the latest zone-feed time could not be read, so no feed request was issued", "error", err)

		return
	}

	fetchCtx, cancelFetch := context.WithTimeout(ctx, p.fetchTimeout)
	body, err := p.fetch(fetchCtx, latest)

	cancelFetch()

	if err != nil {
		if ctx.Err() != nil {
			p.log.Info("the zone-feed poll was cancelled during the request", "error", err)

			return
		}

		p.log.Error("the zone-feed request failed", "after", latest, "error", err)

		return
	}

	batch, err := ParseZoneFeed(body, startedAt)
	if err != nil {
		p.log.Error("the zone-feed response could not be read, so nothing from it was written", "error", err)

		return
	}

	if batch.empty() {
		p.log.Info("the zone feed carried nothing newer than the latest feed time", "after", latest)

		return
	}

	applyCtx, cancelApply := context.WithTimeout(ctx, p.databaseTimeout)
	applied, err := p.store.ApplyFeed(applyCtx, batch)

	cancelApply()

	if errors.Is(err, ErrSyncLockHeld) {
		p.log.Info("the zone sync holds its lock, so this feed batch is left for the next poll", "zones", len(batch.Zones))

		return
	}

	if err != nil {
		p.log.Error("the zone-feed batch was not written and the latest feed time did not move", "zones", len(batch.Zones), "error", err)

		return
	}

	for _, r := range batch.Refusals {
		p.log.Warn("a zone-feed item could not be stored and was recorded instead",
			"zone_id", r.ZoneID, "item_time", r.ItemTime, "reason", string(r.Reason), "subject", r.Subject)
	}

	p.log.Info("a zone-feed batch was written",
		"zones", len(batch.Zones), "inserted", applied.Inserted, "held", applied.Held,
		"refusals", len(batch.Refusals), "latest", batch.Latest)
}

// wireFeedItem is one feed item as the API sends it: `{"type":…, "time":…,
// "zone":{…}}`. The API index promises `time` and `type` on every item and
// elides the zone object to its name; what the zone carries is observed, not
// promised, which is why every field is decided per item below.
type wireFeedItem struct {
	Type *string         `json:"type"`
	Time *string         `json:"time"`
	Zone json.RawMessage `json:"zone"`
}

// feedTimeLayouts are the forms a feed time is read in. The first is the
// ISO 8601 form the API index documents for every date, and the one the live
// feed carried on 20 October 2026 (`2026-10-19T21:46:07+0000`); RFC 3339 is
// accepted beside it.
var feedTimeLayouts = []string{"2006-01-02T15:04:05Z0700", time.RFC3339}

// ParseZoneFeed maps a zone-feed response onto the batch one poll writes.
//
// A pure function over the bytes and the poll's start instant, as
// ParseAllZones is over its bytes, for the same reason: it is the whole of the
// feed's ingest decision and the step most likely to be wrong.
//
// It refuses the response only when it cannot be read as a JSON array at all;
// past that, every item is decided on its own:
//
//   - storable: its zone holds every field the copy requires, each readable,
//     and passes every per-row assertion. An item lacking only `type`,
//     `region.country` or `region.area` is storable, as in the all-zones path.
//   - refused: anything else, recorded with the reason and never guessed.
//
// One id named twice is stored once, from the storable item with the later
// feed time (`RD-042`). Latest is the newest feed time among all items, stored
// or refused (`RD-043`) — except one later than startedAt, which is not
// counted: a time ahead of the poll's own start is either a clock disagreement
// or a bad value, and moving the lower bound past it would hide every zone
// created before it. Such an item is asked for again by the next poll.
func ParseZoneFeed(body []byte, startedAt time.Time) (FeedBatch, error) {
	dec := json.NewDecoder(bytes.NewReader(body))

	open, err := dec.Token()
	if err != nil {
		return FeedBatch{}, fmt.Errorf("reading the start of the zone feed: %w", err)
	}

	if delim, ok := open.(json.Delim); !ok || delim != '[' {
		return FeedBatch{}, fmt.Errorf("the zone feed starts with %v, want a JSON array of feed items", open)
	}

	var (
		batch  FeedBatch
		chosen = map[int32]int{}
		times  []*time.Time
	)

	for index := 0; dec.More(); index++ {
		var raw json.RawMessage

		if err := dec.Decode(&raw); err != nil {
			return FeedBatch{}, fmt.Errorf("reading feed item %d: %w", index, err)
		}

		zone, itemTime, refusals := decideFeedItem(raw, startedAt)

		if itemTime != nil && !itemTime.After(startedAt) && (batch.Latest == nil || itemTime.After(*batch.Latest)) {
			batch.Latest = itemTime
		}

		if len(refusals) > 0 {
			batch.Refusals = append(batch.Refusals, refusals...)

			continue
		}

		if at, held := chosen[zone.ID]; held {
			if laterThan(itemTime, times[at]) {
				batch.Zones[at], times[at] = zone, itemTime
			}

			continue
		}

		chosen[zone.ID] = len(batch.Zones)
		batch.Zones = append(batch.Zones, zone)
		times = append(times, itemTime)
	}

	if _, err := dec.Token(); err != nil {
		return FeedBatch{}, fmt.Errorf("reading the end of the zone feed: %w", err)
	}

	return batch, nil
}

// laterThan reports whether a is a later feed time than b, a time being later
// than none.
func laterThan(a, b *time.Time) bool {
	return a != nil && (b == nil || a.After(*b))
}

// decideFeedItem decides one item: its zone when storable, or the refusals
// that say why not. The item's feed time comes back either way.
func decideFeedItem(raw json.RawMessage, startedAt time.Time) (Zone, *time.Time, []FeedRefusal) {
	var item wireFeedItem

	if err := json.Unmarshal(raw, &item); err != nil {
		return Zone{}, nil, []FeedRefusal{{Reason: FeedUnreadableField, Subject: unreadablePath(err, "")}}
	}

	itemTime := parseFeedTime(item.Time)

	refuse := func(id *int32, reason FeedRefusalReason, subjects ...string) []FeedRefusal {
		out := make([]FeedRefusal, 0, len(subjects))
		for _, s := range subjects {
			out = append(out, FeedRefusal{ZoneID: id, ItemTime: itemTime, Reason: reason, Subject: s})
		}

		return out
	}

	switch {
	case item.Type == nil:
		return Zone{}, itemTime, refuse(nil, FeedMissingField, feedSubjectType)
	case *item.Type != feedItemType:
		return Zone{}, itemTime, refuse(nil, FeedUnreadableField, feedSubjectType)
	case len(item.Zone) == 0 || bytes.Equal(bytes.TrimSpace(item.Zone), []byte("null")):
		return Zone{}, itemTime, refuse(nil, FeedMissingField, feedSubjectZone)
	}

	var w wireZone

	if err := json.Unmarshal(item.Zone, &w); err != nil {
		return Zone{}, itemTime, refuse(w.ID, FeedUnreadableField, unreadablePath(err, feedSubjectZone))
	}

	if missing := w.missingFields(); len(missing) > 0 {
		return Zone{}, itemTime, refuse(w.ID, FeedMissingField, missing...)
	}

	if unreadable := w.unreadableFields(); len(unreadable) > 0 {
		return Zone{}, itemTime, refuse(w.ID, FeedUnreadableField, unreadable...)
	}

	// The all-zones mapping, unchanged. Every condition it refuses has been
	// decided above, so an error here is one this file did not anticipate,
	// and it is recorded rather than guessed past.
	zone, err := w.zone()
	if err != nil {
		return Zone{}, itemTime, refuse(w.ID, FeedUnreadableField, feedSubjectZone)
	}

	if failed := zone.failedAssertions(startedAt); len(failed) > 0 {
		return Zone{}, itemTime, refuse(w.ID, FeedAssertion, failed...)
	}

	return zone, itemTime, nil
}

// parseFeedTime reads a feed time, or nil when there is none it can read. A
// feed time is never required to store an item's zone — the copy holds no
// column for it — so an unreadable one costs only the item's place in Latest.
func parseFeedTime(raw *string) *time.Time {
	if raw == nil {
		return nil
	}

	for _, layout := range feedTimeLayouts {
		if t, err := time.Parse(layout, *raw); err == nil {
			t = t.UTC()

			return &t
		}
	}

	return nil
}

// unreadablePath names the field a decode error is about, as its path in the
// item. A value of the wrong JSON type, or a number its Go field cannot hold,
// is an UnmarshalTypeError naming the field; anything else is about the whole
// of what was being decoded.
func unreadablePath(err error, prefix string) string {
	var typeErr *json.UnmarshalTypeError

	if errors.As(err, &typeErr) && typeErr.Field != "" {
		if prefix == "" {
			return typeErr.Field
		}

		return prefix + "." + typeErr.Field
	}

	if prefix == "" {
		return feedSubjectItem
	}

	return prefix
}

// missingFields names every field the copy requires that the zone lacks, in
// the order of the mapping in `migrations/README.md § The zone ingest field
// mapping`. They are the fields wireZone.zone refuses a record without, which
// are the zone table's NOT NULL columns under `Architecture.md § The zone
// table`.
func (w wireZone) missingFields() []string {
	var missing []string

	for _, f := range []struct {
		path   string
		absent bool
	}{
		{"id", w.ID == nil},
		{"name", w.Name == nil},
		{"latitude", w.Latitude == nil},
		{"longitude", w.Longitude == nil},
		{"dateCreated", w.DateCreated == nil},
		{"totalTakeovers", w.TotalTakeovers == nil},
		{"takeoverPoints", w.TakeoverPoints == nil},
		{"pointsPerHour", w.PointsPerHour == nil},
		{"region", w.Region == nil},
		{"region.id", w.Region != nil && w.Region.ID == nil},
		{"region.name", w.Region != nil && w.Region.Name == nil},
	} {
		if f.absent {
			missing = append(missing, feedSubjectZonePathPrefix+f.path)
		}
	}

	return missing
}

// unreadableFields names every present field whose value the copy cannot
// hold: a dateCreated in no accepted form, and text carrying a NUL byte, which
// PostgreSQL's text type refuses and which would abort the batch's write.
func (w wireZone) unreadableFields() []string {
	var unreadable []string

	if w.DateCreated != nil {
		if _, err := parseDateCreated(*w.DateCreated); err != nil {
			unreadable = append(unreadable, feedSubjectZonePathPrefix+"dateCreated")
		}
	}

	texts := map[string]*string{"name": w.Name}

	if w.Region != nil {
		texts["region.name"] = w.Region.Name
		texts["region.country"] = w.Region.Country

		if w.Region.Area != nil {
			texts["region.area.name"] = w.Region.Area.Name
		}
	}

	// Ranged in a fixed order, so the same item always records the same
	// refusals in the same order.
	for _, path := range []string{"name", "region.name", "region.country", "region.area.name"} {
		if s := texts[path]; s != nil && strings.ContainsRune(*s, 0) {
			unreadable = append(unreadable, feedSubjectZonePathPrefix+path)
		}
	}

	return unreadable
}

// failedAssertions names every per-row assertion the mapped zone fails.
func (z Zone) failedAssertions(startedAt time.Time) []string {
	var failed []string

	if z.Latitude < -90 || z.Latitude > 90 {
		failed = append(failed, AssertLatitudeRange)
	}

	if z.Longitude < -180 || z.Longitude > 180 {
		failed = append(failed, AssertLongitudeRange)
	}

	if z.TotalTakeovers < 0 {
		failed = append(failed, AssertTakeoversNonNeg)
	}

	if z.DateCreated.After(startedAt) {
		failed = append(failed, AssertCreatedNotAfterRun)
	}

	return failed
}

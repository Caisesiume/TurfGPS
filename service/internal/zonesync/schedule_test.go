package zonesync

import (
	"testing"
	"time"

	// The tests read a real zone's daylight-saving rules, and embedding the
	// database is what keeps that from depending on the host, as it is for the
	// service through `internal/config`.
	_ "time/tzdata"
)

// utc is an instant written in UTC, so each expectation below states the UTC
// hour the local time lands on and a daylight-saving change is visible in it.
func utc(year int, month time.Month, day, hour, minute int) time.Time {
	return time.Date(year, month, day, hour, minute, 0, 0, time.UTC)
}

// TestTheMonthlyInstantIsTheFirstSundayAtItsLocalTime binds the calendar rule
// for months that begin on a Sunday, a Monday and a Saturday — the edges of
// "first Sunday", where an off-by-one lands on the 8th, the 0th or the last day
// of the month before — and on either side of daylight saving.
func TestTheMonthlyInstantIsTheFirstSundayAtItsLocalTime(t *testing.T) {
	t.Parallel()

	for _, c := range []struct {
		name   string
		year   int
		month  time.Month
		begins time.Weekday
		want   time.Time
	}{
		{"a month beginning on a Sunday, in winter time", 2026, time.February, time.Sunday, utc(2026, time.February, 1, 6, 45)},
		{"a month beginning on a Sunday, in summer time", 2026, time.November, time.Sunday, utc(2026, time.November, 1, 6, 45)},
		{"a month beginning on a Monday", 2026, time.June, time.Monday, utc(2026, time.June, 7, 5, 45)},
		{"a month beginning on a Saturday", 2026, time.August, time.Saturday, utc(2026, time.August, 2, 5, 45)},
		{"a month beginning on a Wednesday", 2026, time.April, time.Wednesday, utc(2026, time.April, 5, 5, 45)},
	} {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			// The table's own premise, checked rather than trusted, so a wrong
			// row cannot pass by testing a different edge than it names.
			if got := time.Date(c.year, c.month, 1, 12, 0, 0, 0, testSchedule.Location).Weekday(); got != c.begins {
				t.Fatalf("the row says %s %d begins on a %s, and it begins on a %s", c.month, c.year, c.begins, got)
			}

			got := testSchedule.instantOf(c.year, c.month)
			if !got.Equal(c.want) {
				t.Errorf("the instant of %s %d is %s, want %s (the first Sunday at %02d:%02d in %s)",
					c.month, c.year, got.UTC().Format(time.RFC3339), c.want.Format(time.RFC3339),
					testSchedule.Hour, testSchedule.Minute, testSchedule.Location)
			}

			if local := got.In(testSchedule.Location); local.Weekday() != time.Sunday || local.Day() > 7 ||
				local.Hour() != testSchedule.Hour || local.Minute() != testSchedule.Minute {
				t.Errorf("the instant of %s %d reads %s locally, want a Sunday among the first seven days at %02d:%02d",
					c.month, c.year, local.Format(time.RFC1123), testSchedule.Hour, testSchedule.Minute)
			}
		})
	}
}

// TestTheLatestAndNextInstantsKeepTheLocalTimeAcrossDaylightSaving binds the
// two answers the scheduler asks for — the most recent instant at or before
// now, and the earliest after it — across both changes of the clocks, across a
// year boundary, and at an instant exactly.
//
// Each pair straddles a change, so a rule that held a fixed UTC hour, or added
// whole days to the last instant, is an hour wrong on one side of it.
func TestTheLatestAndNextInstantsKeepTheLocalTimeAcrossDaylightSaving(t *testing.T) {
	t.Parallel()

	for _, c := range []struct {
		name         string
		now          time.Time
		latest, next time.Time
	}{
		{
			"just after summer time begins: winter instant behind, summer instant ahead",
			utc(2026, time.March, 30, 12, 0), utc(2026, time.March, 1, 6, 45), utc(2026, time.April, 5, 5, 45),
		},
		{
			"just after summer time ends: summer instant behind, winter instant ahead",
			utc(2026, time.October, 26, 12, 0), utc(2026, time.October, 4, 5, 45), utc(2026, time.November, 1, 6, 45),
		},
		{
			"across the year boundary",
			utc(2026, time.December, 20, 0, 0), utc(2026, time.December, 6, 6, 45), utc(2027, time.January, 3, 6, 45),
		},
		{
			"before the month's instant: last month's is the latest",
			utc(2027, time.January, 2, 23, 0), utc(2026, time.December, 6, 6, 45), utc(2027, time.January, 3, 6, 45),
		},
		{
			"exactly at an instant: it is the latest, and the next is a month on",
			utc(2026, time.April, 5, 5, 45), utc(2026, time.April, 5, 5, 45), utc(2026, time.May, 3, 5, 45),
		},
		{
			"a moment before an instant: it is still the next",
			utc(2026, time.April, 5, 5, 44).Add(59 * time.Second), utc(2026, time.March, 1, 6, 45), utc(2026, time.April, 5, 5, 45),
		},
	} {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			if got := testSchedule.latest(c.now); !got.Equal(c.latest) {
				t.Errorf("at %s the latest instant is %s, want %s",
					c.now.Format(time.RFC3339), got.UTC().Format(time.RFC3339), c.latest.Format(time.RFC3339))
			}

			if got := testSchedule.next(c.now); !got.Equal(c.next) {
				t.Errorf("at %s the next instant is %s, want %s",
					c.now.Format(time.RFC3339), got.UTC().Format(time.RFC3339), c.next.Format(time.RFC3339))
			}
		})
	}
}

// TestAMonthlyRuleNoInstantCanBeBuiltFromIsRefused binds the constructor's
// refusals, each naming its own field.
func TestAMonthlyRuleNoInstantCanBeBuiltFromIsRefused(t *testing.T) {
	t.Parallel()

	for _, c := range []struct {
		name string
		rule Monthly
	}{
		{"no time zone", Monthly{Hour: 7, Minute: 45}},
		{"an hour past the day", Monthly{Hour: 24, Minute: 0, Location: time.UTC}},
		{"a negative hour", Monthly{Hour: -1, Minute: 0, Location: time.UTC}},
		{"a minute past the hour", Monthly{Hour: 7, Minute: 60, Location: time.UTC}},
	} {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			if err := c.rule.validate(); err == nil {
				t.Errorf("the rule %+v was accepted, want it refused", c.rule)
			}
		})
	}
}

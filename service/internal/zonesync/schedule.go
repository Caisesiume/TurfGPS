package zonesync

import (
	"errors"
	"fmt"
	"time"
)

// Monthly is when the month's all-zones refresh falls due: the first Sunday of
// each month, at a local time of day, read in one location (#233).
//
// THE RULE'S SHAPE IS HERE AND ITS FIGURES ARE NOT. The Owner's decision on #20
// fixes the first Sunday; the time of day and the location reach this type
// through `internal/config`, which holds no default for either, so this package
// is a second home for neither. Their home is `Architecture.md § Retrieving
// zones` once the amendment the Owner approved on #20 lands, and that decision
// until then.
//
// THE INSTANT IS LOCAL CALENDAR ARITHMETIC, NOT A FIXED UTC HOUR. It is built
// with time.Date in Location, so the same local time is kept on both sides of a
// daylight-saving change and the UTC hour moves instead. The location comes from
// the time zone database, which `internal/config` embeds so that the answer does
// not depend on the host it runs on.
type Monthly struct {
	Hour     int
	Minute   int
	Location *time.Location
}

// validate refuses a rule no instant can be built from. Each refusal names its
// field, so a crossed mapping in the composition root says which one it reached.
func (m Monthly) validate() error {
	if m.Location == nil {
		return errors.New("the monthly schedule has no time zone")
	}

	if m.Hour < 0 || m.Hour > 23 {
		return fmt.Errorf("the monthly schedule's hour %d is outside 0 to 23", m.Hour)
	}

	if m.Minute < 0 || m.Minute > 59 {
		return fmt.Errorf("the monthly schedule's minute %d is outside 0 to 59", m.Minute)
	}

	return nil
}

// instantOf is the instant of the month that year and month name, after the
// normalisation time.Date applies, so month 0 is the previous December and
// month 13 the next January.
func (m Monthly) instantOf(year int, month time.Month) time.Time {
	// Noon rather than midnight, because a zone that changes its clocks at
	// midnight has no midnight on that day, and only the weekday is read here.
	first := time.Date(year, month, 1, 12, 0, 0, 0, m.Location)
	toSunday := (7 - int(first.Weekday())) % 7

	return time.Date(first.Year(), first.Month(), 1+toSunday, m.Hour, m.Minute, 0, 0, m.Location)
}

// latest is the most recent instant at or before now.
//
// It is what the month's refresh is measured against, and anchoring on it
// rather than on the next instant is what makes an empty copy, or a process that
// was down through the instant, refresh at once instead of waiting a month
// (`DECISIONS.md § RD-044`).
func (m Monthly) latest(now time.Time) time.Time {
	local := now.In(m.Location)

	instant := m.instantOf(local.Year(), local.Month())
	if instant.After(now) {
		instant = m.instantOf(local.Year(), local.Month()-1)
	}

	return instant
}

// next is the earliest instant after now.
func (m Monthly) next(now time.Time) time.Time {
	local := now.In(m.Location)

	instant := m.instantOf(local.Year(), local.Month())
	if !instant.After(now) {
		instant = m.instantOf(local.Year(), local.Month()+1)
	}

	return instant
}

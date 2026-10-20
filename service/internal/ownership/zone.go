package ownership

import (
	"encoding/json"
	"fmt"
	"time"
)

// Zone is a zone record's ownership fields, as a zone response carried them
// (`FR-030`).
//
// A NIL FIELD IS A FIELD THE API DID NOT SEND, and nothing else. Both are
// round-scoped and absent for a zone nobody has taken this round, per
// `Architecture.md § Volatile and optional fields`, so absence is the ordinary
// case. A zero Owner or a zero time in their place would report an owner or a
// date the API never sent, and would erase the only signal the absence carries;
// what the absence may NOT be read as is `FR-031`'s, not this type's.
type Zone struct {
	ID            int32
	CurrentOwner  *Owner
	DateLastTaken *time.Time
}

// Owner is a zone's current owner as the response names them.
type Owner struct {
	ID   int64
	Name string
}

// dateLastTakenLayout is the form `dateLastTaken` takes on the wire: ISO 8601
// with a numeric offset written without a colon, as the API's own documentation gives every
// date ("2013-06-08T12:59:19+0000"). "Z0700" also admits a literal Z.
const dateLastTakenLayout = "2006-01-02T15:04:05Z0700"

type wireZone struct {
	ID            *int32     `json:"id"`
	CurrentOwner  *wireOwner `json:"currentOwner"`
	DateLastTaken *string    `json:"dateLastTaken"`
}

type wireOwner struct {
	ID   *int64 `json:"id"`
	Name string `json:"name"`
}

// ParseZones reads a zone response — the array the API's zones endpoint returns — and
// carries each record's `currentOwner` and `dateLastTaken` through as absent
// where the record omits them.
//
// It refuses a response it cannot read rather than a record with a field
// missing: an absent owner or date is data, while a record with no id, an
// owner with no id, or a date in no form the API documents is a response this
// system does not understand.
func ParseZones(body []byte) ([]Zone, error) {
	var wire []wireZone
	if err := json.Unmarshal(body, &wire); err != nil {
		return nil, fmt.Errorf("the zone response is not a JSON array of zones: %w; body %s", err, snippet(body))
	}

	zones := make([]Zone, 0, len(wire))

	for i, w := range wire {
		if w.ID == nil {
			return nil, fmt.Errorf("zone record %d has no id", i)
		}

		z := Zone{ID: *w.ID}

		if w.CurrentOwner != nil {
			if w.CurrentOwner.ID == nil {
				return nil, fmt.Errorf("zone %d: currentOwner is present with no id", z.ID)
			}

			z.CurrentOwner = &Owner{ID: *w.CurrentOwner.ID, Name: w.CurrentOwner.Name}
		}

		if w.DateLastTaken != nil {
			at, err := time.Parse(dateLastTakenLayout, *w.DateLastTaken)
			if err != nil {
				return nil, fmt.Errorf("zone %d: dateLastTaken=%q is not in the documented form %q: %w", z.ID, *w.DateLastTaken, dateLastTakenLayout, err)
			}

			at = at.UTC()
			z.DateLastTaken = &at
		}

		zones = append(zones, z)
	}

	return zones, nil
}

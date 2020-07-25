package models

import (
	"fmt"
	"strings"
	"time"
)

// Duration allows for string-based representation of durations in JSON.
type Duration struct {
	time.Duration
}

// UnmarshalJSON converts a JSON element into a Duration.
func (d *Duration) UnmarshalJSON(b []byte) (err error) {
	d.Duration, err = time.ParseDuration(strings.Trim(string(b), `"`))
	return
}

// MarshalJSON marshals a Duration into a JSON element.
func (d Duration) MarshalJSON() (b []byte, err error) {
	return []byte(fmt.Sprintf(`"%s"`, d.String())), nil
}

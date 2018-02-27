package main

import (
	"fmt"
	"strings"
	"time"
)

// JSONTime represents a datetime.
type JSONTime time.Time

// MarshalJSON returns the momentjs like 'YYYY-MM-DD hh:mm A' formated datetime.
func (t JSONTime) MarshalJSON() ([]byte, error) {
	stamp := fmt.Sprintf("\"%s\"", time.Time(t).Format("2006-01-02 03:04:05.0000 PM"))
	return []byte(stamp), nil
}

// UnmarshalJSON returns the parsed time object.
func (t *JSONTime) UnmarshalJSON(b []byte) (err error) {
	s := strings.Trim(string(b), "\"")
	if s == "null" {
		*t = JSONTime(time.Now())
		return
	}
	p, err := time.Parse("2006-01-02 03:04:05.0000 PM", s)
	*t = JSONTime(p)
	return err
}

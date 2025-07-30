package main

import (
	"fmt"
	"time"

	"github.com/go-playground/form/v4"
)

// Registers a custom form decoder for time.Time to handle the list of time
// formats. Each timeFormat will be tried in turn and an error returned if none
// of them can be parsed. If the slice of timeFormats is nil or empty, then a
// default selection will be used.
func FormDecoderRegisterTimeType(fd *form.Decoder, timeFormats []string) {
	if timeFormats == nil || len(timeFormats) == 0 {
		timeFormats = []string{
			time.DateOnly,
			time.TimeOnly,
			"2006-01-02T15:04:05",
			"2006-01-02T15:04",
			time.RFC3339Nano,
			time.RFC3339,
			time.DateTime,
		}
	}

	fd.RegisterCustomTypeFunc(func(vals []string) (interface{}, error) {
		timeStr := vals[0]

		for _, format := range timeFormats {
			t, err := time.Parse(format, timeStr)
			if err == nil {
				return t, nil
			}
		}

		msg := "failed to decode time value '%s' from form using formats: %s"
		return time.Time{}, fmt.Errorf(msg, timeStr, timeFormats)
	}, time.Time{})
}

// Registers a custom form decoder for time.Location time zone locations.
func FormDecoderRegisterTimeLocationType(fd *form.Decoder) {
	fd.RegisterCustomTypeFunc(func(vals []string) (interface{}, error) {
		locationStr := vals[0]
		location, err := time.LoadLocation(locationStr)

		if err != nil {
			msg := "failed to decode time location value '%s' from form: %w"
			return time.Location{}, fmt.Errorf(msg, locationStr, err)
		}

		return *location, nil
	}, time.Location{})
}

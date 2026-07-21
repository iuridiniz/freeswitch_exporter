package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// parseBuckets parses a comma-separated list of histogram bucket upper bounds.
// Values must be > 0 and strictly increasing.
func parseBuckets(s string) ([]float64, error) {
	parts := strings.Split(s, ",")

	buckets := make([]float64, 0, len(parts))

	for _, part := range parts {
		trimmed := strings.TrimSpace(part)

		value, err := strconv.ParseFloat(trimmed, 64)

		if err != nil {
			return nil, fmt.Errorf("invalid bucket value %q: not a number", trimmed)
		}

		if value <= 0 {
			return nil, fmt.Errorf("invalid bucket value %q: must be > 0", trimmed)
		}

		if len(buckets) > 0 && value <= buckets[len(buckets)-1] {
			return nil, fmt.Errorf("invalid bucket value %q: buckets must be strictly increasing", trimmed)
		}

		buckets = append(buckets, value)
	}

	return buckets, nil
}

// parseCreatedEpoch parses a FreeSWITCH created_epoch string (epoch seconds).
// Returns an error if the value is not a positive decimal integer.
func parseCreatedEpoch(s string) (time.Time, error) {
	v, err := strconv.ParseInt(strings.TrimSpace(s), 10, 64)

	if err != nil {
		return time.Time{}, fmt.Errorf("invalid created_epoch %q: not a decimal integer", s)
	}

	if v <= 0 {
		return time.Time{}, fmt.Errorf("invalid created_epoch %q: must be > 0", s)
	}

	return time.Unix(v, 0), nil
}

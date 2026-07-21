package main

import (
	"fmt"
	"strconv"
	"strings"
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

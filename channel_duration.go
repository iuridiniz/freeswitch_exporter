package main

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
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

// buildChannelDurationHistogram builds a histogram of active channel ages (in seconds).
// Skips rows with invalid created_epoch and clamps negative ages to 0.
func buildChannelDurationHistogram(rows []map[string]any, now time.Time, buckets []float64) (prometheus.Histogram, int) {
	histogram := prometheus.NewHistogram(prometheus.HistogramOpts{
		Namespace: namespace,
		Name:      "channel_duration_seconds",
		Help:      "Age of currently active FreeSWITCH channels in seconds, observed at scrape time.",
		Buckets:   buckets,
	})

	skipped := 0

	for _, row := range rows {
		raw, ok := row["created_epoch"].(string)

		if !ok {
			skipped++
			continue
		}

		start, err := parseCreatedEpoch(raw)

		if err != nil {
			skipped++
			continue
		}

		age := now.Sub(start).Seconds()

		if age < 0 {
			age = 0
		}

		histogram.Observe(age)
	}

	return histogram, skipped
}

// channelRowsPayload mirrors the JSON shape of "api show channels as json".
type channelRowsPayload struct {
	RowCount int              `json:"row_count"`
	Rows     []map[string]any `json:"rows"`
}

// decodeChannelRows decodes the channel rows JSON payload into a slice of maps.
func decodeChannelRows(payload []byte) ([]map[string]any, error) {
	var p channelRowsPayload

	if err := json.Unmarshal(payload, &p); err != nil {
		return nil, fmt.Errorf("cannot decode channel rows JSON: %w", err)
	}

	return p.Rows, nil
}

// channelDurationMetrics fetches channels, builds an age histogram, and emits it.
func (c *Collector) channelDurationMetrics(ch chan<- prometheus.Metric) error {
	response, err := c.fsCommand("api show channels as json")

	if err != nil {
		return err
	}

	rows, err := decodeChannelRows(response)

	if err != nil {
		return err
	}

	histogram, skipped := buildChannelDurationHistogram(rows, time.Now(), c.channelDurationBuckets)

	if skipped > 0 {
		level.Debug(c.logger).Log("msg", "skipped unparseable channel row", "count", skipped)
	}

	ch <- histogram

	return nil
}

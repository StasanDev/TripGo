package config

import (
	"fmt"
	"strconv"
	"time"
)

func parseDurations(rawValues map[string]string, names []string) (map[string]time.Duration, error) {
	values := make(map[string]time.Duration, len(names))
	for _, name := range names {
		value, err := time.ParseDuration(rawValues[name])
		if err != nil {
			return nil, fmt.Errorf("parse %s: %w", name, err)
		}
		if value <= 0 {
			return nil, fmt.Errorf("%s must be positive", name)
		}
		values[name] = value
	}

	return values, nil
}

func parseInt32s(rawValues map[string]string, names []string) (map[string]int32, error) {
	values := make(map[string]int32, len(names))
	for _, name := range names {
		value, err := strconv.ParseInt(rawValues[name], 10, 32)
		if err != nil {
			return nil, fmt.Errorf("parse %s: %w", name, err)
		}
		if value <= 0 {
			return nil, fmt.Errorf("%s must be positive", name)
		}
		values[name] = int32(value)
	}

	return values, nil
}

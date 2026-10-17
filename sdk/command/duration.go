package command

import (
	"strconv"
	"time"
)

// Duration reads the value of a --timeout under the common duration rules
// of Global flags in contract.md: a duration such as 30s, 5m, or 2h, or a
// number of seconds, above 0. It reports false for any other value.
func Duration(value string) (time.Duration, bool) {
	d, err := time.ParseDuration(value)
	if seconds, convErr := strconv.ParseUint(value, 10, 31); convErr == nil {
		d, err = time.Duration(seconds)*time.Second, nil
	}
	return d, err == nil && d > 0
}

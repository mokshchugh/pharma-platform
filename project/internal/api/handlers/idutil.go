package handlers

import (
	"strconv"
	"strings"
)

// parseTrailingID extracts the trailing numeric ID from strings like
// "machine-5" or "tag-12", falling back to parsing the whole string as a
// number (e.g. "5"). Returns 0 if neither form parses.
func parseTrailingID(s string) int {
	if parts := strings.SplitN(s, "-", 2); len(parts) == 2 {
		if n, err := strconv.Atoi(parts[1]); err == nil {
			return n
		}
	}
	if n, err := strconv.Atoi(s); err == nil {
		return n
	}
	return 0
}

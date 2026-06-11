package pscan

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var memRe = regexp.MustCompile(`^([\d.]+)([GMK])?$`)

// ParseMemToKB converts a memory string like "2.0G" or "494.2M" to KB.
func ParseMemToKB(raw string) int {
	raw = strings.TrimSpace(raw)
	raw = strings.ReplaceAll(raw, ",", "")
	m := memRe.FindStringSubmatch(raw)
	if m == nil {
		return 0
	}
	val, _ := strconv.ParseFloat(m[1], 64)
	switch m[2] {
	case "G":
		return int(val * 1024 * 1024)
	case "M":
		return int(val * 1024)
	case "K":
		return int(val)
	default:
		return int(val) / 1024
	}
}

// FmtKB formats KB into a human-readable string (e.g. "494.2MB").
func FmtKB(kb int) string {
	if kb < 1024 {
		return fmt.Sprintf("%dKB", kb)
	}
	return fmt.Sprintf("%.1fMB", float64(kb)/1024.0)
}

// TruncateLen truncates a string to at most max runes, appending "...".
func TruncateLen(s string, max int) string {
	runes := []rune(s)
	if len(runes) <= max {
		return s
	}
	return string(runes[:max-3]) + "..."
}

// TruncateTail truncates a string to at most max runes, keeping the tail.
func TruncateTail(s string, max int) string {
	runes := []rune(s)
	if len(runes) <= max {
		return s
	}
	return "..." + string(runes[len(runes)-max+3:])
}
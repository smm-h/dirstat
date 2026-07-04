package render

import (
	"fmt"
	"strconv"
)

var sizeUnits = []string{"B", "KB", "MB", "GB", "TB"}

// formatSize renders a byte size. Human mode uses 1024-based units with one
// decimal, except bare bytes which stay integral (prototype format: "123B",
// "1.5KB"). Raw mode returns the plain integer.
func formatSize(n int64, human bool) string {
	if !human {
		return strconv.FormatInt(n, 10)
	}
	f := float64(n)
	for _, unit := range sizeUnits {
		if f < 1024 && f > -1024 {
			if unit == "B" {
				return fmt.Sprintf("%d%s", n, unit)
			}
			return fmt.Sprintf("%.1f%s", f, unit)
		}
		f /= 1024
	}
	return fmt.Sprintf("%.1fPB", f)
}

// formatCount renders an integer, with thousands separators in human mode.
func formatCount(n int64, human bool) string {
	s := strconv.FormatInt(n, 10)
	if !human {
		return s
	}
	neg := false
	if s[0] == '-' {
		neg = true
		s = s[1:]
	}
	var out []byte
	for i, c := range []byte(s) {
		if i > 0 && (len(s)-i)%3 == 0 {
			out = append(out, ',')
		}
		out = append(out, c)
	}
	if neg {
		return "-" + string(out)
	}
	return string(out)
}

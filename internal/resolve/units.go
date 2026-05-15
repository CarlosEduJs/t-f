package resolve

import (
	"regexp"
	"strconv"
	"strings"
)

var remRe = regexp.MustCompile(`([0-9.]+)\s*rem\b`)
var dimensionRe = regexp.MustCompile(`^([0-9.]+)\s*(.*)$`)

const RemBase = 16.0

func RemToPx(value string, base float64) string {
	if base == 0 {
		base = RemBase
	}
	return remRe.ReplaceAllStringFunc(value, func(m string) string {
		parts := remRe.FindStringSubmatch(m)
		if len(parts) < 2 {
			return m
		}
		v, err := strconv.ParseFloat(parts[1], 64)
		if err != nil {
			return m
		}
		px := v * base
		if px == float64(int64(px)) {
			return strconv.FormatFloat(px, 'f', 0, 64) + "px"
		}
		return strconv.FormatFloat(px, 'f', 2, 64) + "px"
	})
}

func ConvertAllRem(vars map[string]string, base float64) map[string]string {
	result := make(map[string]string, len(vars))
	for name, value := range vars {
		result[name] = RemToPx(value, base)
	}
	return result
}

func StripUnit(value string) (float64, string, bool) {
	value = strings.TrimSpace(value)
	m := dimensionRe.FindStringSubmatch(value)
	if m == nil {
		v, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return 0, "", false
		}
		return v, "", true
	}
	v, err := strconv.ParseFloat(m[1], 64)
	if err != nil {
		return 0, "", false
	}
	return v, m[2], true
}

func HasUnit(value string) bool {
	_, _, ok := StripUnit(value)
	return ok
}

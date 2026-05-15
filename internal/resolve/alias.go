package resolve

import (
	"fmt"
	"regexp"
	"strings"
)

var varRe = regexp.MustCompile(`var\((--[^)]+)\)`)

type AliasResolver struct {
	vars map[string]string
}

func NewAliasResolver(vars map[string]string) *AliasResolver {
	return &AliasResolver{vars: vars}
}

func (r *AliasResolver) Resolve(value string) (string, error) {
	seen := make(map[string]bool)
	return r.resolve(value, seen, 0)
}

func (r *AliasResolver) resolve(value string, seen map[string]bool, depth int) (string, error) {
	if depth > 64 {
		return value, fmt.Errorf("max alias depth exceeded")
	}

	for {
		match := varRe.FindStringSubmatch(value)
		if match == nil {
			break
		}
		varName := match[1]

		if seen[varName] {
			return value, fmt.Errorf("circular reference detected: %s", varName)
		}
		seen[varName] = true

		resolved, ok := r.vars[varName]
		if !ok {
			return value, fmt.Errorf("undefined variable: %s", varName)
		}

		resolved, err := r.resolve(resolved, seen, depth+1)
		if err != nil {
			return value, err
		}

		value = strings.Replace(value, match[0], resolved, 1)
	}

	return value, nil
}

func ResolveAll(vars map[string]string) (map[string]string, error) {
	r := NewAliasResolver(vars)
	result := make(map[string]string, len(vars))

	for name, value := range vars {
		resolved, err := r.Resolve(value)
		if err != nil {
			result[name] = value
			continue
		}
		result[name] = resolved
	}

	return result, nil
}

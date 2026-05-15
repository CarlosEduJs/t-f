package tokens

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"t-f/internal/domain"
)

// Generator converts CSS variables into DTCG design tokens.
type Generator struct {
	RemBase   float64
	FigmaMode bool
}

// NewGenerator returns a Generator with sensible defaults (rem base = 16).
func NewGenerator() *Generator {
	return &Generator{RemBase: 16}
}

// Generate produces a DTCG-formatted JSON byte slice from the given variables.
func (g *Generator) Generate(vars []domain.Variable) ([]byte, error) {
	themed := splitByTheme(vars)
	resolved := resolveAll(themed)

	groups := categorizeAndGroup(resolved)
	output := domain.DesignTokens{
		Semantic: make(map[string]any),
	}

	typographyGroups := make(map[string]map[string]string)

	for category, entries := range groups {
		if category == "typography" {
			collectTypography(entries, typographyGroups)
			continue
		}
		if category == "other" {
			continue
		}
		if g.FigmaMode {
			output.Semantic[category] = flattenTree(entries)
		} else {
			output.Semantic[category] = buildTree(entries)
		}
	}

	if len(typographyGroups) > 0 {
		typoTree := make(domain.DTCGGroup)
		for name, props := range typographyGroups {
			buildTypographyToken(typoTree, name, props)
		}
		output.Semantic["typography"] = typoTree
	}

	if g.FigmaMode {
		convertToFigma(output.Semantic)
	}

	return json.MarshalIndent(output, "", "  ")
}

func convertToFigma(semantic map[string]any) {
	for _, node := range semantic {
		if group, ok := node.(domain.DTCGGroup); ok {
			convertNode(group)
		}
	}
}

func convertNode(node domain.DTCGGroup) {
	ttype, _ := node["$type"].(domain.TokenType)

	for key, val := range node {
		if key == "$value" {
			str, ok := val.(string)
			if !ok {
				continue
			}
			switch ttype {
			case domain.TypeColor:
				hex := domain.ConvertColorToHEX(str)
				comps := domain.HexToComponents(hex)
				node[key] = domain.FigmaColorValue{
					ColorSpace: "srgb",
					Components: comps,
					Hex:        hex,
				}
			case domain.TypeDimension:
				v, unit := domain.ParseDimension(str)
				if unit == "" {
					unit = "px"
				}
				node[key] = domain.FigmaDimensionValue{
					Value: v,
					Unit:  unit,
				}
			}
		} else if key == "$type" || key == "$description" {
			continue
		} else if sub, ok := val.(domain.DTCGGroup); ok {
			convertNode(sub)
		}
	}
}

func splitByTheme(vars []domain.Variable) map[string]map[string]string {
	result := map[string]map[string]string{
		"light": make(map[string]string),
		"dark":  make(map[string]string),
		"theme": make(map[string]string),
	}

	for _, v := range vars {
		themeKey := string(v.Theme)
		if _, ok := result[themeKey]; !ok {
			result[themeKey] = make(map[string]string)
		}
		result[themeKey][v.Name] = v.Value
	}

	return result
}

func resolveAll(themed map[string]map[string]string) map[string]map[string]string {
	result := make(map[string]map[string]string)

	for theme, vars := range themed {
		scope := make(map[string]string)
		for k, v := range themed["light"] {
			scope[k] = v
		}
		for k, v := range themed["theme"] {
			scope[k] = v
		}
		if theme == "dark" {
			for k, v := range themed["dark"] {
				scope[k] = v
			}
		}

		resolved := make(map[string]string)

		for k, v := range vars {
			aliased := resolveAlias(k, v, scope)
			remmed := remToPx(aliased, 16)
			calcResolved := evalCalc(remmed)
			resolved[k] = calcResolved
		}

		result[theme] = resolved
	}

	return result
}

func resolveAlias(name, value string, allVars map[string]string) string {
	maxIter := 100
	current := value

	for i := 0; i < maxIter; i++ {
		start := strings.Index(current, "var(")
		if start == -1 {
			break
		}
		end := strings.Index(current[start:], ")")
		if end == -1 {
			break
		}
		inner := current[start+4 : start+end]
		inner = strings.TrimSpace(inner)

		resolved, ok := allVars[inner]
		if !ok {
			break
		}

		if inner == name {
			break
		}

		current = current[:start] + resolved + current[start+end+1:]
	}

	return current
}

func evalCalc(value string) string {
	if !strings.Contains(value, "calc(") {
		return value
	}

	current := value
	maxIter := 10

	for i := 0; i < maxIter; i++ {
		start := strings.Index(current, "calc(")
		if start == -1 {
			break
		}

		depth := 1
		pos := start + 5
		for depth > 0 && pos < len(current) {
			switch current[pos] {
			case '(':
				depth++
			case ')':
				depth--
			}
			pos++
		}
		inner := current[start+5 : pos-1]

		parts := strings.Fields(inner)
		if len(parts) == 3 {
			a, aUnit := splitNumber(parts[0])
			op := parts[1]
			b, bUnit := splitNumber(parts[2])

			var result float64
			unit := bUnit
			if unit == "" {
				unit = aUnit
			}

			switch op {
			case "+":
				result = toPxNum(a, aUnit) + toPxNum(b, bUnit)
				unit = chooseUnit(aUnit, bUnit)
			case "-":
				result = toPxNum(a, aUnit) - toPxNum(b, bUnit)
				unit = chooseUnit(aUnit, bUnit)
			case "*":
				result = a * b
			case "/":
				if b != 0 {
					result = a / b
				}
			}

			if unit == "rem" {
				result *= 16
				unit = "px"
			}
			current = current[:start] + fmt.Sprintf("%.0f%s", result, unit) + current[pos:]
		} else if len(parts) == 2 {
			a, aUnit := splitNumber(parts[0])
			b, bUnit := splitNumber(parts[1])

			result := a * b
			unit := aUnit
			if unit == "" {
				unit = bUnit
			}
			if unit == "rem" {
				result *= 16
				unit = "px"
			}
			current = current[:start] + fmt.Sprintf("%.0f%s", result, unit) + current[pos:]
		}

		current = strings.ReplaceAll(current, "  ", " ")
	}

	return current
}

func chooseUnit(a, b string) string {
	if a == b {
		return a
	}
	if a == "" {
		return b
	}
	if b == "" {
		return a
	}
	return a
}

func splitNumber(s string) (float64, string) {
	s = strings.TrimSpace(s)
	unitStart := -1
	for i, c := range s {
		if c == '.' || (c >= '0' && c <= '9') || c == '-' || c == '+' {
			continue
		}
		unitStart = i
		break
	}
	if unitStart == -1 {
		v, _ := strconv.ParseFloat(s, 64)
		return v, ""
	}
	v, _ := strconv.ParseFloat(s[:unitStart], 64)
	return v, s[unitStart:]
}

func toPxNum(v float64, unit string) float64 {
	if unit == "rem" {
		return v * 16
	}
	return v
}

func remToPx(value string, base float64) string {
	if !strings.Contains(value, "rem") {
		return value
	}
	result := value
	for {
		idx := strings.Index(result, "rem")
		if idx == -1 {
			break
		}
		start := idx
		for start > 0 {
			c := result[start-1]
			if (c >= '0' && c <= '9') || c == '.' || c == '-' {
				start--
			} else {
				break
			}
		}
		if start < idx {
			numStr := result[start:idx]
			v, _ := strconv.ParseFloat(numStr, 64)
			px := v * base
			replacement := ""
			if px == float64(int64(px)) {
				replacement = fmt.Sprintf("%.0fpx", px)
			} else {
				replacement = fmt.Sprintf("%.2fpx", px)
			}
			result = result[:start] + replacement + result[idx+3:]
		} else {
			result = result[:idx] + "px" + result[idx+3:]
		}
	}
	return result
}

func collectTypography(entries []tokenEntry, groups map[string]map[string]string) {
	if groups == nil {
		return
	}
	for _, e := range entries {
		tn := e.name
		for _, prefix := range []string{"family-", "size-", "weight-", "height-", "spacing-"} {
			if strings.HasPrefix(tn, prefix) {
				baseName := tn[len(prefix):]
				propName := strings.TrimSuffix(prefix, "-")
				if groups[baseName] == nil {
					groups[baseName] = make(map[string]string)
				}

				val := e.lightVal
				if e.darkVal != "" {
					val = e.darkVal
				}
				groups[baseName][propName] = val
			}
		}

		for _, suffix := range []string{"-family", "-size", "-weight", "-height", "-spacing"} {
			if strings.HasSuffix(tn, suffix) {
				baseName := tn[:len(tn)-len(suffix)]
				propName := strings.TrimPrefix(suffix, "-")
				if groups[baseName] == nil {
					groups[baseName] = make(map[string]string)
				}
				val := e.lightVal
				if e.darkVal != "" {
					val = e.darkVal
				}
				groups[baseName][propName] = val
			}
		}
	}
}

func buildTypographyToken(tree domain.DTCGGroup, name string, props map[string]string) {
	var composite domain.TypographyValue

	for _, key := range []string{"fontFamily", "fontSize", "fontWeight", "lineHeight", "letterSpacing"} {
		propName := toProp(key)
		if v, ok := props[propName]; ok {
			switch key {
			case "fontFamily":
				composite.FontFamily = v
			case "fontSize":
				composite.FontSize = v
			case "fontWeight":
				composite.FontWeight = v
			case "lineHeight":
				composite.LineHeight = v
			case "letterSpacing":
				composite.LetterSpacing = v
			}
		}
	}

	if composite == (domain.TypographyValue{}) {
		return
	}

	parts := strings.Split(name, "-")
	current := tree
	for i, part := range parts {
		if i == len(parts)-1 {
			leaf := make(domain.DTCGGroup)
			leaf["$value"] = composite
			leaf["$type"] = domain.TypeTypography
			current[part] = leaf
		} else {
			if _, ok := current[part]; !ok {
				current[part] = make(domain.DTCGGroup)
			}
			current = current[part].(domain.DTCGGroup)
		}
	}
}

func toProp(s string) string {
	switch s {
	case "fontFamily":
		return "family"
	case "fontSize":
		return "size"
	case "fontWeight":
		return "weight"
	case "lineHeight":
		return "height"
	case "letterSpacing":
		return "spacing"
	}
	return s
}

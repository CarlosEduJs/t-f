package resolve

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"text/scanner"
)

type calcLexer struct {
	scanner.Scanner
}

type calcExpr struct {
	value float64
	unit  string
}

func EvalCalc(s string) (string, error) {
	if !strings.Contains(s, "calc(") {
		return s, nil
	}

	result, err := evalCalcInner(s)
	if err != nil {
		return s, err
	}
	return result, nil
}

func evalCalcInner(s string) (string, error) {
	start := strings.Index(s, "calc(")
	if start == -1 {
		return s, nil
	}

	depth := 1
	i := start + 5
	for depth > 0 && i < len(s) {
		if s[i] == '(' {
			depth++
		} else if s[i] == ')' {
			depth--
		}
		i++
	}
	inner := s[start+5 : i-1]

	inner, _ = evalCalcInner(inner)

	expr := strings.TrimSpace(inner)
	if expr == "" {
		return s, nil
	}

	result, err := parseCalc(expr)
	if err != nil {
		return s, fmt.Errorf("calc eval error: %w", err)
	}

	replacement := formatCalcResult(result)
	before := s[:start]
	after := s[i:]

	combined := before + replacement + after
	if strings.Contains(combined, "calc(") {
		return evalCalcInner(combined)
	}

	return combined, nil
}

func parseCalc(expr string) (calcExpr, error) {
	tokens := tokenizeCalc(expr)
	if len(tokens) == 0 {
		return calcExpr{}, fmt.Errorf("empty expression")
	}

	pos := 0
	result, err := parseAddSub(tokens, &pos)
	if err != nil {
		return result, err
	}
	if pos != len(tokens) {
		return result, fmt.Errorf("unexpected tokens after expression")
	}
	return result, nil
}

type calcToken struct {
	typ   rune
	value float64
	unit  string
	op    byte
}

func tokenizeCalc(s string) []calcToken {
	var tokens []calcToken
	parts := strings.Fields(s)
	for _, p := range parts {
		switch p {
		case "+":
			tokens = append(tokens, calcToken{typ: '+', op: '+'})
		case "-":
			tokens = append(tokens, calcToken{typ: '-', op: '-'})
		case "*":
			tokens = append(tokens, calcToken{typ: '*', op: '*'})
		case "/":
			tokens = append(tokens, calcToken{typ: '/', op: '/'})
		default:
			tokens = append(tokens, parseNumberToken(p))
		}
	}
	return tokens
}

func parseNumberToken(s string) calcToken {
	s = strings.TrimSpace(s)
	if s == "" {
		return calcToken{typ: 'n', value: 0}
	}

	unitStarts := -1
	for i, c := range s {
		if c == '.' || (c >= '0' && c <= '9') {
			continue
		}
		unitStarts = i
		break
	}

	if unitStarts == -1 {
		v, _ := strconv.ParseFloat(s, 64)
		return calcToken{typ: 'n', value: v}
	}

	numPart := s[:unitStarts]
	unitPart := s[unitStarts:]

	v, err := strconv.ParseFloat(numPart, 64)
	if err != nil {
		return calcToken{typ: 'n', value: 0}
	}
	return calcToken{typ: 'n', value: v, unit: unitPart}
}

func parseAddSub(tokens []calcToken, pos *int) (calcExpr, error) {
	left, err := parseMulDiv(tokens, pos)
	if err != nil {
		return left, err
	}

	for *pos < len(tokens) {
		t := tokens[*pos]
		if t.typ != '+' && t.typ != '-' {
			break
		}
		*pos++

		right, err := parseMulDiv(tokens, pos)
		if err != nil {
			return left, err
		}

		unit := chooseUnit(left.unit, right.unit)
		if t.typ == '+' {
			left = calcExpr{value: toPx(left) + toPx(right), unit: unit}
		} else {
			left = calcExpr{value: toPx(left) - toPx(right), unit: unit}
		}
	}

	return left, nil
}

func parseMulDiv(tokens []calcToken, pos *int) (calcExpr, error) {
	left, err := parseFactor(tokens, pos)
	if err != nil {
		return left, err
	}

	for *pos < len(tokens) {
		t := tokens[*pos]
		if t.typ != '*' && t.typ != '/' {
			break
		}
		*pos++

		right, err := parseFactor(tokens, pos)
		if err != nil {
			return left, err
		}

		if t.typ == '*' {
			unit := left.unit
			if right.unit != "" {
				unit = right.unit
			}
			if right.unit != "" && left.unit != "" {
				unit = left.unit
			}
			left = calcExpr{value: left.value * right.value, unit: unit}
		} else {
			if right.value == 0 {
				return left, fmt.Errorf("division by zero")
			}
			left = calcExpr{value: left.value / right.value, unit: left.unit}
		}
	}

	return left, nil
}

func parseFactor(tokens []calcToken, pos *int) (calcExpr, error) {
	if *pos >= len(tokens) {
		return calcExpr{}, fmt.Errorf("unexpected end")
	}

	t := tokens[*pos]
	*pos++

	if t.typ == 'n' {
		return calcExpr{value: t.value, unit: t.unit}, nil
	}

	if t.op == '(' {
		expr, err := parseAddSub(tokens, pos)
		if err != nil {
			return expr, err
		}
		if *pos >= len(tokens) {
			return expr, fmt.Errorf("unmatched '('")
		}
		if tokens[*pos].op != ')' {
			return expr, fmt.Errorf("expected ')'")
		}
		*pos++
		return expr, nil
	}

	return calcExpr{}, fmt.Errorf("unexpected token")
}

func chooseUnit(a, b string) string {
	if a == b {
		return a
	}
	if a == "px" {
		return a
	}
	if b == "px" {
		return b
	}
	if a != "" {
		return a
	}
	return b
}

func toPx(e calcExpr) float64 {
	if e.unit == "rem" {
		return e.value * 16
	}
	return e.value
}

func formatCalcResult(r calcExpr) string {
	unit := r.unit
	if r.unit == "rem" {
		r.value *= 16
		unit = "px"
	}
	if r.value == math.Trunc(r.value) {
		return fmt.Sprintf("%.0f%s", r.value, unit)
	}
	return fmt.Sprintf("%.2f%s", r.value, unit)
}

func EvalCalcAll(vars map[string]string) map[string]string {
	result := make(map[string]string, len(vars))
	for name, value := range vars {
		resolved, err := EvalCalc(value)
		if err != nil {
			result[name] = value
		} else {
			result[name] = resolved
		}
	}
	return result
}

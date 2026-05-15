package domain

import "fmt"

// TokenType classifies a design token (color, dimension, typography, etc.).
type TokenType string

// Token type constants for DTCG classification.
const (
	TypeColor       TokenType = "color"
	TypeDimension   TokenType = "dimension"
	TypeFontFamily  TokenType = "fontFamily"
	TypeFontSize    TokenType = "fontSize"
	TypeFontWeight  TokenType = "fontWeight"
	TypeLineHeight  TokenType = "lineHeight"
	TypeLetterSpace TokenType = "letterSpacing"
	TypeTypography  TokenType = "typography"
	TypeBoxShadow   TokenType = "boxShadow"
	TypeString      TokenType = "string"
)

// DTCGToken is a single DTCG-format design token.
type DTCGToken struct {
	Value       any       `json:"$value"`
	Type        TokenType `json:"$type"`
	Description string    `json:"$description,omitempty"`
}

// DTCGGroup is a map node in a DTCG token tree.
type DTCGGroup map[string]any

// DesignTokens is the top-level output structure.
type DesignTokens struct {
	Semantic map[string]any `json:"semantic"`
}

// Categorize determines the category and token type from a CSS variable name.
func Categorize(name string) (category string, tokenType TokenType) {
	switch {
	case matchPrefix(name, "color"):
		return "color", TypeColor
	case matchPrefix(name, "spacing"), matchPrefix(name, "space"):
		return "spacing", TypeDimension
	case matchPrefix(name, "radius"), matchPrefix(name, "rounded"):
		return "borderRadius", TypeDimension
	case matchPrefix(name, "font-family"):
		return "typography", TypeFontFamily
	case matchPrefix(name, "font-size"):
		return "typography", TypeFontSize
	case matchPrefix(name, "font-weight"):
		return "typography", TypeFontWeight
	case matchPrefix(name, "line-height"):
		return "typography", TypeLineHeight
	case matchPrefix(name, "letter-spacing"):
		return "typography", TypeLetterSpace
	case matchPrefix(name, "font"), matchPrefix(name, "text"):
		return "typography", TypeTypography
	case matchPrefix(name, "shadow"):
		return "boxShadow", TypeBoxShadow
	default:
		return "other", TypeString
	}
}

func matchPrefix(name, prefix string) bool {
	cleaned := name
	if len(cleaned) > 0 && cleaned[0] == '-' {
		cleaned = cleaned[1:]
	}
	if len(cleaned) > 0 && cleaned[0] == '-' {
		cleaned = cleaned[1:]
	}
	if len(cleaned) >= len(prefix) && cleaned[:len(prefix)] == prefix {
		if len(cleaned) == len(prefix) || cleaned[len(prefix)] == '-' {
			return true
		}
	}
	return false
}

// CategorizeByValue infers category and type from the CSS value itself.
func CategorizeByValue(value string) (category string, tokenType TokenType) {
	if IsColorValue(value) {
		return "color", TypeColor
	}
	if IsDimensionValue(value) {
		return "spacing", TypeDimension
	}
	return "other", TypeString
}

// InferCategory tries name-based categorization first, then falls back to
// value-based inference.
func InferCategory(name, value string) (category string, tokenType TokenType) {
	cat, ttype := Categorize(name)
	if cat != "other" {
		return cat, ttype
	}
	return CategorizeByValue(value)
}

// TokenName strips leading dashes and re-prefixes with "--".
func TokenName(varName string) string {
	cleaned := varName
	if len(cleaned) > 0 && cleaned[0] == '-' {
		cleaned = cleaned[1:]
	}
	if len(cleaned) > 0 && cleaned[0] == '-' {
		cleaned = cleaned[1:]
	}
	return fmt.Sprintf("--%s", cleaned)
}

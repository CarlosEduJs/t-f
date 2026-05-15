package domain

import "fmt"

type TokenType string

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

type DTCGToken struct {
	Value      interface{} `json:"$value"`
	Type       TokenType   `json:"$type"`
	Description string     `json:"$description,omitempty"`
}

type DTCGGroup map[string]interface{}

type DesignTokens struct {
	Semantic map[string]interface{} `json:"semantic"`
}

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

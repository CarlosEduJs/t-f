package cssreader

import (
	"io"
	"strings"

	"github.com/tdewolff/parse/v2"
	"github.com/tdewolff/parse/v2/css"

	"t-f/internal/domain"
)

type Parser struct {
	input *parse.Input
}

func New(r io.Reader) *Parser {
	return &Parser{input: parse.NewInput(r)}
}

func (p *Parser) Parse() ([]domain.Variable, error) {
	parser := css.NewParser(p.input, false)

	var vars []domain.Variable
	var stack []string
	var inTheme bool

	for {
		gt, tt, data := parser.Next()
		if gt == css.ErrorGrammar {
			break
		}

		switch gt {
		case css.BeginRulesetGrammar:
			sel := selectorString(parser.Values())
			stack = append(stack, sel)

		case css.EndRulesetGrammar:
			if len(stack) > 0 {
				stack = stack[:len(stack)-1]
			}

		case css.BeginAtRuleGrammar:
			if isThemeAtRule(tt, data) {
				inTheme = true
			}

		case css.EndAtRuleGrammar:
			inTheme = false

		case css.AtRuleGrammar:
			if isThemeAtRule(tt, data) {
				vars = append(vars, extractInlineVars(parser.Values(), domain.ThemeTheme)...)
			}

		case css.CustomPropertyGrammar:
			name := string(data)
			vals := parser.Values()
			var rawValue string
			for _, v := range vals {
				rawValue += string(v.Data)
			}
			rawValue = strings.TrimSpace(rawValue)

			theme := detectTheme(stack, inTheme)
			vars = append(vars, domain.Variable{
				Name:  name,
				Value: rawValue,
				Theme: theme,
				Raw:   rawValue,
			})
		}
	}

	if err := parser.Err(); err != nil && err != io.EOF {
		return vars, err
	}
	return vars, nil
}

func selectorString(tokens []css.Token) string {
	var b strings.Builder
	for _, t := range tokens {
		if t.TokenType == css.WhitespaceToken {
			b.WriteByte(' ')
		} else {
			b.Write(t.Data)
		}
	}
	return strings.TrimSpace(b.String())
}

func isThemeAtRule(tt css.TokenType, data []byte) bool {
	return tt == css.AtKeywordToken && string(data) == "@theme"
}

func detectTheme(stack []string, inTheme bool) domain.Theme {
	if inTheme {
		return domain.ThemeTheme
	}
	for _, s := range stack {
		lower := strings.ToLower(s)
		if strings.Contains(lower, ":root") {
			return domain.ThemeLight
		}
		if strings.Contains(lower, ".dark") {
			return domain.ThemeDark
		}
	}
	return domain.ThemeLight
}

func extractInlineVars(tokens []css.Token, theme domain.Theme) []domain.Variable {
	var vars []domain.Variable
	for _, t := range tokens {
		if t.TokenType == css.CustomPropertyNameToken {
			name := string(t.Data)
			vars = append(vars, domain.Variable{
				Name:  name,
				Value: "",
				Theme: theme,
				Raw:   "",
			})
		}
	}
	return vars
}

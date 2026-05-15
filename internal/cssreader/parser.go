package cssreader

import (
	"bytes"
	"io"
	"strings"

	"github.com/tdewolff/parse/v2"
	"github.com/tdewolff/parse/v2/css"

	"t-f/internal/domain"
)

type Parser struct {
	input *parse.Input
	raw   []byte
}

func New(r io.Reader) *Parser {
	raw, _ := io.ReadAll(r)
	return &Parser{
		input: parse.NewInput(bytes.NewReader(raw)),
		raw:   raw,
	}
}

func (p *Parser) Parse() ([]domain.Variable, error) {
	parser := css.NewParser(p.input, false)

	var vars []domain.Variable
	var stack []string

	for {
		gt, _, data := parser.Next()
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

		case css.CustomPropertyGrammar:
			name := string(data)
			vals := parser.Values()
			var rawValue string
			for _, v := range vals {
				rawValue += string(v.Data)
			}
			rawValue = strings.TrimSpace(rawValue)

			theme := detectTheme(stack)
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

	themeVars := p.extractThemeBlock()
	vars = append(vars, themeVars...)

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

func detectTheme(stack []string) domain.Theme {
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

func (p *Parser) extractThemeBlock() []domain.Variable {
	text := string(p.raw)
	var vars []domain.Variable

	for {
		start := strings.Index(text, "@theme")
		if start == -1 {
			break
		}

		braceStart := strings.IndexByte(text[start:], '{')
		if braceStart == -1 {
			break
		}
		braceStart += start

		depth := 1
		pos := braceStart + 1
		for pos < len(text) && depth > 0 {
			switch text[pos] {
			case '{':
				depth++
			case '}':
				depth--
			}
			pos++
		}

		if depth != 0 {
			break
		}

		block := text[braceStart+1 : pos-1]
		lines := strings.Split(block, ";")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			idx := strings.IndexByte(line, ':')
			if idx == -1 {
				continue
			}
			name := strings.TrimSpace(line[:idx])
			value := strings.TrimSpace(line[idx+1:])
			if !strings.HasPrefix(name, "--") {
				continue
			}
			vars = append(vars, domain.Variable{
				Name:  name,
				Value: value,
				Theme: domain.ThemeTheme,
				Raw:   value,
			})
		}

		text = text[pos:]
	}

	return vars
}

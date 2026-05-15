package cssreader

import (
	"strings"
	"testing"

	"t-f/internal/domain"
)

func TestParseSimpleRoot(t *testing.T) {
	input := `:root {
		--color-primary: oklch(0.5 0.2 240);
		--spacing-md: 1rem;
	}`
	r := strings.NewReader(input)
	p := New(r)
	vars, err := p.Parse()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(vars) != 2 {
		t.Fatalf("expected 2 vars, got %d", len(vars))
	}
	if vars[0].Name != "--color-primary" || vars[0].Theme != domain.ThemeLight {
		t.Errorf("unexpected first var: %+v", vars[0])
	}
}

func TestParseDark(t *testing.T) {
	input := `:root { --color-bg: oklch(1 0 0); }
	.dark { --color-bg: oklch(0 0 0); }`
	r := strings.NewReader(input)
	p := New(r)
	vars, err := p.Parse()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(vars) != 2 {
		t.Fatalf("expected 2 vars, got %d", len(vars))
	}
	if vars[0].Theme != domain.ThemeLight || vars[1].Theme != domain.ThemeDark {
		t.Errorf("wrong theme detection")
	}
}

func TestParseThemeBlock(t *testing.T) {
	input := `@theme {
		--color-accent: oklch(0.55 0.25 280);
		--spacing-page: 1.5rem;
	}`
	r := strings.NewReader(input)
	p := New(r)
	vars, err := p.Parse()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(vars) != 2 {
		t.Fatalf("expected 2 vars, got %d", len(vars))
	}
	for _, v := range vars {
		if v.Theme != domain.ThemeTheme {
			t.Errorf("expected ThemeTheme, got %s", v.Theme)
		}
	}
}

func TestParseMixed(t *testing.T) {
	input := `:root { --color: oklch(0.5 0 0); }
	.dark { --color: oklch(0.9 0 0); }
	@theme { --color-accent: oklch(0.6 0 0); }`
	r := strings.NewReader(input)
	p := New(r)
	vars, err := p.Parse()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(vars) != 3 {
		t.Fatalf("expected 3 vars, got %d", len(vars))
	}

	themes := map[domain.Theme]int{}
	for _, v := range vars {
		themes[v.Theme]++
	}
	if themes[domain.ThemeLight] != 1 || themes[domain.ThemeDark] != 1 || themes[domain.ThemeTheme] != 1 {
		t.Errorf("unexpected theme distribution: %v", themes)
	}
}

func TestValues(t *testing.T) {
	input := `:root {
		--color: oklch(0.5 0.2 240 / 0.8);
		--size: calc(4 * 1rem);
	}`
	r := strings.NewReader(input)
	p := New(r)
	vars, err := p.Parse()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(vars) != 2 {
		t.Fatalf("expected 2 vars, got %d", len(vars))
	}
	if vars[0].Value != "oklch(0.5 0.2 240 / 0.8)" {
		t.Errorf("unexpected value: %q", vars[0].Value)
	}
	if vars[1].Value != "calc(4 * 1rem)" {
		t.Errorf("unexpected value: %q", vars[1].Value)
	}
}

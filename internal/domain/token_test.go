package domain

import (
	"testing"
)

func TestInferCategoryByName(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		wantCat  string
		wantType TokenType
	}{
		{"--color-primary", "oklch(0.5 0.2 240)", "color", TypeColor},
		{"--spacing-md", "1rem", "spacing", TypeDimension},
		{"--radius-lg", "8px", "borderRadius", TypeDimension},
		{"--font-size-base", "16px", "typography", TypeFontSize},
		{"--shadow-lg", "0 1px 2px black", "boxShadow", TypeBoxShadow},
	}
	for _, tt := range tests {
		cat, typ := InferCategory(tt.name, tt.value)
		if cat != tt.wantCat || typ != tt.wantType {
			t.Errorf("InferCategory(%q, %q) = (%q, %q), want (%q, %q)",
				tt.name, tt.value, cat, typ, tt.wantCat, tt.wantType)
		}
	}
}

func TestInferCategoryByValue(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		wantCat  string
		wantType TokenType
	}{
		{"--background", "oklch(0.2 0.1 240)", "color", TypeColor},
		{"--foreground", "oklch(0.9 0.01 260)", "color", TypeColor},
		{"--primary", "oklch(0.5 0.2 240)", "color", TypeColor},
		{"--card", "oklch(1 0 0)", "color", TypeColor},
		{"--sidebar", "oklch(0.15 0.02 260)", "color", TypeColor},
		{"--gap", "16px", "spacing", TypeDimension},
		{"--size", "1.5rem", "spacing", TypeDimension},
		{"--pad", "calc(4 * 1rem)", "spacing", TypeDimension},
		{"--unknown", "some-string", "other", TypeString},
	}
	for _, tt := range tests {
		cat, typ := InferCategory(tt.name, tt.value)
		if cat != tt.wantCat || typ != tt.wantType {
			t.Errorf("InferCategory(%q, %q) = (%q, %q), want (%q, %q)",
				tt.name, tt.value, cat, typ, tt.wantCat, tt.wantType)
		}
	}
}

func TestIsColorValue(t *testing.T) {
	tests := []struct {
		value string
		want  bool
	}{
		{"oklch(0.5 0.2 240)", true},
		{"oklch(0.5 0.2 240 / 0.8)", true},
		{"rgb(255 0 0)", true},
		{"rgba(255 0 0 / 0.5)", true},
		{"hsl(0 100% 50%)", true},
		{"hsla(0 100% 50% / 0.5)", true},
		{"#ff0000", true},
		{"#fff", true},
		{"16px", false},
		{"1.5rem", false},
		{"some-string", false},
	}
	for _, tt := range tests {
		got := IsColorValue(tt.value)
		if got != tt.want {
			t.Errorf("IsColorValue(%q) = %v, want %v", tt.value, got, tt.want)
		}
	}
}

func TestIsDimensionValue(t *testing.T) {
	tests := []struct {
		value string
		want  bool
	}{
		{"16px", true},
		{"1.5rem", true},
		{"calc(4 * 1rem)", true},
		{"oklch(0.5 0.2 240)", false},
		{"#ff0000", false},
	}
	for _, tt := range tests {
		got := IsDimensionValue(tt.value)
		if got != tt.want {
			t.Errorf("IsDimensionValue(%q) = %v, want %v", tt.value, got, tt.want)
		}
	}
}

func TestParseOKLCH(t *testing.T) {
	tests := []struct {
		input  string
		wantOK bool
		wantL  float64
		wantC  float64
		wantH  float64
		wantA  float64
	}{
		{"oklch(0.5 0.2 240)", true, 0.5, 0.2, 240, 1.0},
		{"oklch(0.5 0.2 240 / 0.8)", true, 0.5, 0.2, 240, 0.8},
		{"rgb(255 0 0)", false, 0, 0, 0, 0},
		{"oklch(abc def)", false, 0, 0, 0, 0},
	}
	for _, tt := range tests {
		c, ok := ParseOKLCH(tt.input)
		if ok != tt.wantOK {
			t.Errorf("ParseOKLCH(%q) ok = %v, want %v", tt.input, ok, tt.wantOK)
		}
		if ok && (c.L != tt.wantL || c.C != tt.wantC || c.H != tt.wantH || c.Alpha != tt.wantA) {
			t.Errorf("ParseOKLCH(%q) = (%f,%f,%f,%f), want (%f,%f,%f,%f)",
				tt.input, c.L, c.C, c.H, c.Alpha, tt.wantL, tt.wantC, tt.wantH, tt.wantA)
		}
	}
}

func TestOKLCHToHEX(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"oklch(0 0 0)", "#000000"},
		{"oklch(1 0 0)", "#ffffff"},
		{"oklch(0.5 0 0)", "#636363"},
	}
	for _, tt := range tests {
		got := ConvertColorToHEX(tt.input)
		if got != tt.want {
			t.Errorf("ConvertColorToHEX(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestConvertColorToHEX(t *testing.T) {
	hex := ConvertColorToHEX("oklch(0.5 0.2 240)")
	if hex == "" || hex[0] != '#' {
		t.Errorf("expected hex string starting with #, got %q", hex)
	}
	if len(hex) != 7 {
		t.Errorf("expected 7-char hex, got %q (len=%d)", hex, len(hex))
	}

	identity := ConvertColorToHEX("#ff0000")
	if identity != "#ff0000" {
		t.Errorf("expected identity for hex input, got %q", identity)
	}
}

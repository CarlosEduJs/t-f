package tokens

import (
	"encoding/json"
	"testing"

	"t-f/internal/domain"
)

func TestCategorize(t *testing.T) {
	tests := []struct {
		name         string
		wantCategory string
		wantType     domain.TokenType
	}{
		{"--color-primary", "color", domain.TypeColor},
		{"--spacing-md", "spacing", domain.TypeDimension},
		{"--radius-sm", "borderRadius", domain.TypeDimension},
		{"--font-family-body", "typography", domain.TypeFontFamily},
		{"--font-size-base", "typography", domain.TypeFontSize},
		{"--font-weight-bold", "typography", domain.TypeFontWeight},
		{"--shadow-lg", "boxShadow", domain.TypeBoxShadow},
		{"--unknown-prop", "other", domain.TypeString},
	}
	for _, tt := range tests {
		cat, typ := domain.Categorize(tt.name)
		if cat != tt.wantCategory || typ != tt.wantType {
			t.Errorf("Categorize(%q) = (%q, %q), want (%q, %q)",
				tt.name, cat, typ, tt.wantCategory, tt.wantType)
		}
	}
}

func TestSplitByTheme(t *testing.T) {
	vars := []domain.Variable{
		{Name: "--color-a", Value: "red", Theme: domain.ThemeLight},
		{Name: "--color-a", Value: "blue", Theme: domain.ThemeDark},
		{Name: "--color-b", Value: "green", Theme: domain.ThemeTheme},
	}
	themed := splitByTheme(vars)
	if len(themed["light"]) != 1 || len(themed["dark"]) != 1 || len(themed["theme"]) != 1 {
		t.Errorf("unexpected split: light=%d dark=%d theme=%d",
			len(themed["light"]), len(themed["dark"]), len(themed["theme"]))
	}
}

func TestGenerateMinimal(t *testing.T) {
	vars := []domain.Variable{
		{Name: "--color-primary", Value: "oklch(0.5 0.2 240)", Theme: domain.ThemeLight},
		{Name: "--color-primary", Value: "oklch(0.7 0.2 240)", Theme: domain.ThemeDark},
	}
	gen := NewGenerator()
	data, err := gen.Generate(vars)
	if err != nil {
		t.Fatalf("Generate error: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	semantic := result["semantic"].(map[string]interface{})
	colors := semantic["color"].(map[string]interface{})
	primary := colors["primary"].(map[string]interface{})
	if primary["$type"] != "color" {
		t.Errorf("expected color type")
	}
	if _, ok := primary["dark"]; !ok {
		t.Errorf("expected dark variant")
	}
}

func TestGenerateTypography(t *testing.T) {
	vars := []domain.Variable{
		{Name: "--font-family-body", Value: "Inter, sans-serif", Theme: domain.ThemeLight},
		{Name: "--font-size-body", Value: "16px", Theme: domain.ThemeLight},
		{Name: "--font-weight-body", Value: "400", Theme: domain.ThemeLight},
	}
	gen := NewGenerator()
	data, err := gen.Generate(vars)
	if err != nil {
		t.Fatalf("Generate error: %v", err)
	}

	var result map[string]interface{}
	json.Unmarshal(data, &result)

	semantic := result["semantic"].(map[string]interface{})
	typo := semantic["typography"].(map[string]interface{})
	body := typo["body"].(map[string]interface{})

	if body["$type"] != "typography" {
		t.Errorf("expected typography type, got %v", body["$type"])
	}
	val := body["$value"].(map[string]interface{})
	if val["fontFamily"] != "Inter, sans-serif" {
		t.Errorf("expected fontFamily")
	}
	if val["fontSize"] != "16px" {
		t.Errorf("expected fontSize")
	}
}

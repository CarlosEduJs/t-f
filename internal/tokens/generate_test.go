package tokens

import (
	"encoding/json"
	"testing"

	"t-f/internal/domain"
)

func checkMixed(obj map[string]any, path string, t *testing.T) {
	hasLeaf := false
	hasChild := false
	for k := range obj {
		if k == "$value" || k == "$type" || k == "$description" {
			hasLeaf = true
		} else if k != "dark" {
			if _, ok := obj[k].(map[string]any); ok {
				hasChild = true
			}
		}
	}
	if hasLeaf && hasChild {
		t.Errorf("MIXED leaf/group at %s: keys=%v", path, keysOf(obj))
	}
	for k, v := range obj {
		if sub, ok := v.(map[string]any); ok {
			checkMixed(sub, path+"."+k, t)
		}
	}
}

func keysOf(m map[string]any) []string {
	ks := make([]string, 0, len(m))
	for k := range m {
		ks = append(ks, k)
	}
	return ks
}

func getHex(val any) string {
	m, ok := val.(map[string]any)
	if !ok {
		return ""
	}
	h, _ := m["hex"].(string)
	return h
}

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

	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	semantic := result["semantic"].(map[string]any)
	checkMixed(semantic, "semantic", t)

	colors := semantic["color"].(map[string]any)
	primary := colors["primary"].(map[string]any)
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

	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	semantic := result["semantic"].(map[string]any)
	checkMixed(semantic, "semantic", t)

	typo := semantic["typography"].(map[string]any)
	body := typo["body"].(map[string]any)

	if body["$type"] != "typography" {
		t.Errorf("expected typography type, got %v", body["$type"])
	}
	val := body["$value"].(map[string]any)
	if val["fontFamily"] != "Inter, sans-serif" {
		t.Errorf("expected fontFamily")
	}
	if val["fontSize"] != "16px" {
		t.Errorf("expected fontSize")
	}
}

func TestSemanticFirstNaming(t *testing.T) {
	vars := []domain.Variable{
		{Name: "--background", Value: "oklch(0.98 0 0)", Theme: domain.ThemeLight},
		{Name: "--foreground", Value: "oklch(0.15 0.02 260)", Theme: domain.ThemeLight},
		{Name: "--primary", Value: "oklch(0.5 0.2 240)", Theme: domain.ThemeLight},
		{Name: "--background", Value: "oklch(0.12 0.02 260)", Theme: domain.ThemeDark},
	}
	gen := NewGenerator()
	data, err := gen.Generate(vars)
	if err != nil {
		t.Fatalf("Generate error: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	semantic := result["semantic"].(map[string]any)
	checkMixed(semantic, "semantic", t)

	if _, hasOther := semantic["other"]; hasOther {
		t.Errorf("'other' category should not appear; semantic tokens must be inferred by value")
	}

	colors, hasColor := semantic["color"].(map[string]any)
	if !hasColor {
		t.Fatalf("expected 'color' category for semantic-first tokens")
	}

	for _, name := range []string{"background", "foreground", "primary"} {
		if _, ok := colors[name]; !ok {
			t.Errorf("expected color.%s token", name)
		}
	}

	bg := colors["background"].(map[string]any)
	if _, hasDark := bg["dark"]; !hasDark {
		t.Errorf("expected dark variant for background")
	}
}

func TestNoDuplicateEmission(t *testing.T) {
	vars := []domain.Variable{
		{Name: "--background", Value: "oklch(0.98 0 0)", Theme: domain.ThemeLight},
		{Name: "--color-primary", Value: "oklch(0.5 0.2 240)", Theme: domain.ThemeLight},
	}
	gen := NewGenerator()
	data, err := gen.Generate(vars)
	if err != nil {
		t.Fatalf("Generate error: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	semantic := result["semantic"].(map[string]any)
	checkMixed(semantic, "semantic", t)

	if _, hasOther := semantic["other"]; hasOther {
		t.Errorf("no token should appear in 'other' category")
	}

	colors := semantic["color"].(map[string]any)
	if _, ok := colors["background"]; !ok {
		t.Errorf("background should be in color category")
	}
	if _, ok := colors["primary"]; !ok {
		t.Errorf("primary should be in color category")
	}
}

func TestFigmaMode(t *testing.T) {
	vars := []domain.Variable{
		{Name: "--background", Value: "oklch(0.98 0 0)", Theme: domain.ThemeLight},
		{Name: "--primary", Value: "oklch(0.5 0.2 240)", Theme: domain.ThemeLight},
		{Name: "--primary", Value: "oklch(0.7 0.2 240)", Theme: domain.ThemeDark},
	}
	gen := NewGenerator()
	gen.FigmaMode = true
	data, err := gen.Generate(vars)
	if err != nil {
		t.Fatalf("Generate error: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	semantic := result["semantic"].(map[string]any)
	checkMixed(semantic, "semantic", t)

	colors := semantic["color"].(map[string]any)

	if _, ok := colors["background"]; !ok {
		t.Errorf("expected color.background")
	}
	if _, ok := colors["primary"]; !ok {
		t.Errorf("expected color.primary")
	}
	if _, ok := colors["primary-dark"]; !ok {
		t.Errorf("expected color.primary-dark in flat figma mode")
	}

	for _, name := range []string{"background", "primary", "primary-dark"} {
		token := colors[name].(map[string]any)
		if len(token) != 2 {
			t.Errorf("figma token %q should have exactly 2 keys ($type, $value), got %v", name, keysOf(token))
		}
		hex := getHex(token["$value"])
		if hex == "" {
			t.Errorf("figma mode: expected structured color value for %s, got %T=%v", name, token["$value"], token["$value"])
		}
		if token["$type"] != "color" {
			t.Errorf("figma mode: expected color type for %s", name)
		}
	}
}

func TestDefaultModePreservesOKLCH(t *testing.T) {
	vars := []domain.Variable{
		{Name: "--background", Value: "oklch(0.98 0 0)", Theme: domain.ThemeLight},
	}
	gen := NewGenerator()
	gen.FigmaMode = false
	data, err := gen.Generate(vars)
	if err != nil {
		t.Fatalf("Generate error: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	semantic := result["semantic"].(map[string]any)
	colors := semantic["color"].(map[string]any)
	bg := colors["background"].(map[string]any)

	if bg["$value"] != "oklch(0.98 0 0)" {
		t.Errorf("default mode should preserve OKLCH, got %q", bg["$value"])
	}
}

func TestNoMixedLeafGroupDefault(t *testing.T) {
	vars := []domain.Variable{
		{Name: "--color-primary", Value: "oklch(0.5 0.2 240)", Theme: domain.ThemeLight},
		{Name: "--color-primary-hover", Value: "oklch(0.6 0.2 240)", Theme: domain.ThemeLight},
		{Name: "--color-primary", Value: "oklch(0.7 0.2 240)", Theme: domain.ThemeDark},
		{Name: "--color-primary-hover", Value: "oklch(0.8 0.2 240)", Theme: domain.ThemeDark},
	}
	gen := NewGenerator()
	data, err := gen.Generate(vars)
	if err != nil {
		t.Fatalf("Generate error: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	semantic := result["semantic"].(map[string]any)
	checkMixed(semantic, "semantic", t)

	colors := semantic["color"].(map[string]any)
	primary := colors["primary"].(map[string]any)

	if _, ok := primary["$value"]; ok {
		t.Errorf("primary should NOT have $value when it has children")
	}
	if _, ok := primary["base"]; !ok {
		t.Errorf("primary should have 'base' child for its value")
	}
	if _, ok := primary["hover"]; !ok {
		t.Errorf("primary should have 'hover' child")
	}
}

func TestNoMixedLeafGroupFigma(t *testing.T) {
	vars := []domain.Variable{
		{Name: "--color-primary", Value: "oklch(0.5 0.2 240)", Theme: domain.ThemeLight},
		{Name: "--color-primary-hover", Value: "oklch(0.6 0.2 240)", Theme: domain.ThemeLight},
		{Name: "--color-primary", Value: "oklch(0.7 0.2 240)", Theme: domain.ThemeDark},
		{Name: "--color-primary-hover", Value: "oklch(0.8 0.2 240)", Theme: domain.ThemeDark},
	}
	gen := NewGenerator()
	gen.FigmaMode = true
	data, err := gen.Generate(vars)
	if err != nil {
		t.Fatalf("Generate error: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	semantic := result["semantic"].(map[string]any)
	checkMixed(semantic, "semantic", t)

	colors := semantic["color"].(map[string]any)

	expected := []string{"primary", "primary-dark", "primary-hover", "primary-hover-dark"}
	for _, name := range expected {
		if _, ok := colors[name]; !ok {
			t.Errorf("expected flat token color.%s", name)
		}
	}

	for name, token := range colors {
		m := token.(map[string]any)
		if len(m) != 2 {
			t.Errorf("figma token %q should have exactly 2 keys ($type, $value), got %v", name, keysOf(m))
		}
		if _, ok := m["$value"]; !ok {
			t.Errorf("figma token %q missing $value", name)
		}
		if _, ok := m["$type"]; !ok {
			t.Errorf("figma token %q missing $type", name)
		}
	}
}

func TestFigmaDarkTokenFlattening(t *testing.T) {
	vars := []domain.Variable{
		{Name: "--color-primary", Value: "oklch(0.5 0.2 240)", Theme: domain.ThemeLight},
		{Name: "--color-primary-hover", Value: "oklch(0.6 0.2 240)", Theme: domain.ThemeLight},
		{Name: "--color-primary", Value: "oklch(0.7 0.2 240)", Theme: domain.ThemeDark},
		{Name: "--color-primary-hover", Value: "oklch(0.8 0.2 240)", Theme: domain.ThemeDark},
	}
	gen := NewGenerator()
	gen.FigmaMode = true
	data, err := gen.Generate(vars)
	if err != nil {
		t.Fatalf("Generate error: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	semantic := result["semantic"].(map[string]any)
	colors := semantic["color"].(map[string]any)

	tests := []struct {
		name     string
		wantHex  string
		wantType string
	}{
		{"primary", "#0069c7", "color"},
		{"primary-dark", "#00a9ff", "color"},
		{"primary-hover", "#0089e9", "color"},
		{"primary-hover-dark", "#00caff", "color"},
	}
	for _, tt := range tests {
		token, ok := colors[tt.name].(map[string]any)
		if !ok {
			t.Errorf("missing token: %s", tt.name)
			continue
		}
		if token["$type"] != tt.wantType {
			t.Errorf("%s $type = %q, want %q", tt.name, token["$type"], tt.wantType)
		}
		gotHex := getHex(token["$value"])
		if gotHex != tt.wantHex {
			t.Errorf("%s hex = %q, want %q", tt.name, gotHex, tt.wantHex)
		}
	}
}

func TestFigmaStructuredValues(t *testing.T) {
	vars := []domain.Variable{
		{Name: "--color-primary", Value: "oklch(0.5 0.2 240)", Theme: domain.ThemeLight},
		{Name: "--radius-md", Value: "12px", Theme: domain.ThemeLight},
	}
	gen := NewGenerator()
	gen.FigmaMode = true
	data, err := gen.Generate(vars)
	if err != nil {
		t.Fatalf("Generate error: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	semantic := result["semantic"].(map[string]any)

	colors := semantic["color"].(map[string]any)
	primary := colors["primary"].(map[string]any)
	cv := primary["$value"].(map[string]any)

	if cv["colorSpace"] != "srgb" {
		t.Errorf("expected colorSpace=srgb, got %q", cv["colorSpace"])
	}
	comps, ok := cv["components"].([]any)
	if !ok || len(comps) != 3 {
		t.Errorf("expected 3 components, got %v", comps)
	} else {
		for i, c := range comps {
			v, _ := c.(float64)
			if v < 0 || v > 1 {
				t.Errorf("component %d out of range [0,1]: %f", i, v)
			}
		}
	}
	if cv["hex"] != "#0069c7" {
		t.Errorf("expected hex #0069c7, got %q", cv["hex"])
	}

	borders := semantic["borderRadius"].(map[string]any)
	md := borders["md"].(map[string]any)
	dv := md["$value"].(map[string]any)

	val, _ := dv["value"].(float64)
	if val != 12 {
		t.Errorf("expected dimension value 12, got %f", val)
	}
	unit, _ := dv["unit"].(string)
	if unit != "px" {
		t.Errorf("expected dimension unit px, got %q", unit)
	}
}

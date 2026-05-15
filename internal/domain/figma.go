package domain

// FigmaColorValue represents a color in a format compatible with Figma.
type FigmaColorValue struct {
	ColorSpace string    `json:"colorSpace"`
	Components []float64 `json:"components"`
	Hex        string    `json:"hex"`
}

// FigmaDimensionValue represents a dimension in a format compatible with Figma.
type FigmaDimensionValue struct {
	Value float64 `json:"value"`
	Unit  string  `json:"unit"`
}

// TypographyValue represents a composite typography token.
type TypographyValue struct {
	FontFamily    string `json:"fontFamily,omitempty"`
	FontSize      string `json:"fontSize,omitempty"`
	FontWeight    string `json:"fontWeight,omitempty"`
	LineHeight    string `json:"lineHeight,omitempty"`
	LetterSpacing string `json:"letterSpacing,omitempty"`
}

package domain

// Theme identifies which visual mode a CSS variable belongs to.
type Theme string

// Theme constants for light, dark, and @theme scopes.
const (
	ThemeLight Theme = "light"
	ThemeDark  Theme = "dark"
	ThemeTheme Theme = "theme"
)

// Variable represents a parsed CSS custom property.
type Variable struct {
	Name  string
	Value string
	Theme Theme
	Raw   string
}

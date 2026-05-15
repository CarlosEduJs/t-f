package domain

type Theme string

const (
	ThemeLight Theme = "light"
	ThemeDark  Theme = "dark"
	ThemeTheme Theme = "theme"
)

type Variable struct {
	Name  string
	Value string
	Theme Theme
	Raw   string
}

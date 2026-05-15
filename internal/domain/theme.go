package domain

type ThemeContext struct {
	Name      string
	Variables []Variable
	Selectors []string
}

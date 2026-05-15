package domain

// ThemeContext groups variables under a named theme scope with its CSS selectors.
type ThemeContext struct {
	Name      string
	Variables []Variable
	Selectors []string
}

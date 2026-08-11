package login

// RenderHTML returns the HTML representation of the login screen.
// This is used by SSR to serve the pre-rendered markup.
// If l.Title is empty (meaning it's a zero-value instance during auto-discovery),
// it returns an empty string to avoid injecting empty login screens in every app page.
func (l *Login) RenderHTML() string {
	if l.Title == "" {
		return ""
	}
	return l.Render().String()
}

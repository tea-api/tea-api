package setting

// Default regex rules for sensitive words.
// Users may load these from configuration later.

var DefaultRegexRules = map[string][]string{
	"sensitive": {`badword`, `\d{6}password`},
}

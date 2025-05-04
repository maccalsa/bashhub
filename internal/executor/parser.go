package executor

import "regexp"

var placeholderRegexp = regexp.MustCompile(`{{\s*([^}]+)\s*}}`)

// ParsePlaceholders returns all unique placeholder names found.
func ParsePlaceholders(script string) []string {
	matches := placeholderRegexp.FindAllStringSubmatch(script, -1)
	unique := make(map[string]bool)
	var placeholders []string

	for _, match := range matches {
		name := match[1]
		if !unique[name] {
			unique[name] = true
			placeholders = append(placeholders, name)
		}
	}

	return placeholders
}

// ReplacePlaceholders replaces placeholders with user inputs.
func ReplacePlaceholders(script string, inputs map[string]string) string {
	return placeholderRegexp.ReplaceAllStringFunc(script, func(match string) string {
		nameMatch := placeholderRegexp.FindStringSubmatch(match)
		if val, exists := inputs[nameMatch[1]]; exists {
			return val
		}
		return match
	})
}

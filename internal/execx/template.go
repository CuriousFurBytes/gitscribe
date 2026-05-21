package execx

import "strings"

func ExpandArgs(template []string, replacements map[string]string) []string {
	expanded := make([]string, len(template))
	for i, part := range template {
		for placeholder, value := range replacements {
			part = strings.ReplaceAll(part, placeholder, value)
		}
		expanded[i] = part
	}
	return expanded
}

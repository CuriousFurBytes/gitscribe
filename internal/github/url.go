package github

import "strings"

// ParsePRURL extracts the pull request URL from the output of `gh pr create`.
// `gh` prints the new PR URL to stdout on success, typically as the last
// non-empty line. To be tolerant of additional informational output, this
// scans every line and returns the last non-empty line that looks like an
// HTTP(S) URL. An empty string is returned when no URL is found.
func ParsePRURL(output string) string {
	lines := strings.Split(output, "\n")
	for i := len(lines) - 1; i >= 0; i-- {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "http://") || strings.HasPrefix(line, "https://") {
			return line
		}
	}
	return ""
}

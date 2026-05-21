package templates

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

var defaultTemplateCandidates = []string{
	".github/pull_request_template.md",
	".github/PULL_REQUEST_TEMPLATE.md",
	"docs/pull_request_template.md",
}

func DiscoverPRTemplate(repoRoot string, customPath string) (string, string, error) {
	candidates := append([]string{}, defaultTemplateCandidates...)
	if customPath != "" {
		candidates = append(candidates, customPath)
	}

	cleanRoot := filepath.Clean(repoRoot)
	for _, candidate := range candidates {
		path := candidate
		if !filepath.IsAbs(path) {
			path = filepath.Join(repoRoot, candidate)
		}

		cleanPath := filepath.Clean(path)
		if !strings.HasPrefix(cleanPath, cleanRoot+string(filepath.Separator)) {
			return "", "", fmt.Errorf("template path %q is outside the repository root", candidate)
		}

		info, err := os.Stat(path)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return "", "", fmt.Errorf("stat template %s: %w", path, err)
		}
		if info.IsDir() {
			continue
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return "", "", fmt.Errorf("read template %s: %w", path, err)
		}
		return path, string(content), nil
	}

	return "", "", nil
}

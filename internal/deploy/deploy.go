package deploy

import (
	"fmt"
	"os"
	"strings"
)

// Known providers for .gosupabase.yaml and setup import.
const (
	ProviderVercel   = "vercel"
	ProviderFly      = "fly"
	ProviderRailway  = "railway"
	ProviderRender   = "render"
	ProviderNone     = "none"
)

// NormalizeProvider maps user input to a canonical provider id.
func NormalizeProvider(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	switch s {
	case "vercel", "v", "1":
		return ProviderVercel
	case "fly", "fly.io", "flyio", "2":
		return ProviderFly
	case "railway", "rail", "3":
		return ProviderRailway
	case "render", "4":
		return ProviderRender
	case "none", "local", "skip", "", "5", "0":
		return ProviderNone
	default:
		return ProviderNone
	}
}

// ShouldOverwrite decides whether to replace an existing file.
type ShouldOverwrite func(relPath string) bool

// WriteFiles writes provider-specific deploy config files. Existing files are
// skipped unless shouldOverwrite returns true for that path.
func WriteFiles(provider string, flyAppName string, shouldOverwrite ShouldOverwrite) (written, skipped []string, err error) {
	if shouldOverwrite == nil {
		shouldOverwrite = func(string) bool { return false }
	}
	provider = NormalizeProvider(provider)
	if provider == ProviderNone {
		return nil, nil, nil
	}

	type pair struct {
		name    string
		content string
	}
	var pairs []pair

	switch provider {
	case ProviderVercel:
		pairs = []pair{{"vercel.json", vercelJSON}}
	case ProviderFly:
		app := strings.TrimSpace(flyAppName)
		if app == "" {
			app = "my-gosupabase-app"
		}
		pairs = []pair{{"fly.toml", flyTOML(app)}}
	case ProviderRailway:
		pairs = []pair{{"railway.toml", railwayTOML}}
	case ProviderRender:
		pairs = []pair{{"render.yaml", renderYAML}}
	default:
		return nil, nil, nil
	}

	for _, p := range pairs {
		if _, statErr := os.Stat(p.name); statErr == nil && !shouldOverwrite(p.name) {
			skipped = append(skipped, p.name)
			continue
		}
		if err := os.WriteFile(p.name, []byte(p.content), 0644); err != nil {
			return written, skipped, fmt.Errorf("write %s: %w", p.name, err)
		}
		written = append(written, p.name)
	}

	// Single markdown with next steps (skip if exists and not overwriting)
	const docName = "DEPLOY.md"
	if _, statErr := os.Stat(docName); statErr == nil && !shouldOverwrite(docName) {
		skipped = append(skipped, docName)
	} else {
		body := deployMarkdown(provider)
		if err := os.WriteFile(docName, []byte(body), 0644); err != nil {
			return written, skipped, fmt.Errorf("write %s: %w", docName, err)
		}
		written = append(written, docName)
	}

	return written, skipped, nil
}

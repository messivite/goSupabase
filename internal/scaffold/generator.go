package scaffold

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	yamlcfg "github.com/messivite/gosupabase/internal/yaml"
)

type GenerateOptions struct {
	HandlersDir  string
	ServerDir    string
	HandlersOnly bool
	Module       string
}

func Generate(apiPath string, opts GenerateOptions) error {
	cfg, err := yamlcfg.Load(apiPath)
	if err != nil {
		return fmt.Errorf("loading api.yaml: %w", err)
	}

	if err := os.MkdirAll(opts.HandlersDir, 0755); err != nil {
		return fmt.Errorf("creating handlers dir: %w", err)
	}

	registryPath := filepath.Join(opts.HandlersDir, "registry.go")
	if _, err := os.Stat(registryPath); os.IsNotExist(err) {
		if err := os.WriteFile(registryPath, []byte(registryTemplate), 0644); err != nil {
			return fmt.Errorf("writing registry.go: %w", err)
		}
		fmt.Printf("  created %s\n", registryPath)
	}

	for _, ep := range cfg.Endpoints {
		name := ep.Handler
		if name == "" {
			name = yamlcfg.DeriveHandlerName(ep.Method, ep.Path)
		}
		filename := toSnakeCase(name) + ".go"
		handlerPath := filepath.Join(opts.HandlersDir, filename)

		if _, err := os.Stat(handlerPath); err == nil {
			fmt.Printf("  exists  %s (skipped)\n", handlerPath)
			continue
		}

		content, err := renderTemplate(handlerTemplate, map[string]string{"Name": name})
		if err != nil {
			return fmt.Errorf("rendering handler %s: %w", name, err)
		}
		if err := os.WriteFile(handlerPath, []byte(content), 0644); err != nil {
			return fmt.Errorf("writing %s: %w", handlerPath, err)
		}
		fmt.Printf("  created %s\n", handlerPath)
	}

	if opts.HandlersOnly {
		fmt.Println("  --handlers-only: skipping server generation")
		return nil
	}

	if err := os.MkdirAll(opts.ServerDir, 0755); err != nil {
		return fmt.Errorf("creating server dir: %w", err)
	}

	module := opts.Module
	if module == "" {
		module = detectModule()
	}

	serverPath := filepath.Join(opts.ServerDir, "server.go")
	content, err := renderTemplate(serverTemplate, map[string]string{"Module": module})
	if err != nil {
		return fmt.Errorf("rendering server.go: %w", err)
	}
	if err := os.WriteFile(serverPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("writing server.go: %w", err)
	}
	fmt.Printf("  created %s\n", serverPath)

	return nil
}

// ScaffoldNew creates a full project skeleton in the given directory.
func ScaffoldNew(dir, module string) error {
	dirs := []string{
		"cmd/gosupabase",
		"cmd/server",
		"internal/yaml",
		"internal/scaffold",
		"middleware",
		"auth",
		"handlers",
		"server",
		"config",
	}
	for _, d := range dirs {
		if err := os.MkdirAll(filepath.Join(dir, d), 0755); err != nil {
			return err
		}
	}

	files := map[string]string{
		"api.yaml":     apiYAMLTemplate,
		"Makefile":     makefileTemplate,
		".env.example": "PORT=8080\nSUPABASE_URL=\nSUPABASE_ANON_KEY=\nSUPABASE_SERVICE_ROLE_KEY=\nSUPABASE_JWT_SECRET=\nSUPABASE_JWT_VALIDATION_MODE=auto\n",
	}
	for name, content := range files {
		path := filepath.Join(dir, name)
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			return err
		}
		fmt.Printf("  created %s\n", path)
	}

	goMod := fmt.Sprintf("module %s\n\ngo 1.22\n", module)
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte(goMod), 0644); err != nil {
		return err
	}
	fmt.Printf("  created go.mod\n")

	mainContent, err := renderTemplate(mainServerTemplate, map[string]string{"Module": module})
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(dir, "cmd/server/main.go"), []byte(mainContent), 0644); err != nil {
		return err
	}
	fmt.Printf("  created cmd/server/main.go\n")

	return nil
}

func renderTemplate(tmpl string, data map[string]string) (string, error) {
	t, err := template.New("").Parse(tmpl)
	if err != nil {
		return "", err
	}
	var sb strings.Builder
	if err := t.Execute(&sb, data); err != nil {
		return "", err
	}
	return sb.String(), nil
}

func toSnakeCase(s string) string {
	var result strings.Builder
	for i, r := range s {
		if r >= 'A' && r <= 'Z' {
			if i > 0 {
				result.WriteByte('_')
			}
			result.WriteRune(r + 32)
		} else {
			result.WriteRune(r)
		}
	}
	return result.String()
}

func detectModule() string {
	data, err := os.ReadFile("go.mod")
	if err != nil {
		return "github.com/example/myapp"
	}
	for _, line := range strings.Split(string(data), "\n") {
		if strings.HasPrefix(line, "module ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "module "))
		}
	}
	return "github.com/example/myapp"
}

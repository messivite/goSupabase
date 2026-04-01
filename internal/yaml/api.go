package yamlcfg

import (
	"fmt"
	"os"
	"strings"
	"unicode"

	"gopkg.in/yaml.v3"
)

type APIConfig struct {
	Version   string       `yaml:"version"`
	BasePath  string       `yaml:"basePath"`
	Output    OutputConfig `yaml:"output"`
	Endpoints []Endpoint   `yaml:"endpoints"`
}

type OutputConfig struct {
	ServerDir   string `yaml:"serverDir"`
	HandlersDir string `yaml:"handlersDir"`
}

type Endpoint struct {
	Method  string   `yaml:"method"`
	Path    string   `yaml:"path"`
	Handler string   `yaml:"handler"`
	Auth    bool     `yaml:"auth"`
	Roles   []string `yaml:"roles,omitempty"`
}

func Load(path string) (*APIConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %w", path, err)
	}
	var cfg APIConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse %s: %w", path, err)
	}
	return &cfg, nil
}

func Save(path string, cfg *APIConfig) error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write %s: %w", path, err)
	}
	return nil
}

func (cfg *APIConfig) AddEndpoint(ep Endpoint) error {
	for _, existing := range cfg.Endpoints {
		if strings.EqualFold(existing.Method, ep.Method) && existing.Path == ep.Path {
			return fmt.Errorf("endpoint %s %s already exists", ep.Method, ep.Path)
		}
	}
	if ep.Handler == "" {
		ep.Handler = DeriveHandlerName(ep.Method, ep.Path)
	}
	cfg.Endpoints = append(cfg.Endpoints, ep)
	return nil
}

func (cfg *APIConfig) ListEndpoints() []Endpoint {
	return cfg.Endpoints
}

// DeriveHandlerName builds a PascalCase handler name from method + path.
// POST /tracks -> PostTracks, GET /tracks/:id -> GetTracksById
func DeriveHandlerName(method, path string) string {
	method = strings.ToUpper(method)
	prefix := capitalize(strings.ToLower(method))

	path = strings.TrimPrefix(path, "/")
	segments := strings.Split(path, "/")

	var sb strings.Builder
	sb.WriteString(prefix)

	for _, seg := range segments {
		if seg == "" {
			continue
		}
		if strings.HasPrefix(seg, ":") {
			param := strings.TrimPrefix(seg, ":")
			sb.WriteString("By")
			sb.WriteString(capitalize(param))
		} else {
			sb.WriteString(capitalize(seg))
		}
	}
	return sb.String()
}

// NormalizePath converts :param to {param} for chi router compatibility.
func NormalizePath(path string) string {
	segments := strings.Split(path, "/")
	for i, seg := range segments {
		if strings.HasPrefix(seg, ":") {
			segments[i] = "{" + strings.TrimPrefix(seg, ":") + "}"
		}
	}
	return strings.Join(segments, "/")
}

func capitalize(s string) string {
	if s == "" {
		return s
	}
	runes := []rune(s)
	runes[0] = unicode.ToUpper(runes[0])
	return string(runes)
}

package config

import (
	"bufio"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Port                   string
	SupabaseURL            string
	SupabaseAnonKey        string
	SupabaseServiceRoleKey string
	SupabaseJWTSecret      string
	SupabaseJWTValidationMode string
}

type OutputPaths struct {
	ServerDir   string
	HandlersDir string
}

func LoadEnv() *Config {
	loadDotEnv(".env")

	c := &Config{
		Port:                   os.Getenv("PORT"),
		SupabaseURL:            os.Getenv("SUPABASE_URL"),
		SupabaseAnonKey:        os.Getenv("SUPABASE_ANON_KEY"),
		SupabaseServiceRoleKey: os.Getenv("SUPABASE_SERVICE_ROLE_KEY"),
		SupabaseJWTSecret:      os.Getenv("SUPABASE_JWT_SECRET"),
		SupabaseJWTValidationMode: os.Getenv("SUPABASE_JWT_VALIDATION_MODE"),
	}
	if c.Port == "" {
		c.Port = "8080"
	}
	if c.SupabaseJWTValidationMode == "" {
		c.SupabaseJWTValidationMode = "auto"
	}
	return c
}

// ResolveOutputPaths applies precedence: flags > .gosupabase.yaml > api.yaml output > defaults.
func ResolveOutputPaths(flagServer, flagHandlers string) *OutputPaths {
	out := &OutputPaths{ServerDir: "server", HandlersDir: "handlers"}

	if data, err := os.ReadFile("api.yaml"); err == nil {
		var parsed struct {
			Output struct {
				ServerDir   string `yaml:"serverDir"`
				HandlersDir string `yaml:"handlersDir"`
			} `yaml:"output"`
		}
		if yaml.Unmarshal(data, &parsed) == nil {
			if parsed.Output.ServerDir != "" {
				out.ServerDir = parsed.Output.ServerDir
			}
			if parsed.Output.HandlersDir != "" {
				out.HandlersDir = parsed.Output.HandlersDir
			}
		}
	}

	if data, err := os.ReadFile(".gosupabase.yaml"); err == nil {
		var parsed struct {
			Output struct {
				ServerDir   string `yaml:"serverDir"`
				HandlersDir string `yaml:"handlersDir"`
			} `yaml:"output"`
		}
		if yaml.Unmarshal(data, &parsed) == nil {
			if parsed.Output.ServerDir != "" {
				out.ServerDir = parsed.Output.ServerDir
			}
			if parsed.Output.HandlersDir != "" {
				out.HandlersDir = parsed.Output.HandlersDir
			}
		}
	}

	if flagServer != "" {
		out.ServerDir = flagServer
	}
	if flagHandlers != "" {
		out.HandlersDir = flagHandlers
	}
	return out
}

// ParseEnvFile reads a KEY=VALUE file and returns the entries as a map.
// Skips blank lines and comments (#). Used by both runtime loading and setup --from-file.
func ParseEnvFile(path string) (map[string]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	result := make(map[string]string)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		k, v, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		result[strings.TrimSpace(k)] = strings.TrimSpace(v)
	}
	return result, scanner.Err()
}

func loadDotEnv(path string) {
	entries, err := ParseEnvFile(path)
	if err != nil {
		return
	}
	for k, v := range entries {
		if os.Getenv(k) == "" {
			os.Setenv(k, v)
		}
	}
}

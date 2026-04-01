package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/messivite/gosupabase/config"
	"github.com/messivite/gosupabase/internal/scaffold"
	yamlcfg "github.com/messivite/gosupabase/internal/yaml"
)

var stdinReader = bufio.NewReader(os.Stdin)

const usage = `gosupabase - Supabase API scaffolding & code generation tool

Usage:
  gosupabase new <name>                              Create a new project
  gosupabase init                                    Initialize in existing project
  gosupabase setup [flags]                           Interactive setup (.env, .gosupabase.yaml)
  gosupabase add endpoint "METHOD /path" [--auth]    Add an endpoint to api.yaml
  gosupabase gen [flags]                             Generate handlers and server code
  gosupabase dev                                     Run server with auto-restart on changes
  gosupabase list                                    List all endpoints

Setup flags:
  --from-file <path>    Import config from an env-style file instead of prompting

Gen flags:
  --server-dir DIR       Override server output directory
  --handlers-dir DIR     Override handlers output directory
  --handlers-only        Generate only handler stubs (skip server)
`

func main() {
	args := os.Args[1:]
	if len(args) == 0 {
		fmt.Print(usage)
		os.Exit(0)
	}

	switch args[0] {
	case "new":
		cmdNew(args[1:])
	case "init":
		cmdInit()
	case "setup":
		cmdSetup(args[1:])
	case "add":
		cmdAdd(args[1:])
	case "gen":
		cmdGen(args[1:])
	case "dev":
		cmdDev()
	case "list":
		cmdList()
	case "help", "--help", "-h":
		fmt.Print(usage)
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n\n%s", args[0], usage)
		os.Exit(1)
	}
}

func promptString(label, defaultVal string) string {
	if defaultVal != "" {
		fmt.Printf("%s [%s]: ", label, defaultVal)
	} else {
		fmt.Printf("%s: ", label)
	}
	line, _ := stdinReader.ReadString('\n')
	line = strings.TrimSpace(line)
	if line == "" {
		return defaultVal
	}
	return line
}

func promptYesNo(label string, defaultYes bool) bool {
	hint := "Y/n"
	if !defaultYes {
		hint = "y/N"
	}
	fmt.Printf("%s [%s]: ", label, hint)
	line, _ := stdinReader.ReadString('\n')
	line = strings.TrimSpace(strings.ToLower(line))
	if line == "" {
		return defaultYes
	}
	return line == "y" || line == "yes"
}

func promptValidationMode(defaultVal string) string {
	fmt.Printf("JWT validation mode [auto/jwks/hs256] [%s]: ", defaultVal)
	line, _ := stdinReader.ReadString('\n')
	line = strings.TrimSpace(strings.ToLower(line))
	switch line {
	case "":
		return defaultVal
	case "auto", "jwks", "hs256":
		return line
	default:
		fmt.Println("  invalid mode, using default:", defaultVal)
		return defaultVal
	}
}

type conflictPolicy int

const (
	policyOverwrite conflictPolicy = iota
	policyMerge
	policySkip
)

func promptConflict(filename string) conflictPolicy {
	fmt.Printf("\n%s already exists. What would you like to do?\n", filename)
	fmt.Println("  [o] Overwrite  - replace the file entirely")
	fmt.Println("  [m] Merge      - add missing keys, keep existing values")
	fmt.Println("  [s] Skip       - don't touch the file")
	fmt.Print("Choose [o/m/s] (default: s): ")
	line, _ := stdinReader.ReadString('\n')
	line = strings.TrimSpace(strings.ToLower(line))
	switch line {
	case "o", "overwrite":
		return policyOverwrite
	case "m", "merge":
		return policyMerge
	default:
		return policySkip
	}
}

func writeEnvFile(path string, entries map[string]string, orderedKeys []string) error {
	var sb strings.Builder
	for _, k := range orderedKeys {
		sb.WriteString(k)
		sb.WriteString("=")
		sb.WriteString(entries[k])
		sb.WriteString("\n")
	}
	return os.WriteFile(path, []byte(sb.String()), 0644)
}

func mergeEnvFile(path string, newEntries map[string]string, orderedKeys []string) error {
	existing, err := config.ParseEnvFile(path)
	if err != nil {
		existing = make(map[string]string)
	}
	for _, k := range orderedKeys {
		if _, found := existing[k]; !found {
			existing[k] = newEntries[k]
		}
	}
	return writeEnvFile(path, existing, orderedKeys)
}

func writeGosupabaseYAML(path, serverDir, handlersDir string) error {
	content := fmt.Sprintf("output:\n  serverDir: %s\n  handlersDir: %s\n", serverDir, handlersDir)
	return os.WriteFile(path, []byte(content), 0644)
}

func setupInteractive() {
	fmt.Println("goSupabase interactive setup")
	fmt.Println(strings.Repeat("-", 40))

	port := promptString("Server port", "8080")
	supaURL := promptString("Supabase URL", "")
	anonKey := promptString("Supabase anon key", "")
	jwtSecret := promptString("Supabase JWT secret", "")
	fmt.Println("JWKS URL source: SUPABASE_URL + /auth/v1/.well-known/jwks.json")
	validationMode := promptValidationMode("auto")

	includeServiceKey := promptYesNo("Include service role key? (server-side only, never expose publicly)", false)
	serviceKey := ""
	if includeServiceKey {
		serviceKey = promptString("Supabase service role key", "")
	}

	fmt.Println()
	serverDir := promptString("Server output directory", "server")
	handlersDir := promptString("Handlers output directory", "handlers")

	envKeys := []string{"PORT", "SUPABASE_URL", "SUPABASE_ANON_KEY", "SUPABASE_JWT_SECRET", "SUPABASE_JWT_VALIDATION_MODE"}
	envMap := map[string]string{
		"PORT":                         port,
		"SUPABASE_URL":                 supaURL,
		"SUPABASE_ANON_KEY":            anonKey,
		"SUPABASE_JWT_SECRET":          jwtSecret,
		"SUPABASE_JWT_VALIDATION_MODE": validationMode,
	}
	if includeServiceKey {
		envKeys = append(envKeys, "SUPABASE_SERVICE_ROLE_KEY")
		envMap["SUPABASE_SERVICE_ROLE_KEY"] = serviceKey
	}

	fmt.Println()
	applyFileWithPolicy(".env", envMap, envKeys, serverDir, handlersDir, true)
	applyFileWithPolicy(".gosupabase.yaml", envMap, envKeys, serverDir, handlersDir, false)

	if includeServiceKey {
		fmt.Println("\n  Warning: SUPABASE_SERVICE_ROLE_KEY is for server-side use only.")
		fmt.Println("  Never expose it in client code or public endpoints.")
	}

	fmt.Println("\nSetup complete!")
}

func applyFileWithPolicy(filename string, envMap map[string]string, envKeys []string, serverDir, handlersDir string, isEnv bool) {
	_, err := os.Stat(filename)
	exists := err == nil

	if exists {
		policy := promptConflict(filename)
		switch policy {
		case policySkip:
			fmt.Printf("  skipped %s\n", filename)
			return
		case policyMerge:
			if isEnv {
				if err := mergeEnvFile(filename, envMap, envKeys); err != nil {
					fmt.Fprintf(os.Stderr, "  error merging %s: %v\n", filename, err)
					return
				}
				fmt.Printf("  merged  %s (added missing keys)\n", filename)
			} else {
				fmt.Printf("  skipped %s (merge not applicable for yaml config)\n", filename)
			}
			return
		case policyOverwrite:
			// fall through to write
		}
	}

	if isEnv {
		if err := writeEnvFile(filename, envMap, envKeys); err != nil {
			fmt.Fprintf(os.Stderr, "  error writing %s: %v\n", filename, err)
			return
		}
	} else {
		if err := writeGosupabaseYAML(filename, serverDir, handlersDir); err != nil {
			fmt.Fprintf(os.Stderr, "  error writing %s: %v\n", filename, err)
			return
		}
	}
	verb := "created"
	if exists {
		verb = "overwrote"
	}
	fmt.Printf("  %s %s\n", verb, filename)
}

func setupFromFile(path string) {
	fmt.Printf("Importing config from %s ...\n", path)

	entries, err := config.ParseEnvFile(path)
	if err != nil {
		fatal("cannot read %s: %v", path, err)
	}
	if len(entries) == 0 {
		fatal("no valid KEY=VALUE entries found in %s", path)
	}

	required := []string{"SUPABASE_URL", "SUPABASE_ANON_KEY", "SUPABASE_JWT_SECRET"}
	var missing []string
	for _, k := range required {
		if v, ok := entries[k]; !ok || v == "" {
			missing = append(missing, k)
		}
	}
	if len(missing) > 0 {
		fmt.Fprintf(os.Stderr, "  Warning: missing required keys: %s\n", strings.Join(missing, ", "))
		fmt.Println("  Continuing with available values...")
	}

	orderedKeys := []string{"PORT", "SUPABASE_URL", "SUPABASE_ANON_KEY", "SUPABASE_JWT_SECRET", "SUPABASE_JWT_VALIDATION_MODE"}
	if _, ok := entries["SUPABASE_SERVICE_ROLE_KEY"]; ok {
		orderedKeys = append(orderedKeys, "SUPABASE_SERVICE_ROLE_KEY")
	}

	if _, ok := entries["PORT"]; !ok {
		entries["PORT"] = "8080"
	}
	if _, ok := entries["SUPABASE_JWT_VALIDATION_MODE"]; !ok {
		entries["SUPABASE_JWT_VALIDATION_MODE"] = "auto"
	}

	envMap := make(map[string]string)
	for _, k := range orderedKeys {
		envMap[k] = entries[k]
	}

	serverDir := "server"
	handlersDir := "handlers"
	if v, ok := entries["SERVER_DIR"]; ok && v != "" {
		serverDir = v
	}
	if v, ok := entries["HANDLERS_DIR"]; ok && v != "" {
		handlersDir = v
	}

	applyFileWithPolicy(".env", envMap, orderedKeys, serverDir, handlersDir, true)
	applyFileWithPolicy(".gosupabase.yaml", envMap, orderedKeys, serverDir, handlersDir, false)

	if _, ok := entries["SUPABASE_SERVICE_ROLE_KEY"]; ok {
		fmt.Println("\n  Warning: SUPABASE_SERVICE_ROLE_KEY is for server-side use only.")
		fmt.Println("  Never expose it in client code or public endpoints.")
	}

	fmt.Println("\nImport complete!")
}

func cmdNew(args []string) {
	if len(args) < 1 {
		fatal("usage: gosupabase new <project-name>")
	}
	name := args[0]
	module := "github.com/example/" + name

	fmt.Printf("Creating new goSupabase project: %s\n", name)
	if err := scaffold.ScaffoldNew(name, module); err != nil {
		fatal("scaffold error: %v", err)
	}
	fmt.Printf("\nDone! Next steps:\n  cd %s\n  go mod tidy\n  gosupabase gen\n  go run ./cmd/server\n", name)
}

func cmdInit() {
	fmt.Println("Initializing goSupabase in current directory...")

	if _, err := os.Stat("api.yaml"); os.IsNotExist(err) {
		cfg := &yamlcfg.APIConfig{
			Version:  "1",
			BasePath: "/api",
			Output:   yamlcfg.OutputConfig{ServerDir: "server", HandlersDir: "handlers"},
			Endpoints: []yamlcfg.Endpoint{
				{Method: "GET", Path: "/health", Handler: "Health", Auth: false},
			},
		}
		if err := yamlcfg.Save("api.yaml", cfg); err != nil {
			fatal("writing api.yaml: %v", err)
		}
		fmt.Println("  created api.yaml")
	} else {
		fmt.Println("  api.yaml already exists")
	}

	if _, err := os.Stat(".env.example"); os.IsNotExist(err) {
		content := "PORT=8080\nSUPABASE_URL=\nSUPABASE_ANON_KEY=\nSUPABASE_SERVICE_ROLE_KEY=\nSUPABASE_JWT_SECRET=\nSUPABASE_JWT_VALIDATION_MODE=auto\n"
		os.WriteFile(".env.example", []byte(content), 0644)
		fmt.Println("  created .env.example")
	}

	fmt.Println("\nDone! Next: gosupabase add endpoint \"GET /items\" --auth")
}

func cmdSetup(args []string) {
	var fromFile string
	for i := 0; i < len(args); i++ {
		if args[i] == "--from-file" && i+1 < len(args) {
			i++
			fromFile = args[i]
		}
	}

	if fromFile != "" {
		setupFromFile(fromFile)
	} else {
		setupInteractive()
	}
}

func cmdAdd(args []string) {
	if len(args) < 2 || args[0] != "endpoint" {
		fatal("usage: gosupabase add endpoint \"METHOD /path\" [--auth]")
	}

	spec := args[1]
	parts := strings.SplitN(spec, " ", 2)
	if len(parts) != 2 {
		fatal("invalid endpoint spec %q, expected \"METHOD /path\"", spec)
	}

	method := strings.ToUpper(parts[0])
	path := parts[1]
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	authFlag := false
	for _, a := range args[2:] {
		if a == "--auth" {
			authFlag = true
		}
	}

	handler := yamlcfg.DeriveHandlerName(method, path)

	cfg, err := yamlcfg.Load("api.yaml")
	if err != nil {
		fatal("loading api.yaml: %v", err)
	}

	ep := yamlcfg.Endpoint{
		Method:  method,
		Path:    path,
		Handler: handler,
		Auth:    authFlag,
	}
	if err := cfg.AddEndpoint(ep); err != nil {
		fatal("%v", err)
	}

	if err := yamlcfg.Save("api.yaml", cfg); err != nil {
		fatal("saving api.yaml: %v", err)
	}

	fmt.Printf("Added endpoint: %s %s -> %s (auth=%v)\n", method, path, handler, authFlag)
}

func cmdGen(args []string) {
	var flagServer, flagHandlers string
	handlersOnly := false

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--server-dir":
			if i+1 < len(args) {
				i++
				flagServer = args[i]
			}
		case "--handlers-dir":
			if i+1 < len(args) {
				i++
				flagHandlers = args[i]
			}
		case "--handlers-only":
			handlersOnly = true
		}
	}

	paths := config.ResolveOutputPaths(flagServer, flagHandlers)

	fmt.Printf("Generating code (handlers=%s, server=%s, handlers-only=%v)\n", paths.HandlersDir, paths.ServerDir, handlersOnly)

	opts := scaffold.GenerateOptions{
		HandlersDir:  paths.HandlersDir,
		ServerDir:    paths.ServerDir,
		HandlersOnly: handlersOnly,
	}

	if err := scaffold.Generate("api.yaml", opts); err != nil {
		fatal("generation error: %v", err)
	}

	fmt.Println("\nDone! Run: go build ./...")
}

func cmdDev() {
	fmt.Println("[dev] starting server with auto-restart")
	fmt.Println("[dev] watching: *.go, *.yaml, *.yml, .env")

	stopSignal := make(chan os.Signal, 1)
	signal.Notify(stopSignal, os.Interrupt, syscall.SIGTERM)

	startServer := func() (*exec.Cmd, error) {
		cmd := exec.Command("go", "run", "./cmd/server")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin
		if err := cmd.Start(); err != nil {
			return nil, err
		}
		go func() { _ = cmd.Wait() }()
		return cmd, nil
	}

	stopServer := func(cmd *exec.Cmd) {
		if cmd == nil || cmd.Process == nil {
			return
		}
		_ = cmd.Process.Signal(os.Interrupt)
		time.Sleep(300 * time.Millisecond)
		_ = cmd.Process.Kill()
	}

	cmd, err := startServer()
	if err != nil {
		fatal("dev start error: %v", err)
	}
	last, _ := snapshotWatchedFiles(".")
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-stopSignal:
			fmt.Println("\n[dev] stopping...")
			stopServer(cmd)
			return
		case <-ticker.C:
			curr, _ := snapshotWatchedFiles(".")
			if !sameSnapshot(last, curr) {
				fmt.Println("[dev] change detected, restarting server...")
				stopServer(cmd)
				cmd, err = startServer()
				if err != nil {
					fmt.Printf("[dev] restart failed: %v\n", err)
				}
				last = curr
			}
		}
	}
}

func snapshotWatchedFiles(root string) (map[string]time.Time, error) {
	out := make(map[string]time.Time)
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			base := filepath.Base(path)
			if base == ".git" || base == ".cursor" || base == "node_modules" {
				return filepath.SkipDir
			}
			return nil
		}
		base := filepath.Base(path)
		if !(strings.HasSuffix(path, ".go") || strings.HasSuffix(path, ".yaml") || strings.HasSuffix(path, ".yml") || base == ".env") {
			return nil
		}
		info, statErr := d.Info()
		if statErr != nil {
			return nil
		}
		out[path] = info.ModTime()
		return nil
	})
	return out, err
}

func sameSnapshot(a, b map[string]time.Time) bool {
	if len(a) != len(b) {
		return false
	}
	for k, v := range a {
		if bv, ok := b[k]; !ok || !bv.Equal(v) {
			return false
		}
	}
	return true
}

func cmdList() {
	cfg, err := yamlcfg.Load("api.yaml")
	if err != nil {
		fatal("loading api.yaml: %v", err)
	}

	endpoints := cfg.ListEndpoints()
	if len(endpoints) == 0 {
		fmt.Println("No endpoints defined.")
		return
	}

	fmt.Printf("%-8s %-25s %-20s %-6s %s\n", "METHOD", "PATH", "HANDLER", "AUTH", "ROLES")
	fmt.Println(strings.Repeat("-", 80))
	for _, ep := range endpoints {
		roles := "-"
		if len(ep.Roles) > 0 {
			roles = strings.Join(ep.Roles, ", ")
		}
		fmt.Printf("%-8s %-25s %-20s %-6v %s\n", ep.Method, ep.Path, ep.Handler, ep.Auth, roles)
	}
}

func fatal(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "Error: "+format+"\n", args...)
	os.Exit(1)
}

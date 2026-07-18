package config

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/caarlos0/env/v11"
)

type Env string

const (
	Env_Test Env = "test"
	Env_Dev  Env = "dev"
)

type Config struct {
	DatabaseName     string `env:"DB_NAME"`
	DatabaseHost     string `env:"DB_HOST"`
	DatabasePort     string `env:"DB_PORT"`
	DatabasePortTest string `env:"DB_PORT_TEST"`
	DatabaseUser     string `env:"DB_USER"`
	DatabasePassword string `env:"DB_PASSWORD"`
	Env              Env    `env:"ENV" envDefault:"dev"`
	ProjectRoot      string `env:"PROJECT_ROOT"`
}

func (c *Config) DatabaseUrl() string {
	port := c.DatabasePort
	if c.Env == Env_Test {
		port = c.DatabasePortTest
	}
	return fmt.Sprintf("postgresql://%s:%s@%s:%s/%s?sslmode=disable",
		c.DatabaseUser,
		c.DatabasePassword,
		c.DatabaseHost,
		port,
		c.DatabaseName)
}

func New() (*Config, error) {
	// Resolve the project root first (independent of any env), then load the
	// .env file that sits next to go.mod. This makes the app work even when it
	// is run outside a direnv shell (e.g. VS Code's code-runner), so no
	// hardcoded defaults are needed.
	root, rootErr := findProjectRoot()
	if rootErr != nil {
		return nil, rootErr
	}
	if err := loadDotEnv(filepath.Join(root, ".env")); err != nil {
		return nil, err
	}

	cfg, err := env.ParseAs[Config]()
	if err != nil {
		return nil, fmt.Errorf("Could not load config %w", err)
	}

	// PROJECT_ROOT in .env is often the literal "$(pwd)" that never expanded,
	// so always trust the resolved root.
	cfg.ProjectRoot = root

	return &cfg, nil
}

// loadDotEnv reads a simple KEY=VALUE .env file and sets any variables that are
// not already present in the environment, so real environment variables (e.g.
// those exported by direnv) always take precedence. A missing file is not an
// error.
func loadDotEnv(path string) error {
	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("could not open %s: %w", path, err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		line = strings.TrimPrefix(line, "export ")

		key, value, found := strings.Cut(line, "=")
		if !found {
			continue
		}
		key = strings.TrimSpace(key)
		value = strings.Trim(strings.TrimSpace(value), `"'`)
		if key == "" {
			continue
		}
		if _, exists := os.LookupEnv(key); !exists {
			if err := os.Setenv(key, value); err != nil {
				return fmt.Errorf("could not set %s: %w", key, err)
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("could not read %s: %w", path, err)
	}
	return nil
}

// findProjectRoot walks up from the current working directory until it finds
// the directory containing go.mod. The path is returned with forward slashes so
// it can be embedded directly in a file:// migration URL on any platform.
func findProjectRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("could not determine working directory: %w", err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return filepath.ToSlash(dir), nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("could not locate project root (go.mod not found)")
		}
		dir = parent
	}
}

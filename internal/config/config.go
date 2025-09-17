package config

import (
	"errors"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	AppEnv      string
	HTTPPort    string
	DatabaseURL string
}

func MustLoad() *Config {
	_ = loadDotenvIfExists("configs/.env")

	if fileExists("configs/config.yaml") {
		return mustLoadYAML("configs/config.yaml")
	}

	cfg := &Config{
		AppEnv:      getEnv("APP_ENV", "dev"),
		HTTPPort:    getEnv("HTTP_PORT", "8080"),
		DatabaseURL: getEnv("DATABASE_URL", ""),
	}
	if cfg.DatabaseURL == "" {
		panic("DATABASE_URL is required")
	}
	return cfg
}

type yamlCfg struct {
	App struct {
		Env      string `yaml:"env"`
		HTTPPort string `yaml:"http_port"`
	} `yaml:"app"`
	Database struct {
		URL string `yaml:"url"`
	} `yaml:"database"`
}

func mustLoadYAML(path string) *Config {
	data, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}
	var yc yamlCfg
	if err := yaml.Unmarshal(data, &yc); err != nil {
		panic(err)
	}
	cfg := &Config{
		AppEnv:      firstNonEmpty(yc.App.Env, "dev"),
		HTTPPort:    firstNonEmpty(yc.App.HTTPPort, "8080"),
		DatabaseURL: yc.Database.URL,
	}
	if cfg.DatabaseURL == "" {
		panic(errors.New("database.url must be set in config.yaml"))
	}
	return cfg
}

func firstNonEmpty(s, def string) string {
	if s != "" {
		return s
	}
	return def
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func getEnv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func loadDotenvIfExists(path string) error {
	if !fileExists(path) {
		return nil
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	lines := splitLines(string(b))
	for _, line := range lines {
		if len(line) == 0 || line[0] == '#' {
			continue
		}
		kv := splitKV(line)
		if kv[0] != "" && kv[1] != "" {
			os.Setenv(kv[0], kv[1])
		}
	}
	return nil
}

func splitLines(s string) []string {
	var res []string
	cur := ""
	for _, r := range s {
		if r == '\r' {
			continue
		}
		if r == '\n' {
			res = append(res, cur)
			cur = ""
			continue
		}
		cur += string(r)
	}
	if cur != "" {
		res = append(res, cur)
	}
	return res
}

func splitKV(s string) [2]string {
	for i := 0; i < len(s); i++ {
		if s[i] == '=' {
			return [2]string{s[:i], s[i+1:]}
		}
	}
	return [2]string{"", ""}
}

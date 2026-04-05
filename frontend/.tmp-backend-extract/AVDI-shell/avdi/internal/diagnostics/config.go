package diagnostics

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config представляет конфигурационный файл диагностики
type Config struct {
	Version  string    `yaml:"version"`
	Commands []Command `yaml:"commands"`
}

// Command определяет одну диагностическую команду
type Command struct {
	Name        string            `yaml:"name"`
	Description string            `yaml:"description"`
	Command     string            `yaml:"command"`
	Args        []string          `yaml:"args,omitempty"`
	Timeout     int               `yaml:"timeout,omitempty"`
	Variables   []Variable        `yaml:"variables,omitempty"`
	ParseOutput string            `yaml:"parse_output,omitempty"`
	Env         map[string]string `yaml:"env,omitempty"`
}

// Variable определяет переменную для подстановки в команду
type Variable struct {
	Name     string `yaml:"name"`
	Required bool   `yaml:"required"`
	Default  string `yaml:"default,omitempty"`
}

// LoadConfig загружает конфигурацию из файла YAML
func LoadConfig(path string) (*Config, error) {
	if path == "" {
		// Попробуем найти конфиг в стандартных местах
		candidates := []string{
			"./diagnostics.yaml",
			"/etc/avdi/diagnostics.yaml",
			filepath.Join(os.Getenv("HOME"), ".config/avdi/diagnostics.yaml"),
		}
		for _, candidate := range candidates {
			if _, err := os.Stat(candidate); err == nil {
				path = candidate
				break
			}
		}
		if path == "" {
			return nil, fmt.Errorf("no diagnostics config found")
		}
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("parse yaml: %w", err)
	}

	if config.Version == "" {
		config.Version = "1.0"
	}

	return &config, nil
}

// FindCommand ищет команду по имени
func (c *Config) FindCommand(name string) (*Command, bool) {
	for _, cmd := range c.Commands {
		if cmd.Name == name {
			return &cmd, true
		}
	}
	return nil, false
}

// BuildCommand строит полную команду с подстановкой переменных
func (c *Command) BuildCommand(vars map[string]string) (string, []string, error) {
	// Подставляем переменные в команду
	cmd := c.Command
	for _, v := range c.Variables {
		value, ok := vars[v.Name]
		if !ok {
			if v.Required && v.Default == "" {
				return "", nil, fmt.Errorf("required variable %s not provided", v.Name)
			}
			value = v.Default
		}
		placeholder := "{{" + v.Name + "}}"
		cmd = strings.ReplaceAll(cmd, placeholder, value)
	}

	// Подставляем переменные в аргументы
	args := make([]string, len(c.Args))
	for i, arg := range c.Args {
		for _, v := range c.Variables {
			value, ok := vars[v.Name]
			if !ok {
				value = v.Default
			}
			placeholder := "{{" + v.Name + "}}"
			arg = strings.ReplaceAll(arg, placeholder, value)
		}
		args[i] = arg
	}

	return cmd, args, nil
}
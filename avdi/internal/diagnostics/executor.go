package diagnostics

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// ExecutionResult содержит результат выполнения команды
type ExecutionResult struct {
	Command   string          `json:"command"`
	Args      []string        `json:"args,omitempty"`
	ExitCode  int             `json:"exit_code"`
	Stdout    string          `json:"stdout"`
	Stderr    string          `json:"stderr"`
	Duration  float64         `json:"duration_seconds"`
	Parsed    json.RawMessage `json:"parsed,omitempty"`
	Error     string          `json:"error,omitempty"`
}

// ExecuteCommand выполняет команду с заданными аргументами и таймаутом
func ExecuteCommand(cmdName string, args []string, timeout int, env map[string]string) (*ExecutionResult, error) {
	start := time.Now()

	cmd := exec.Command(cmdName, args...)
	if env != nil {
		for k, v := range env {
			cmd.Env = append(cmd.Env, k+"="+v)
		}
	}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	var err error
	if timeout > 0 {
		// Запускаем с таймаутом
		err = cmd.Start()
		if err != nil {
			return nil, fmt.Errorf("start command: %w", err)
		}

		timer := time.AfterFunc(time.Duration(timeout)*time.Second, func() {
			cmd.Process.Kill()
		})

		err = cmd.Wait()
		timer.Stop()
	} else {
		err = cmd.Run()
	}

	duration := time.Since(start).Seconds()

	result := &ExecutionResult{
		Command:  cmdName,
		Args:     args,
		Stdout:   strings.TrimSpace(stdout.String()),
		Stderr:   strings.TrimSpace(stderr.String()),
		Duration: duration,
	}

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitErr.ExitCode()
		} else {
			result.ExitCode = 1
			result.Error = err.Error()
		}
	} else {
		result.ExitCode = 0
	}

	return result, nil
}

// ExecuteDiagnostic выполняет диагностическую команду по имени с переменными
func ExecuteDiagnostic(configPath, commandName string, variables map[string]string) (*ExecutionResult, error) {
	config, err := LoadConfig(configPath)
	if err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}

	cmd, found := config.FindCommand(commandName)
	if !found {
		return nil, fmt.Errorf("command %s not found in config", commandName)
	}

	cmdName, args, err := cmd.BuildCommand(variables)
	if err != nil {
		return nil, fmt.Errorf("build command: %w", err)
	}

	result, err := ExecuteCommand(cmdName, args, cmd.Timeout, cmd.Env)
	if err != nil {
		return nil, fmt.Errorf("execute command: %w", err)
	}

	// Парсинг вывода, если указано
	if cmd.ParseOutput != "" {
		// Базовая поддержка парсинга JSON
		if cmd.ParseOutput == "json" {
			var parsed json.RawMessage
			if err := json.Unmarshal([]byte(result.Stdout), &parsed); err == nil {
				result.Parsed = parsed
			}
		}
		// Можно добавить другие форматы (csv, xml, etc.)
	}

	return result, nil
}
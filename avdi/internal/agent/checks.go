package agent

import (
    "bytes"
    "encoding/json"
    "fmt"
    "os"
    "os/exec"
    "strings"

    "diag-system/internal/diagnostics"
)

type CheckResult struct {
    ExitCode   int
    ResultJSON string
    Stdout     string
    Stderr     string
    Logs       string
}

func RunCheck(checkType, payload string) CheckResult {
    switch checkType {
    case "hostname":
        return hostnameCheck()
    case "ports":
        return portsCheck()
    case "ping":
        return pingCheck(payload)
    case "diagnostic":
        return diagnosticCheck(payload)
    case "bash":
        return bashCheck(payload)
    default:
        return CheckResult{
            ExitCode: 1,
            Stderr:   "unsupported check type",
            Logs:     "unsupported check type: " + checkType,
        }
    }
}

func hostnameCheck() CheckResult {
    return runCommandJSON("hostname", nil, func(stdout string) (string, error) {
        out, err := json.Marshal(map[string]string{"hostname": strings.TrimSpace(stdout)})
        return string(out), err
    })
}

func portsCheck() CheckResult {
    return runCommandJSON("ss", []string{"-tulpn"}, func(stdout string) (string, error) {
        out, err := json.Marshal(map[string]string{"ports": stdout})
        return string(out), err
    })
}

func pingCheck(target string) CheckResult {
    target = strings.TrimSpace(target)
    if target == "" {
        return CheckResult{ExitCode: 1, Stderr: "empty ping target", Logs: "empty ping target"}
    }
    return runCommandJSON("ping", []string{"-c", "2", target}, func(stdout string) (string, error) {
        out, err := json.Marshal(map[string]string{"target": target, "output": stdout})
        return string(out), err
    })
}

func runCommandJSON(name string, args []string, mapper func(stdout string) (string, error)) CheckResult {
    cmd := exec.Command(name, args...)
    var stdout, stderr bytes.Buffer
    cmd.Stdout = &stdout
    cmd.Stderr = &stderr

    err := cmd.Run()
    exitCode := 0
    if err != nil {
        exitCode = 1
    }

    resultJSON, mapErr := mapper(stdout.String())
    if mapErr != nil {
        exitCode = 1
        stderr.WriteString("; json map error: " + mapErr.Error())
    }

    logs := fmt.Sprintf("cmd: %s %s\nstdout:\n%s\nstderr:\n%s", name, strings.Join(args, " "), stdout.String(), stderr.String())

    return CheckResult{
        ExitCode:   exitCode,
        ResultJSON: resultJSON,
        Stdout:     stdout.String(),
        Stderr:     stderr.String(),
        Logs:       logs,
    }
}

func diagnosticCheck(payload string) CheckResult {
    // Парсим payload как JSON: {"command": "имя_команды", "vars": {"var1": "value1"}}
    var req struct {
        Command string            `json:"command"`
        Vars    map[string]string `json:"vars,omitempty"`
        Config  string            `json:"config,omitempty"`
    }
    
    if err := json.Unmarshal([]byte(payload), &req); err != nil {
        return CheckResult{
            ExitCode: 1,
            Stderr:   "invalid payload JSON",
            Logs:     "failed to parse diagnostic payload: " + err.Error(),
        }
    }
    
    if req.Command == "" {
        return CheckResult{
            ExitCode: 1,
            Stderr:   "command name is required",
            Logs:     "diagnostic payload missing 'command' field",
        }
    }
    
    // Выполняем диагностическую команду
    result, err := diagnostics.ExecuteDiagnostic(req.Config, req.Command, req.Vars)
    if err != nil {
        return CheckResult{
            ExitCode: 1,
            Stderr:   err.Error(),
            Logs:     "diagnostic execution failed: " + err.Error(),
        }
    }
    
    // Конвертируем результат в JSON
    resultJSON, err := json.Marshal(result)
    if err != nil {
        return CheckResult{
            ExitCode: 1,
            Stderr:   "failed to marshal result",
            Logs:     "json marshal error: " + err.Error(),
        }
    }
    
    logs := fmt.Sprintf("diagnostic command: %s\nstdout:\n%s\nstderr:\n%s",
        result.Command, result.Stdout, result.Stderr)
    
    return CheckResult{
        ExitCode:   result.ExitCode,
        ResultJSON: string(resultJSON),
        Stdout:     result.Stdout,
        Stderr:     result.Stderr,
        Logs:       logs,
    }
}

func bashCheck(payload string) CheckResult {
    // payload может быть либо непосредственно bash-скриптом, либо JSON с полями script и args
    var req struct {
        Script string   `json:"script"`
        Args   []string `json:"args,omitempty"`
        Env    map[string]string `json:"env,omitempty"`
    }
    
    // Пытаемся разобрать как JSON, если не получается, считаем payload самим скриптом
    if err := json.Unmarshal([]byte(payload), &req); err != nil || req.Script == "" {
        req.Script = payload
    }
    
    if req.Script == "" {
        return CheckResult{
            ExitCode: 1,
            Stderr:   "empty script",
            Logs:     "bash script is empty",
        }
    }
    
    // Создаем временный файл со скриптом
    tmpFile, err := os.CreateTemp("", "avdi-bash-*.sh")
    if err != nil {
        return CheckResult{
            ExitCode: 1,
            Stderr:   "failed to create temp file",
            Logs:     "temp file creation error: " + err.Error(),
        }
    }
    defer os.Remove(tmpFile.Name())
    
    // Записываем скрипт
    if _, err := tmpFile.WriteString(req.Script); err != nil {
        return CheckResult{
            ExitCode: 1,
            Stderr:   "failed to write script",
            Logs:     "script write error: " + err.Error(),
        }
    }
    tmpFile.Close()
    
    // Делаем файл исполняемым
    if err := os.Chmod(tmpFile.Name(), 0755); err != nil {
        return CheckResult{
            ExitCode: 1,
            Stderr:   "failed to make script executable",
            Logs:     "chmod error: " + err.Error(),
        }
    }
    
    // Выполняем скрипт
    cmd := exec.Command("/bin/bash", append([]string{tmpFile.Name()}, req.Args...)...)
    if req.Env != nil {
        for k, v := range req.Env {
            cmd.Env = append(os.Environ(), k+"="+v)
        }
    }
    
    var stdout, stderr bytes.Buffer
    cmd.Stdout = &stdout
    cmd.Stderr = &stderr
    
    err = cmd.Run()
    exitCode := 0
    if err != nil {
        if exitErr, ok := err.(*exec.ExitError); ok {
            exitCode = exitErr.ExitCode()
        } else {
            exitCode = 1
        }
    }
    
    resultJSON, _ := json.Marshal(map[string]interface{}{
        "script": req.Script[:min(100, len(req.Script))] + "...",
        "exit_code": exitCode,
        "stdout_length": len(stdout.String()),
        "stderr_length": len(stderr.String()),
    })
    
    logs := fmt.Sprintf("bash script executed\nstdout:\n%s\nstderr:\n%s", stdout.String(), stderr.String())
    
    return CheckResult{
        ExitCode:   exitCode,
        ResultJSON: string(resultJSON),
        Stdout:     stdout.String(),
        Stderr:     stderr.String(),
        Logs:       logs,
    }
}

func min(a, b int) int {
    if a < b {
        return a
    }
    return b
}
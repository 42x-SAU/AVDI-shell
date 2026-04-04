package agent

import (
    "bytes"
    "encoding/json"
    "fmt"
    "os/exec"
    "strings"
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
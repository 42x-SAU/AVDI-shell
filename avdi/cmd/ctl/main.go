package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	"diag-system/internal/deploy"

	"golang.org/x/term"
)

type createTaskRequest struct {
	AgentID    int64  `json:"agent_id"`
	CheckType  string `json:"check_type"`
	Payload    string `json:"payload"`
	MaxRetries int    `json:"max_retries"`
}

type Agent struct {
	ID            int64     `json:"id"`
	Name          string    `json:"name"`
	LastHeartbeat time.Time `json:"last_heartbeat"`
	CreatedAt     time.Time `json:"created_at"`
}

type Task struct {
	ID         int64      `json:"id"`
	AgentID    int64      `json:"agent_id"`
	CheckType  string     `json:"check_type"`
	Payload    string     `json:"payload"`
	Status     string     `json:"status"`
	RetryCount int        `json:"retry_count"`
	MaxRetries int        `json:"max_retries"`
	CreatedAt  time.Time  `json:"created_at"`
	StartedAt  *time.Time `json:"started_at,omitempty"`
	FinishedAt *time.Time `json:"finished_at,omitempty"`
}

type TaskResult struct {
	ID         int64     `json:"id"`
	TaskID     int64     `json:"task_id"`
	ExitCode   int       `json:"exit_code"`
	ResultJSON string    `json:"result_json"`
	Stdout     string    `json:"stdout"`
	Stderr     string    `json:"stderr"`
	Logs       string    `json:"logs"`
	CreatedAt  time.Time `json:"created_at"`
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	client := &http.Client{Timeout: 20 * time.Second}
	defaultServer := getenv("AVDI_SERVER", "http://localhost:8081")

	switch os.Args[1] {
	case "health":
		serverURL := parseServerFlag(defaultServer, os.Args[2:])
		if err := cmdHealth(client, serverURL); err != nil {
			exitErr(err)
		}

	case "agents":
		serverURL := parseServerFlag(defaultServer, os.Args[2:])
		if err := cmdAgents(client, serverURL); err != nil {
			exitErr(err)
		}

	case "tasks":
		serverURL := parseServerFlag(defaultServer, os.Args[2:])
		if err := cmdTasks(client, serverURL); err != nil {
			exitErr(err)
		}

	case "results":
		if err := cmdResults(client, defaultServer, os.Args[2:]); err != nil {
			exitErr(err)
		}

	case "create-task":
		if err := cmdCreateTask(client, defaultServer, os.Args[2:]); err != nil {
			exitErr(err)
		}

	case "deploy-agent":
		if err := cmdDeployAgent(os.Args[2:]); err != nil {
			exitErr(err)
		}

	case "help", "--help", "-h":
		printUsage()

	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}
}

func cmdHealth(client *http.Client, serverURL string) error {
	resp, err := client.Get(joinURL(serverURL, "/health"))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	fmt.Println(string(body))

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}
	return nil
}

func cmdDeployAgent(args []string) error {
	fs := flag.NewFlagSet("deploy-agent", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	sshHost := fs.String("ssh-host", "", "Remote host (hostname or IP)")
	sshUser := fs.String("ssh-user", "", "SSH user")
	sshPassword := fs.String("ssh-password", "", "SSH password (optional: use AVDI_SSH_PASSWORD or prompt)")
	sshPort := fs.Int("ssh-port", 22, "SSH TCP port (default 22; if omitted, AVDI_SSH_PORT is used when set)")
	serverURL := fs.String("server-url", "", "AVDI API base URL reachable from the agent host")
	agentName := fs.String("agent-name", "", "Agent display name (registered on first start)")
	image := fs.String("image", "", "Docker image ref (must exist in a registry you can pull, or use --skip-pull if already on the host)")
	container := fs.String("container", "avdi-agent", "Container name on the remote host")
	skipPull := fs.Bool("skip-pull", false, "Skip docker pull (use when the image is already on the remote: docker load, local build, etc.)")

	if err := fs.Parse(args); err != nil {
		return fmt.Errorf("parse flags: %w", err)
	}

	if *sshHost == "" {
		return fmt.Errorf("--ssh-host is required")
	}
	if *sshUser == "" {
		return fmt.Errorf("--ssh-user is required")
	}
	if *serverURL == "" {
		return fmt.Errorf("--server-url is required")
	}
	if *agentName == "" {
		return fmt.Errorf("--agent-name is required")
	}
	if *image == "" {
		return fmt.Errorf("--image is required")
	}

	port := *sshPort
	explicitSSHPort := false
	fs.Visit(func(f *flag.Flag) {
		if f.Name == "ssh-port" {
			explicitSSHPort = true
		}
	})
	if !explicitSSHPort {
		if v := strings.TrimSpace(os.Getenv("AVDI_SSH_PORT")); v != "" {
			p, err := strconv.Atoi(v)
			if err != nil || p < 1 || p > 65535 {
				return fmt.Errorf("AVDI_SSH_PORT must be an integer 1–65535")
			}
			port = p
		}
	}

	password := *sshPassword
	if password == "" {
		password = os.Getenv("AVDI_SSH_PASSWORD")
	}
	if password == "" {
		fmt.Fprint(os.Stderr, "SSH password: ")
		b, err := term.ReadPassword(int(os.Stdin.Fd()))
		if err != nil {
			return fmt.Errorf("read password: %w", err)
		}
		fmt.Fprintln(os.Stderr)
		password = string(b)
	}

	return deploy.DeployDockerAgent(deploy.Options{
		Host:          *sshHost,
		Port:          port,
		User:          *sshUser,
		Secret:        password,
		ServerURL:     *serverURL,
		AgentName:     *agentName,
		Image:         *image,
		ContainerName: *container,
		SkipPull:      *skipPull,
	})
}

func cmdCreateTask(client *http.Client, defaultServer string, args []string) error {
	fs := flag.NewFlagSet("create-task", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	serverURL := fs.String("server", defaultServer, "Server base URL")
	agentID := fs.Int64("agent", 0, "Agent ID")
	checkType := fs.String("check", "", "Check type: hostname | ping | ports")
	payload := fs.String("payload", "", "Optional payload")
	retries := fs.Int("retries", 0, "Max retries")

	if err := fs.Parse(args); err != nil {
		return fmt.Errorf("failed to parse flags: %w", err)
	}

	if *agentID <= 0 {
		return fmt.Errorf("--agent is required and must be > 0")
	}
	if *checkType == "" {
		return fmt.Errorf("--check is required")
	}

	switch *checkType {
	case "hostname", "ping", "ports":
	default:
		return fmt.Errorf("unsupported --check value: %s", *checkType)
	}

	if *retries < 0 {
		return fmt.Errorf("--retries cannot be negative")
	}

	reqBody := createTaskRequest{
		AgentID:    *agentID,
		CheckType:  *checkType,
		Payload:    *payload,
		MaxRetries: *retries,
	}

	var respBody map[string]any
	status, err := postJSON(client, joinURL(*serverURL, "/tasks"), reqBody, &respBody)
	if err != nil {
		return err
	}
	if status != http.StatusCreated {
		return fmt.Errorf("unexpected status: %d", status)
	}

	prettyPrintJSON(respBody)
	return nil
}

func cmdAgents(client *http.Client, serverURL string) error {
	var agents []Agent
	status, err := getJSON(client, joinURL(serverURL, "/agents"), &agents)
	if err != nil {
		return err
	}
	if status != http.StatusOK {
		return fmt.Errorf("unexpected status: %d", status)
	}

	if len(agents) == 0 {
		fmt.Println("No agents found")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tNAME\tLAST_HEARTBEAT\tCREATED_AT")
	for _, a := range agents {
		fmt.Fprintf(
			w,
			"%d\t%s\t%s\t%s\n",
			a.ID,
			a.Name,
			formatTime(a.LastHeartbeat),
			formatTime(a.CreatedAt),
		)
	}
	return w.Flush()
}

func cmdTasks(client *http.Client, serverURL string) error {
	var tasks []Task
	status, err := getJSON(client, joinURL(serverURL, "/tasks"), &tasks)
	if err != nil {
		return err
	}
	if status != http.StatusOK {
		return fmt.Errorf("unexpected status: %d", status)
	}

	if len(tasks) == 0 {
		fmt.Println("No tasks found")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tAGENT\tCHECK\tSTATUS\tRETRIES\tCREATED_AT")
	for _, t := range tasks {
		fmt.Fprintf(
			w,
			"%d\t%d\t%s\t%s\t%d/%d\t%s\n",
			t.ID,
			t.AgentID,
			t.CheckType,
			t.Status,
			t.RetryCount,
			t.MaxRetries,
			formatTime(t.CreatedAt),
		)
	}
	return w.Flush()
}

func cmdResults(client *http.Client, defaultServer string, args []string) error {
	fs := flag.NewFlagSet("results", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	serverURL := fs.String("server", defaultServer, "Server base URL")
	showLogs := fs.Bool("logs", false, "Show logs column")
	if err := fs.Parse(args); err != nil {
		return fmt.Errorf("failed to parse flags: %w", err)
	}

	var results []TaskResult
	status, err := getJSON(client, joinURL(*serverURL, "/results"), &results)
	if err != nil {
		return err
	}
	if status != http.StatusOK {
		return fmt.Errorf("unexpected status: %d", status)
	}

	if len(results) == 0 {
		fmt.Println("No results found")
		return nil
	}

	if *showLogs {
		for _, r := range results {
			fmt.Printf("TASK %d | EXIT %d | CREATED %s\n", r.TaskID, r.ExitCode, formatTime(r.CreatedAt))
			fmt.Printf("RESULT_JSON: %s\n", r.ResultJSON)
			fmt.Printf("STDOUT:\n%s\n", r.Stdout)
			fmt.Printf("STDERR:\n%s\n", r.Stderr)
			fmt.Printf("LOGS:\n%s\n", r.Logs)
			fmt.Println(strings.Repeat("-", 80))
		}
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tTASK_ID\tEXIT_CODE\tRESULT_JSON\tCREATED_AT")
	for _, r := range results {
		fmt.Fprintf(
			w,
			"%d\t%d\t%d\t%s\t%s\n",
			r.ID,
			r.TaskID,
			r.ExitCode,
			trimForTable(r.ResultJSON, 60),
			formatTime(r.CreatedAt),
		)
	}
	return w.Flush()
}

func getJSON(client *http.Client, url string, out any) (int, error) {
	resp, err := client.Get(url)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return resp.StatusCode, fmt.Errorf("request failed: %s", strings.TrimSpace(string(body)))
	}

	if out != nil {
		if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
			return resp.StatusCode, err
		}
	}

	return resp.StatusCode, nil
}

func postJSON(client *http.Client, url string, payload any, out any) (int, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return 0, err
	}

	resp, err := client.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		respBody, _ := io.ReadAll(resp.Body)
		return resp.StatusCode, fmt.Errorf("request failed: %s", strings.TrimSpace(string(respBody)))
	}

	if out != nil {
		if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
			return resp.StatusCode, err
		}
	}

	return resp.StatusCode, nil
}

func parseServerFlag(defaultServer string, args []string) string {
	fs := flag.NewFlagSet("global", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	server := fs.String("server", defaultServer, "Server base URL")
	_ = fs.Parse(args)
	return *server
}

func joinURL(base, path string) string {
	return strings.TrimRight(base, "/") + path
}

func trimForTable(s string, limit int) string {
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.TrimSpace(s)
	if len(s) <= limit {
		return s
	}
	return s[:limit-3] + "..."
}

func formatTime(t time.Time) string {
	if t.IsZero() {
		return "-"
	}
	return t.Format(time.RFC3339)
}

func prettyPrintJSON(v any) {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		fmt.Println(v)
		return
	}
	fmt.Println(string(b))
}

func printUsage() {
	fmt.Print(`avdi - CLI for AVDI backend

Usage:
  avdi [command] [flags]

Commands:
  health
      Check server health

  agents
      List registered agents

  tasks
      List tasks

  results [--logs]
      List task results
      Use --logs to print full stdout/stderr/logs

  create-task --agent <id> --check <hostname|ping|ports> [--payload <value>] [--retries <n>]
      Create a new task

  deploy-agent --ssh-host <host> --ssh-user <user> --server-url <url> --agent-name <name> --image <ref> [--ssh-port <n>]
      Deploy agent container on a remote host over SSH (password: flag, AVDI_SSH_PASSWORD, or prompt)
      SSH port: use --ssh-port (default 22) or AVDI_SSH_PORT when --ssh-port is not passed

Global config:
  Environment variables:
      AVDI_SERVER=http://localhost:8081
      AVDI_SSH_PASSWORD=...   (optional, for deploy-agent)
      AVDI_SSH_PORT=...       (optional SSH port if you omit --ssh-port)

Examples:
  avdi health
  avdi agents
  avdi tasks
  avdi results
  avdi results --logs
  avdi create-task --agent 1 --check hostname
  avdi create-task --agent 1 --check ping --payload 8.8.8.8
  avdi create-task --agent 1 --check ports
  avdi create-task --server http://localhost:8081 --agent 1 --check hostname

  avdi deploy-agent --ssh-host 203.0.113.10 --ssh-user ubuntu --ssh-port 2222 \
      --server-url http://203.0.113.1:8081 --agent-name edge-1 --image ghcr.io/you/avdi-agent:latest

  avdi deploy-agent ... --image avdi-agent:local --skip-pull
      (after docker load / docker build on the remote host)
`)
}

func exitErr(err error) {
	fmt.Fprintln(os.Stderr, "Error:", err)
	os.Exit(1)
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"text/tabwriter"
	"time"
)

const (
	colorReset  = "\033[0m"
	colorCyan   = "\033[36m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorRed    = "\033[31m"
	colorGray   = "\033[90m"
	colorBold   = "\033[1m"
)

type createTaskRequest struct {
	AgentID    int64  `json:"agent_id"`
	CheckType  string `json:"check_type"`
	Payload    string `json:"payload"`
	MaxRetries int    `json:"max_retries"`
}

type registerAgentRequest struct {
	Name string `json:"name"`
}

type registerAgentResponse struct {
	AgentID int64  `json:"agent_id"`
	Token   string `json:"token"`
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
	serverURL := getenv("AVDI_SERVER", "http://localhost:8081")
	client := &http.Client{Timeout: 20 * time.Second}

	clearScreen()
	printBanner(serverURL)

	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Printf("%savdi%s> ", colorCyan, colorReset)

		if !scanner.Scan() {
			fmt.Println()
			return
		}

		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		args := splitArgs(line)
		if len(args) == 0 {
			continue
		}

		switch args[0] {
		case "exit", "quit":
			fmt.Println(colorGray + "bye" + colorReset)
			return

		case "help":
			printHelp()

		case "clear":
			clearScreen()
			printBanner(serverURL)

		case "server":
			if len(args) < 2 {
				printError("usage: server <url>")
				continue
			}
			serverURL = strings.TrimSpace(args[1])
			printSuccess("server set to " + serverURL)

		case "health":
			if err := cmdHealth(client, serverURL); err != nil {
				printError(err.Error())
			}

		case "agents":
			if err := cmdAgents(client, serverURL); err != nil {
				printError(err.Error())
			}

		case "tasks":
			if err := cmdTasks(client, serverURL); err != nil {
				printError(err.Error())
			}

		case "results":
			if err := cmdResults(client, serverURL, args[1:]); err != nil {
				printError(err.Error())
			}

		case "create-task":
			if err := cmdCreateTask(client, serverURL, args[1:]); err != nil {
				printError(err.Error())
			}

		case "register-agent", "add-agent":
			if err := cmdRegisterAgent(client, serverURL, args[1:]); err != nil {
				printError(err.Error())
			}

		case "get":
			if err := cmdRawGet(client, serverURL, args[1:]); err != nil {
				printError(err.Error())
			}

		case "post":
			if err := cmdRawPost(client, serverURL, args[1:]); err != nil {
				printError(err.Error())
			}

		default:
			printError("unknown command: " + args[0])
			fmt.Println("type 'help' to see available commands")
		}
	}
}

func printBanner(serverURL string) {
	fmt.Println(colorBold + colorCyan + "AVDI shell" + colorReset)
	fmt.Println(colorGray + strings.Repeat("=", 60) + colorReset)
	fmt.Println("Interactive CLI for AVDI backend")
	fmt.Println()
	fmt.Println("Current server:", colorYellow+serverURL+colorReset)
	fmt.Println()
	printHelp()
}

func printHelp() {
	fmt.Println("Commands:")
	fmt.Println("  health")
	fmt.Println("      Check server health")
	fmt.Println()
	fmt.Println("  agents")
	fmt.Println("      Show registered agents")
	fmt.Println()
	fmt.Println("  tasks")
	fmt.Println("      Show tasks")
	fmt.Println()
	fmt.Println("  results [--logs]")
	fmt.Println("      Show task results")
	fmt.Println("      --logs   print full stdout/stderr/logs")
	fmt.Println()
	fmt.Println("  register-agent --name <value>")
	fmt.Println("      Register a new agent through POST /agents/register")
	fmt.Println()
	fmt.Println("  add-agent --name <value>")
	fmt.Println("      Alias for register-agent")
	fmt.Println()
	fmt.Println("  create-task --agent <id> --check <hostname|ping|ports> [--payload <value>] [--retries <n>]")
	fmt.Println("      Create a new diagnostic task")
	fmt.Println()
	fmt.Println("  get /path")
	fmt.Println("      Raw GET request")
	fmt.Println()
	fmt.Println("  post /path '{\"key\":\"value\"}'")
	fmt.Println("      Raw POST request")
	fmt.Println()
	fmt.Println("  server <url>")
	fmt.Println("      Change active server URL inside shell")
	fmt.Println()
	fmt.Println("  help")
	fmt.Println("      Show this help")
	fmt.Println()
	fmt.Println("  exit | quit")
	fmt.Println("      Exit shell")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  health")
	fmt.Println("  agents")
	fmt.Println("  tasks")
	fmt.Println("  results")
	fmt.Println("  results --logs")
	fmt.Println("  register-agent --name demo-agent")
	fmt.Println("  add-agent --name demo-agent")
	fmt.Println("  create-task --agent 3 --check hostname")
	fmt.Println("  create-task --agent 3 --check ping --payload 8.8.8.8")
	fmt.Println("  create-task --agent 3 --check ports")
	fmt.Println("  post /agents/register '{\"name\":\"manual-agent\"}'")
	fmt.Println()
}

func cmdHealth(client *http.Client, serverURL string) error {
	var out map[string]any
	status, err := getJSON(client, joinURL(serverURL, "/health"), &out)
	if err != nil {
		return err
	}
	if status != http.StatusOK {
		return fmt.Errorf("unexpected status: %d", status)
	}
	printSuccess("server is healthy")
	prettyPrintJSON(out)
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

	fmt.Println(colorBold + "Agents" + colorReset)
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tNAME\tLAST_HEARTBEAT\tCREATED_AT")
	for _, a := range agents {
		fmt.Fprintf(w, "%d\t%s\t%s\t%s\n",
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

	fmt.Println(colorBold + "Tasks" + colorReset)
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tAGENT\tCHECK\tSTATUS\tRETRIES\tCREATED_AT")
	for _, t := range tasks {
		fmt.Fprintf(w, "%d\t%d\t%s\t%s\t%d/%d\t%s\n",
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

func cmdResults(client *http.Client, serverURL string, args []string) error {
	fs := flag.NewFlagSet("results", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	showLogs := fs.Bool("logs", false, "show logs")
	if err := fs.Parse(args); err != nil {
		return fmt.Errorf("failed to parse flags: %w", err)
	}

	var results []TaskResult
	status, err := getJSON(client, joinURL(serverURL, "/results"), &results)
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
		fmt.Println(colorBold + "Results with logs" + colorReset)
		for _, r := range results {
			fmt.Printf("Task %d | Exit %d | Created %s\n", r.TaskID, r.ExitCode, formatTime(r.CreatedAt))
			fmt.Printf("Result JSON: %s\n", r.ResultJSON)
			fmt.Printf("Stdout:\n%s\n", r.Stdout)
			fmt.Printf("Stderr:\n%s\n", r.Stderr)
			fmt.Printf("Logs:\n%s\n", r.Logs)
			fmt.Println(strings.Repeat("-", 80))
		}
		return nil
	}

	fmt.Println(colorBold + "Results" + colorReset)
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tTASK_ID\tEXIT_CODE\tRESULT_JSON\tCREATED_AT")
	for _, r := range results {
		fmt.Fprintf(w, "%d\t%d\t%d\t%s\t%s\n",
			r.ID,
			r.TaskID,
			r.ExitCode,
			trimForTable(r.ResultJSON, 60),
			formatTime(r.CreatedAt),
		)
	}
	return w.Flush()
}

func cmdCreateTask(client *http.Client, serverURL string, args []string) error {
	fs := flag.NewFlagSet("create-task", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	agentID := fs.Int64("agent", 0, "Agent ID")
	checkType := fs.String("check", "", "Check type")
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

	reqBody := createTaskRequest{
		AgentID:    *agentID,
		CheckType:  *checkType,
		Payload:    *payload,
		MaxRetries: *retries,
	}

	var out map[string]any
	status, err := postJSON(client, joinURL(serverURL, "/tasks"), reqBody, &out)
	if err != nil {
		return err
	}
	if status != http.StatusCreated {
		return fmt.Errorf("unexpected status: %d", status)
	}

	printSuccess("task created")
	prettyPrintJSON(out)
	return nil
}

func cmdRegisterAgent(client *http.Client, serverURL string, args []string) error {
	fs := flag.NewFlagSet("register-agent", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	name := fs.String("name", "", "Agent name")
	if err := fs.Parse(args); err != nil {
		return fmt.Errorf("failed to parse flags: %w", err)
	}
	if strings.TrimSpace(*name) == "" {
		return fmt.Errorf("--name is required")
	}

	reqBody := registerAgentRequest{Name: strings.TrimSpace(*name)}
	var out registerAgentResponse
	status, err := postJSON(client, joinURL(serverURL, "/agents/register"), reqBody, &out)
	if err != nil {
		return err
	}
	if status != http.StatusCreated {
		return fmt.Errorf("unexpected status: %d", status)
	}

	printSuccess("agent registered")
	fmt.Printf("Agent ID: %d\n", out.AgentID)
	fmt.Printf("Token:    %s\n", out.Token)
	return nil
}

func cmdRawGet(client *http.Client, serverURL string, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: get /path")
	}

	path := args[0]
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	var out any
	status, err := getJSON(client, joinURL(serverURL, path), &out)
	if err != nil {
		return err
	}

	fmt.Printf("GET %s [%d]\n", path, status)
	prettyPrintJSON(out)
	return nil
}

func cmdRawPost(client *http.Client, serverURL string, args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: post /path '{\"key\":\"value\"}'")
	}

	path := args[0]
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	var payload any
	if err := json.Unmarshal([]byte(args[1]), &payload); err != nil {
		return fmt.Errorf("invalid JSON body: %w", err)
	}

	var out any
	status, err := postJSON(client, joinURL(serverURL, path), payload, &out)
	if err != nil {
		return err
	}

	fmt.Printf("POST %s [%d]\n", path, status)
	prettyPrintJSON(out)
	return nil
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

func splitArgs(line string) []string {
	var args []string
	var current strings.Builder
	inQuotes := false
	var quoteChar rune

	for _, r := range line {
		switch {
		case r == '"' || r == '\'':
			if !inQuotes {
				inQuotes = true
				quoteChar = r
			} else if quoteChar == r {
				inQuotes = false
			} else {
				current.WriteRune(r)
			}
		case r == ' ' && !inQuotes:
			if current.Len() > 0 {
				args = append(args, current.String())
				current.Reset()
			}
		default:
			current.WriteRune(r)
		}
	}

	if current.Len() > 0 {
		args = append(args, current.String())
	}

	return args
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

func printSuccess(msg string) {
	fmt.Println(colorGreen + "✓ " + msg + colorReset)
}

func printError(msg string) {
	fmt.Println(colorRed + "✗ " + msg + colorReset)
}

func clearScreen() {
	fmt.Print("\033[H\033[2J")
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
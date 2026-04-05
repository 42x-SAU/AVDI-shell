package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	"diag-system/internal/deploy"
	"diag-system/internal/diagnostics"
	"diag-system/internal/httputil"

	"golang.org/x/term"
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

// scripts stores user-defined bash scripts by name
var scripts = make(map[string]string) // name -> content

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
	serverURL := httputil.NormalizeHTTPBaseURL(getenv("AVDI_SERVER", "http://localhost:8081"))
	client := &http.Client{Timeout: 20 * time.Second}
	aliases := newAliasStore()

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

		if strings.HasPrefix(line, "!") {
			if err := runOSCommand(strings.TrimPrefix(line, "!")); err != nil {
				printError(err.Error())
			}
			continue
		}

		lineIn := line
		if !isAliasCommandLine(line) {
			lineIn = aliases.apply(line)
		}

		args := splitArgs(lineIn)
		if len(args) == 0 {
			continue
		}

		switch args[0] {
		case "exit", "quit":
			fmt.Println(colorGray + "bye" + colorReset)
			return

		case "help":
			if len(args) == 1 {
				printHelp()
			} else {
				if !tryHelpCommands(args[1:]) {
					printError("unknown help topic: " + strings.Join(args[1:], " "))
					fmt.Println("Try: help help")
				}
			}

		case "alias":
			rest := ""
			if len(args) > 1 {
				rest = strings.Join(args[1:], " ")
			}
			if err := cmdAlias(aliases, rest); err != nil {
				printError(err.Error())
			}

		case "unalias":
			if len(args) < 2 {
				printError("usage: unalias <name>")
				continue
			}
			if err := cmdUnalias(aliases, args[1]); err != nil {
				printError(err.Error())
			}

		case "stats":
			if err := cmdStats(client, serverURL); err != nil {
				printError(err.Error())
			}

		case "stats-all":
			if err := cmdStatsAll(client, serverURL); err != nil {
				printError(err.Error())
			}

		case "server-add":
			if len(args) < 2 {
				printError("usage: server-add <url>")
				continue
			}
			if err := cmdServersAdd(strings.Join(args[1:], " ")); err != nil {
				printError(err.Error())
			}

		case "server-list":
			if err := cmdServersList(); err != nil {
				printError(err.Error())
			}

		case "recurring-list":
			if err := cmdRecurringList(client, serverURL); err != nil {
				printError(err.Error())
			}

		case "recurring-add":
			if err := cmdRecurringAdd(client, serverURL, args[1:]); err != nil {
				printError(err.Error())
			}

		case "recurring-delete":
			if err := cmdRecurringDelete(client, serverURL, args[1:]); err != nil {
				printError(err.Error())
			}

		case "recurring-enable":
			if err := cmdRecurringEnable(client, serverURL, args[1:], true); err != nil {
				printError(err.Error())
			}

		case "recurring-disable":
			if err := cmdRecurringEnable(client, serverURL, args[1:], false); err != nil {
				printError(err.Error())
			}

		case "clear":
			clearScreen()
			printBanner(serverURL)

		case "server":
			if len(args) < 2 {
				printError("usage: server <url>")
				continue
			}
			serverURL = httputil.NormalizeHTTPBaseURL(strings.TrimSpace(args[1]))
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

		case "export-logs":
			if err := cmdResults(client, serverURL, append([]string{"--export"}, args[1:]...)); err != nil {
				printError(err.Error())
			}

		case "create-task":
			if err := cmdCreateTask(client, serverURL, args[1:]); err != nil {
				printError(err.Error())
			}

		case "diagnostic-list":
			if err := cmdDiagnosticList(args[1:]); err != nil {
				printError(err.Error())
			}

		case "diagnostic-run":
			if err := cmdDiagnosticRun(args[1:]); err != nil {
				printError(err.Error())
			}

		case "deploy-agent":
			if err := cmdDeployAgent(args[1:]); err != nil {
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

		case "script":
			if len(args) < 2 {
				printError("usage: script add <name> <content> OR script list")
				continue
			}
			subcmd := args[1]
			switch subcmd {
			case "add":
				if err := cmdScriptAdd(args[2:]); err != nil {
					printError(err.Error())
				}
			case "list":
				if err := cmdScriptList(); err != nil {
					printError(err.Error())
				}
			default:
				printError("unknown script subcommand: " + subcmd)
			}

		case "run":
			if err := cmdRun(client, serverURL, args[1:]); err != nil {
				printError(err.Error())
			}

		default:
			printError("unknown command: " + args[0])
			fmt.Println("type 'help' to see available commands")
		}
	}
}

func isAliasCommandLine(line string) bool {
	a := splitArgs(line)
	if len(a) == 0 {
		return false
	}
	return a[0] == "alias" || a[0] == "unalias"
}

func printBanner(serverURL string) {
	printTurtleLogo()
	fmt.Println(colorBold + colorCyan + "AVDI shell" + colorReset)
	fmt.Println(colorGray + strings.Repeat("=", 60) + colorReset)
	fmt.Println("Current server:", colorYellow+serverURL+colorReset)
	fmt.Println(colorGray + "Type help or help <command> (e.g. help create-task)." + colorReset)
	fmt.Println()
}

func printTurtleLogo() {
	// Логотип (символьная графика), зелёный жирный в терминале с поддержкой Unicode.
	g := colorGreen + colorBold
	r := colorReset
	logo := `
 .


⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀
⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⣠⣾⣷⣄⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀
⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⢰⣿⣿⣿⣿⡆⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀
⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠘⣿⡿⢿⣿⠃⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀
⠀⠀⠀⠀⣠⣴⣾⣿⣿⣷⢀⠤⡒⣺⣭⣭⣗⢒⠤⡀⣾⣿⣿⣷⣦⣄⠀⠀⠀⠀
⠀⠀⢀⣾⣿⠿⠿⠿⡿⡕⠵⠿⢱⣿⣿⣿⣿⡎⠿⠮⢪⢿⠿⠿⠿⣿⣷⡀⠀⠀
⠀⠀⣾⠟⠁⠀⠀⠀⢸⣸⣿⣿⣷⢝⣛⣛⡫⣾⣿⣿⣇⡇⠀⠀⠀⠈⠻⣷⡀⠀
⠀⠸⠋⠀⠀⠀⠀⠀⢺⣛⣛⣛⡱⣿⣿⣿⣿⢎⣛⣛⣛⡗⠀⠀⠀⠀⠀⠙⠇⠀
⠀⠀⠀⠀⠀⠀⠀⠀⢸⢹⣿⣿⣿⣜⣛⣛⣣⣿⣿⣿⡏⡇⠀⠀⠀⠀⠀⠀⠀⠀
⠀⠀⠀⠀⠀⠀⠀⠀⠀⢧⠿⢿⡫⣾⣿⣿⣷⢝⡿⠿⡼⠀⠀⠀⠀⠀⠀⠀⠀⠀
⠀⠀⠀⠀⠀⠀⠀⠀⠀⢈⢮⣻⣿⢎⣭⣭⡱⣿⣟⡵⡁⠀⠀⠀⠀⠀⠀⠀⠀⠀
⠀⠀⠀⠀⠀⠀⠀⢀⣼⣿⣷⠕⢧⣻⠿⠿⣟⡬⠪⣾⣿⣧⡀⠀⠀⠀⠀⠀⠀⠀
⠀⠀⠀⠀⠀⠀⠀⣿⣿⣿⠋⠀⠀⠀⠉⠉⠀⠀⠀⠙⣿⣿⣿⠀⠀⠀⠀⠀⠀⠀
⠀⠀⠀⠀⠀⠀⠀⠈⢿⠇⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠸⡿⠁⠀⠀⠀⠀⠀⠀⠀
⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀
`
	fmt.Println()
	for _, line := range strings.Split(strings.TrimPrefix(logo, "\n"), "\n") {
		fmt.Println(g + line + r)
	}
	fmt.Println()
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
	fmt.Println("  results [--logs] [--export] [--output-dir <path>] [--limit N] [--task ID] [--result-id ID] [--exit-code N]")
	fmt.Println("      Show task results; --logs: full stdout/stderr/logs on screen")
	fmt.Println("      --export: write each result to files under output/ (see --output-dir)")
	fmt.Println("      --limit: only last N results (newest first); filters apply before export")
	fmt.Println("      --task / --result-id / --exit-code: optional filters (combine as needed)")
	fmt.Println("  export-logs [same flags as results --export]")
	fmt.Println("      Same as: results --export")
	fmt.Println()
	fmt.Println("  deploy-agent --ssh-host <host> --ssh-user <user> --server-url <url> --agent-name <name> --image <ref>")
	fmt.Println("      Deploy agent container on remote host over SSH (auto-registers on server)")
	fmt.Println("      Optional flags: --ssh-port <n> (default 22) | --ssh-password <pwd> | --container <name> | --skip-pull")
	fmt.Println()
	fmt.Println("  create-task --agent <id> --check <hostname|ping|ports|diagnostic> [--payload <value>] [--retries <n>]")
	fmt.Println("      Create a new diagnostic task (diagnostic check requires payload with command name)")
	fmt.Println()
	fmt.Println("  diagnostic-list [--config <path>] [--json]")
	fmt.Println("      List available diagnostic commands from config file")
	fmt.Println()
	fmt.Println("  diagnostic-run --command <name> [--config <path>] [--var key=value]... [--json]")
	fmt.Println("      Execute a diagnostic command with optional variables")
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
	fmt.Println("  stats")
	fmt.Println("      JSON summary: agents, tasks by status, results, recurring jobs (current server)")
	fmt.Println("  stats-all")
	fmt.Println("      Table of /stats for current server + URLs from server-list file")
	fmt.Println("  server-add <url>")
	fmt.Println("      Append server URL to ~/.config/avdi/servers.txt (for stats-all)")
	fmt.Println("  server-list")
	fmt.Println("      Show saved server URLs (# comments allowed)")
	fmt.Println()
	fmt.Println("  alias [name=value | name]")
	fmt.Println("      List aliases, show one, or set: alias my=create-task --agent 1 --check ping --payload 8.8.8.8")
	fmt.Println("  unalias <name>")
	fmt.Println("  !<shell command>")
	fmt.Println("      Run a system shell command (cmd /C on Windows, $SHELL -c on Unix)")
	fmt.Println()
	fmt.Println("  recurring-list")
	fmt.Println("      List scheduled jobs (POST /recurring on server)")
	fmt.Println("  recurring-add --agent N --check ping|hostname|ports|diagnostic [--payload ...] [--interval 60] [--retries 0]")
	fmt.Println("      Repeat: enqueue same check every --interval seconds (min 10, max 86400)")
	fmt.Println("  recurring-delete --id N")
	fmt.Println("  recurring-enable --id N | recurring-disable --id N")
	fmt.Println()
	fmt.Println("  help [command]")
	fmt.Println("      This list, or detailed help for one command (e.g. help create-task, help results)")
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
	fmt.Println("  results --export")
	fmt.Println("  results --export --limit 10")
	fmt.Println("  results --export --task 5 --limit 3")
	fmt.Println("  export-logs --limit 20")
	fmt.Println("  deploy-agent --ssh-host 192.168.1.100 --ssh-user root --server-url http://localhost:8081 --agent-name prod-agent1 --image avdi:latest")
	fmt.Println("  create-task --agent 3 --check hostname")
	fmt.Println("  create-task --agent 3 --check ping --payload 8.8.8.8")
	fmt.Println("  create-task --agent 3 --check ports")
	fmt.Println("  create-task --agent 3 --check diagnostic --payload '{\"command\":\"disk-usage\"}'")
	fmt.Println("  diagnostic-list")
	fmt.Println("  diagnostic-run --command hostname --json")
	fmt.Println("  diagnostic-run --command ping-target --var target=google.com")
	fmt.Println("  deploy-agent --ssh-host 192.168.1.100 --ssh-user root --server-url http://server:8081 --agent-name agent1 --image avdi:latest")
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
	doExport := fs.Bool("export", false, "write full logs to files under --output-dir")
	outDir := fs.String("output-dir", "output", "directory for --export (created if missing)")
	limitN := fs.Int("limit", 0, "fetch at most N newest results (0 = all, server caps at 10000)")
	taskID := fs.Int64("task", 0, "only results for this task id")
	resultID := fs.Int64("result-id", 0, "only this result row id")
	exitCodeStr := fs.String("exit-code", "", "filter by exit code (empty = no filter)")
	if err := fs.Parse(args); err != nil {
		return fmt.Errorf("failed to parse flags: %w", err)
	}

	var exitPtr *int
	if s := strings.TrimSpace(*exitCodeStr); s != "" {
		ec, err := strconv.Atoi(s)
		if err != nil {
			return fmt.Errorf("invalid --exit-code: %w", err)
		}
		exitPtr = &ec
	}

	apiURL := buildResultsAPIURL(serverURL, *limitN, *resultID, *taskID, exitPtr)

	var results []TaskResult
	status, err := getJSON(client, apiURL, &results)
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

	if *doExport {
		if err := exportResultsToDir(*outDir, results); err != nil {
			return err
		}
		printSuccess(fmt.Sprintf("exported %d result(s) to %s", len(results), *outDir))
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

	if *doExport {
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

func buildResultsAPIURL(serverURL string, limit int, resultID, taskID int64, exitCode *int) string {
	base := strings.TrimRight(httputil.NormalizeHTTPBaseURL(serverURL), "/") + "/results"
	v := url.Values{}
	if limit > 0 {
		v.Set("limit", strconv.Itoa(limit))
	}
	if resultID > 0 {
		v.Set("result_id", strconv.FormatInt(resultID, 10))
	}
	if taskID > 0 {
		v.Set("task_id", strconv.FormatInt(taskID, 10))
	}
	if exitCode != nil {
		v.Set("exit_code", strconv.Itoa(*exitCode))
	}
	if enc := v.Encode(); enc != "" {
		return base + "?" + enc
	}
	return base
}

func exportResultsToDir(dir string, results []TaskResult) error {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create output dir: %w", err)
	}
	for _, r := range results {
		name := filepath.Join(dir, fmt.Sprintf("result-%d-task-%d.txt", r.ID, r.TaskID))
		var b strings.Builder
		fmt.Fprintf(&b, "AVDI task result export\n")
		fmt.Fprintf(&b, "Result ID: %d\n", r.ID)
		fmt.Fprintf(&b, "Task ID: %d\n", r.TaskID)
		fmt.Fprintf(&b, "Exit code: %d\n", r.ExitCode)
		fmt.Fprintf(&b, "Created at: %s\n", formatTime(r.CreatedAt))
		fmt.Fprintf(&b, "\n=== result_json ===\n%s\n", r.ResultJSON)
		fmt.Fprintf(&b, "\n=== stdout ===\n%s\n", r.Stdout)
		fmt.Fprintf(&b, "\n=== stderr ===\n%s\n", r.Stderr)
		fmt.Fprintf(&b, "\n=== logs ===\n%s\n", r.Logs)
		if err := os.WriteFile(name, []byte(b.String()), 0644); err != nil {
			return fmt.Errorf("write %s: %w", name, err)
		}
	}
	return nil
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
	case "hostname", "ping", "ports", "diagnostic":
	default:
		return fmt.Errorf("unsupported --check value: %s", *checkType)
	}

	if *checkType == "diagnostic" && strings.TrimSpace(*payload) == "" {
		return fmt.Errorf(`--payload is required for --check diagnostic (JSON: {"command":"<name>"}[, "vars":{...}, "config":"..."])`)
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
		ServerURL:     httputil.NormalizeHTTPBaseURL(*serverURL),
		AgentName:     *agentName,
		Image:         *image,
		ContainerName: *container,
		SkipPull:      *skipPull,
	})
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
	return strings.TrimRight(httputil.NormalizeHTTPBaseURL(base), "/") + path
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

func cmdDiagnosticList(args []string) error {
	fs := flag.NewFlagSet("diagnostic-list", flag.ContinueOnError)
	configPath := fs.String("config", "", "Path to diagnostics config YAML (default: auto-discover)")
	outputJSON := fs.Bool("json", false, "Output in JSON format")

	if err := fs.Parse(args); err != nil {
		return fmt.Errorf("parse flags: %w", err)
	}

	config, err := diagnostics.LoadConfig(*configPath)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	if *outputJSON {
		prettyPrintJSON(config.Commands)
		return nil
	}

	fmt.Println("Available diagnostic commands:")
	for _, cmd := range config.Commands {
		fmt.Printf("  %-20s %s\n", cmd.Name, cmd.Description)
		if len(cmd.Variables) > 0 {
			fmt.Printf("    Variables: ")
			for i, v := range cmd.Variables {
				if i > 0 {
					fmt.Printf(", ")
				}
				fmt.Printf("%s", v.Name)
				if v.Required {
					fmt.Printf("(required)")
				}
			}
			fmt.Println()
		}
	}
	return nil
}

func cmdDiagnosticRun(args []string) error {
	fs := flag.NewFlagSet("diagnostic-run", flag.ContinueOnError)
	configPath := fs.String("config", "", "Path to diagnostics config YAML")
	commandName := fs.String("command", "", "Command name to execute (required)")
	varFlags := fs.String("var", "", "Variable in format key=value (can be repeated)")
	outputJSON := fs.Bool("json", true, "Output in JSON format (default true)")

	if err := fs.Parse(args); err != nil {
		return fmt.Errorf("parse flags: %w", err)
	}

	if *commandName == "" {
		return fmt.Errorf("--command is required")
	}

	// Parse variables
	variables := make(map[string]string)
	if *varFlags != "" {
		// Support multiple --var flags (need custom parsing)
		// For simplicity, assume single --var for now
		parts := strings.SplitN(*varFlags, "=", 2)
		if len(parts) != 2 {
			return fmt.Errorf("invalid variable format, expected key=value")
		}
		variables[parts[0]] = parts[1]
	}

	// Parse additional args as variables
	for _, arg := range fs.Args() {
		if strings.Contains(arg, "=") {
			parts := strings.SplitN(arg, "=", 2)
			variables[parts[0]] = parts[1]
		}
	}

	result, err := diagnostics.ExecuteDiagnostic(*configPath, *commandName, variables)
	if err != nil {
		return fmt.Errorf("execute diagnostic: %w", err)
	}

	if *outputJSON {
		prettyPrintJSON(result)
	} else {
		fmt.Printf("Command: %s %v\n", result.Command, result.Args)
		fmt.Printf("Exit code: %d\n", result.ExitCode)
		if result.Stdout != "" {
			fmt.Printf("Stdout:\n%s\n", result.Stdout)
		}
		if result.Stderr != "" {
			fmt.Printf("Stderr:\n%s\n", result.Stderr)
		}
		if result.Error != "" {
			fmt.Printf("Error: %s\n", result.Error)
		}
	}
	return nil
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

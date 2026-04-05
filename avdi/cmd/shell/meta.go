package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"text/tabwriter"

	"diag-system/internal/httputil"
)

func runOSCommand(cmdLine string) error {
	cmdLine = strings.TrimSpace(cmdLine)
	if cmdLine == "" {
		return fmt.Errorf("empty command")
	}
	var c *exec.Cmd
	if runtime.GOOS == "windows" {
		c = exec.Command("cmd", "/C", cmdLine)
	} else {
		shell := os.Getenv("SHELL")
		if shell == "" {
			shell = "/bin/sh"
		}
		c = exec.Command(shell, "-c", cmdLine)
	}
	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	return c.Run()
}

func avdiConfigDir() (string, error) {
	d, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(d, "avdi"), nil
}

func serversListPath() (string, error) {
	dir, err := avdiConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "servers.txt"), nil
}

func readServerList() []string {
	path, err := serversListPath()
	if err != nil {
		return nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	var out []string
	seen := make(map[string]struct{})
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		u := httputil.NormalizeHTTPBaseURL(line)
		if _, ok := seen[u]; ok {
			continue
		}
		seen[u] = struct{}{}
		out = append(out, u)
	}
	return out
}

func cmdServersAdd(urlStr string) error {
	urlStr = strings.TrimSpace(urlStr)
	if urlStr == "" {
		return fmt.Errorf("usage: server-add <url>")
	}
	u := httputil.NormalizeHTTPBaseURL(urlStr)
	dir, err := avdiConfigDir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	path := filepath.Join(dir, "servers.txt")
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	if _, err := fmt.Fprintln(f, u); err != nil {
		return err
	}
	printSuccess("saved server URL to " + path)
	return nil
}

func cmdServersList() error {
	path, err := serversListPath()
	if err != nil {
		return err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("(no servers file yet; use server-add)")
			return nil
		}
		return err
	}
	fmt.Print(string(data))
	return nil
}

type statsPayload struct {
	AgentsTotal     int64 `json:"agents_total"`
	TasksPending    int64 `json:"tasks_pending"`
	TasksRunning    int64 `json:"tasks_running"`
	TasksDone       int64 `json:"tasks_done"`
	TasksFailed     int64 `json:"tasks_failed"`
	TasksTotal      int64 `json:"tasks_total"`
	ResultsTotal    int64 `json:"results_total"`
	RecurringActive int64 `json:"recurring_enabled"`
	RecurringTotal  int64 `json:"recurring_total"`
}

func cmdStats(client *http.Client, serverURL string) error {
	var st statsPayload
	status, err := getJSON(client, joinURL(serverURL, "/stats"), &st)
	if err != nil {
		return err
	}
	if status != http.StatusOK {
		return fmt.Errorf("unexpected status: %d", status)
	}
	b, err := json.MarshalIndent(st, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(b))
	return nil
}

func cmdStatsAll(client *http.Client, currentURL string) error {
	list := readServerList()
	seen := make(map[string]struct{})
	var urls []string
	add := func(u string) {
		u = httputil.NormalizeHTTPBaseURL(u)
		if _, ok := seen[u]; ok {
			return
		}
		seen[u] = struct{}{}
		urls = append(urls, u)
	}
	add(currentURL)
	for _, u := range list {
		add(u)
	}
	fmt.Println(colorBold + "Stats per server" + colorReset)
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "SERVER\tAGENTS\tTASKS\tDONE\tFAIL\tRESULTS\tRECUR")
	for _, base := range urls {
		var st statsPayload
		status, err := getJSON(client, joinURL(base, "/stats"), &st)
		if err != nil {
			fmt.Fprintf(w, "%s\terror: %v\n", base, err)
			continue
		}
		if status != http.StatusOK {
			fmt.Fprintf(w, "%s\thttp %d\n", base, status)
			continue
		}
		fmt.Fprintf(w, "%s\t%d\t%d\t%d\t%d\t%d\t%d/%d\n",
			base, st.AgentsTotal, st.TasksTotal, st.TasksDone, st.TasksFailed, st.ResultsTotal, st.RecurringActive, st.RecurringTotal)
	}
	return w.Flush()
}

type recurringJob struct {
	ID              int64  `json:"id"`
	AgentID         int64  `json:"agent_id"`
	CheckType       string `json:"check_type"`
	Payload         string `json:"payload"`
	MaxRetries      int    `json:"max_retries"`
	IntervalSeconds int    `json:"interval_seconds"`
	Enabled         bool   `json:"enabled"`
	NextRunAt       string `json:"next_run_at"`
}

func cmdRecurringList(client *http.Client, serverURL string) error {
	var list []recurringJob
	status, err := getJSON(client, joinURL(serverURL, "/recurring"), &list)
	if err != nil {
		return err
	}
	if status != http.StatusOK {
		return fmt.Errorf("unexpected status: %d", status)
	}
	if len(list) == 0 {
		fmt.Println("No recurring jobs")
		return nil
	}
	b, _ := json.MarshalIndent(list, "", "  ")
	fmt.Println(string(b))
	return nil
}

func cmdRecurringAdd(client *http.Client, serverURL string, args []string) error {
	fs := flag.NewFlagSet("recurring-add", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	agentID := fs.Int64("agent", 0, "agent id")
	checkType := fs.String("check", "", "hostname|ping|ports|diagnostic|bash")
	payload := fs.String("payload", "", "payload string")
	interval := fs.Int("interval", 60, "seconds between runs (10-86400)")
	maxRetries := fs.Int("retries", 0, "max_retries for each spawned task")
	startIn := fs.Int("start-in", 0, "delay seconds before first enqueue")
	until := fs.String("until", "", "end time (RFC3339, e.g., 2026-04-05T23:59:59Z)")
	maxRuns := fs.Int("max-runs", 0, "maximum number of executions (0 = unlimited)")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *agentID < 1 || *checkType == "" {
		return fmt.Errorf("--agent and --check are required")
	}
	if *interval < 10 || *interval > 86400 {
		return fmt.Errorf("--interval must be between 10 and 86400")
	}
	body := map[string]any{
		"agent_id":         *agentID,
		"check_type":       *checkType,
		"payload":          *payload,
		"max_retries":      *maxRetries,
		"interval_seconds": *interval,
		"start_in_seconds": *startIn,
	}
	if *until != "" {
		body["until"] = *until
	}
	if *maxRuns > 0 {
		body["max_runs"] = *maxRuns
	}
	var out struct {
		ID int64 `json:"id"`
	}
	status, err := postJSON(client, joinURL(serverURL, "/recurring"), body, &out)
	if err != nil {
		return err
	}
	if status != http.StatusCreated {
		return fmt.Errorf("unexpected status: %d", status)
	}
	printSuccess(fmt.Sprintf("recurring job id=%d", out.ID))
	return nil
}

func cmdRecurringDelete(client *http.Client, serverURL string, args []string) error {
	fs := flag.NewFlagSet("recurring-delete", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	id := fs.Int64("id", 0, "recurring job id")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *id < 1 {
		return fmt.Errorf("--id is required")
	}
	req, err := http.NewRequest(http.MethodDelete, joinURL(serverURL, fmt.Sprintf("/recurring/%d", *id)), nil)
	if err != nil {
		return err
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return fmt.Errorf("%s: %s", resp.Status, strings.TrimSpace(string(body)))
	}
	printSuccess("deleted recurring job " + fmt.Sprint(*id))
	return nil
}

func cmdRecurringEnable(client *http.Client, serverURL string, args []string, enabled bool) error {
	fs := flag.NewFlagSet("recurring-enable", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	id := fs.Int64("id", 0, "recurring job id")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *id < 1 {
		return fmt.Errorf("--id is required")
	}
	buf := new(bytes.Buffer)
	_ = json.NewEncoder(buf).Encode(map[string]bool{"enabled": enabled})
	req, err := http.NewRequest(http.MethodPatch, joinURL(serverURL, fmt.Sprintf("/recurring/%d", *id)), buf)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return fmt.Errorf("%s: %s", resp.Status, strings.TrimSpace(string(body)))
	}
	printSuccess(fmt.Sprintf("recurring job %d enabled=%v", *id, enabled))
	return nil
}

// cmdScriptAdd adds a new bash script with a unique name
func cmdScriptAdd(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: script add <name> <content>")
	}
	name := args[0]
	content := strings.Join(args[1:], " ")
	
	// Check if script with this name already exists
	if _, exists := scripts[name]; exists {
		return fmt.Errorf("script '%s' already exists", name)
	}
	
	scripts[name] = content
	printSuccess(fmt.Sprintf("script '%s' added", name))
	return nil
}

// cmdScriptList lists all saved scripts
func cmdScriptList() error {
	if len(scripts) == 0 {
		fmt.Println("No scripts saved")
		return nil
	}
	fmt.Println(colorBold + "Saved scripts:" + colorReset)
	for name, content := range scripts {
		preview := content
		if len(preview) > 50 {
			preview = preview[:47] + "..."
		}
		fmt.Printf("  %s: %s\n", colorCyan+name+colorReset, preview)
	}
	return nil
}

// cmdRun executes a saved script on an agent
func cmdRun(client *http.Client, serverURL string, args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: run <script-name> --agent <agent-id> [--args arg1,arg2] [--env KEY=VAL,KEY2=VAL2]")
	}
	scriptName := args[0]
	
	// Look up script
	content, exists := scripts[scriptName]
	if !exists {
		return fmt.Errorf("script '%s' not found", scriptName)
	}
	
	// Parse remaining flags
	fs := flag.NewFlagSet("run", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	agentID := fs.Int64("agent", 0, "agent id")
	argsList := fs.String("args", "", "comma-separated arguments")
	env := fs.String("env", "", "comma-separated KEY=VALUE pairs")
	if err := fs.Parse(args[1:]); err != nil {
		return err
	}
	if *agentID < 1 {
		return fmt.Errorf("--agent is required")
	}
	
	// Parse arguments
	var argSlice []string
	if *argsList != "" {
		argSlice = strings.Split(*argsList, ",")
	}
	
	// Parse env
	envMap := make(map[string]string)
	if *env != "" {
		pairs := strings.Split(*env, ",")
		for _, pair := range pairs {
			kv := strings.SplitN(pair, "=", 2)
			if len(kv) == 2 {
				envMap[kv[0]] = kv[1]
			}
		}
	}
	
	// Build payload JSON
	payload := map[string]interface{}{
		"script": content,
	}
	if len(argSlice) > 0 {
		payload["args"] = argSlice
	}
	if len(envMap) > 0 {
		payload["env"] = envMap
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}
	
	// Create task
	body := map[string]any{
		"agent_id":    *agentID,
		"check_type":  "bash",
		"payload":     string(payloadBytes),
		"max_retries": 0,
	}
	var out struct {
		ID int64 `json:"id"`
	}
	status, err := postJSON(client, joinURL(serverURL, "/tasks"), body, &out)
	if err != nil {
		return err
	}
	if status != http.StatusCreated {
		return fmt.Errorf("unexpected status: %d", status)
	}
	printSuccess(fmt.Sprintf("task created id=%d", out.ID))
	return nil
}


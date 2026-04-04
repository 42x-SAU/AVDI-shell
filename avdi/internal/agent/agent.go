package agent

import (
	"log"
	"os"
	"strconv"
	"time"
)

type Agent struct {
	name      string
	serverURL string
	agentID   int64
	token     string
	client    *Client
}

type Task struct {
	ID        int64  `json:"id"`
	AgentID   int64  `json:"agent_id"`
	CheckType string `json:"check_type"`
	Payload   string `json:"payload"`
	Status    string `json:"status"`
}

type TaskResultRequest struct {
	TaskID     int64  `json:"task_id"`
	ExitCode   int    `json:"exit_code"`
	ResultJSON string `json:"result_json"`
	Stdout     string `json:"stdout"`
	Stderr     string `json:"stderr"`
	Logs       string `json:"logs"`
}

func NewFromEnv() (*Agent, error) {
	name := getenv("AGENT_NAME", "agent-1")
	serverURL := getenv("SERVER_URL", "http://server:8080")
	token := os.Getenv("AGENT_TOKEN")
	idStr := os.Getenv("AGENT_ID")

	var agentID int64
	if idStr != "" {
		parsed, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			return nil, err
		}
		agentID = parsed
	}

	return &Agent{
		name:      name,
		serverURL: serverURL,
		agentID:   agentID,
		token:     token,
		client:    NewClient(serverURL),
	}, nil
}

func (a *Agent) Run() error {
	if a.agentID == 0 || a.token == "" {
		id, token, err := a.client.Register(a.name)
		if err != nil {
			return err
		}
		a.agentID = id
		a.token = token
		log.Printf("registered agent id=%d (token set; not logged)", id)
	}

	go func() {
		ticker := time.NewTicker(15 * time.Second)
		defer ticker.Stop()

		for {
			if err := a.client.Heartbeat(a.agentID, a.token); err != nil {
				log.Printf("heartbeat error: %v", err)
			}
			<-ticker.C
		}
	}()

	for {
		task, status, err := a.client.NextTask(a.agentID, a.token)
		if err != nil {
			log.Printf("poll error: %v", err)
			time.Sleep(5 * time.Second)
			continue
		}

		if status == 204 || task == nil {
			time.Sleep(3 * time.Second)
			continue
		}

		res := RunCheck(task.CheckType, task.Payload)

		err = a.client.SubmitResult(a.agentID, a.token, TaskResultRequest{
			TaskID:     task.ID,
			ExitCode:   res.ExitCode,
			ResultJSON: res.ResultJSON,
			Stdout:     res.Stdout,
			Stderr:     res.Stderr,
			Logs:       res.Logs,
		})
		if err != nil {
			log.Printf("submit result error: %v", err)
		}
	}
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
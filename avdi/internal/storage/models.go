package storage

import "time"

type Agent struct {
    ID            int64     `json:"id"`
    Name          string    `json:"name"`
    Token         string    `json:"-"`
    LastHeartbeat time.Time `json:"last_heartbeat"`
    CreatedAt     time.Time `json:"created_at"`
}

type Task struct {
    ID          int64      `json:"id"`
    AgentID     int64      `json:"agent_id"`
    CheckType   string     `json:"check_type"`
    Payload     string     `json:"payload"`
    Status      string     `json:"status"`
    RetryCount  int        `json:"retry_count"`
    MaxRetries  int        `json:"max_retries"`
    CreatedAt   time.Time  `json:"created_at"`
    StartedAt   *time.Time `json:"started_at,omitempty"`
    FinishedAt  *time.Time `json:"finished_at,omitempty"`
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

type Script struct {
    ID        int64     `json:"id"`
    Name      string    `json:"name"`
    Content   string    `json:"content"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}
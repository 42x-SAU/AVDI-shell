package server

type registerAgentRequest struct {
    Name string `json:"name"`
}

type registerAgentResponse struct {
    AgentID int64  `json:"agent_id"`
    Token   string `json:"token"`
}

type createTaskRequest struct {
    AgentID    int64  `json:"agent_id"`
    CheckType  string `json:"check_type"`
    Payload    string `json:"payload"`
    MaxRetries int    `json:"max_retries"`
}

type resultRequest struct {
    TaskID      int64  `json:"task_id"`
    ExitCode    int    `json:"exit_code"`
    ResultJSON  string `json:"result_json"`
    Stdout      string `json:"stdout"`
    Stderr      string `json:"stderr"`
    Logs        string `json:"logs"`
}
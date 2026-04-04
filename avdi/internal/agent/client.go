package agent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type Client struct {
	baseURL string
	http    *http.Client
}

func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		http:    &http.Client{Timeout: 30 * time.Second},
	}
}

func (c *Client) Register(name string) (int64, string, error) {
	body, _ := json.Marshal(map[string]string{"name": name})

	resp, err := c.http.Post(c.baseURL+"/agents/register", "application/json", bytes.NewReader(body))
	if err != nil {
		return 0, "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		b, _ := io.ReadAll(resp.Body)
		return 0, "", fmt.Errorf("register failed: %s", string(b))
	}

	var out struct {
		AgentID int64  `json:"agent_id"`
		Token   string `json:"token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return 0, "", err
	}

	return out.AgentID, out.Token, nil
}

func (c *Client) Heartbeat(agentID int64, token string) error {
	body, _ := json.Marshal(map[string]int64{"agent_id": agentID})

	req, err := http.NewRequest(http.MethodPost, c.baseURL+"/agents/heartbeat", bytes.NewReader(body))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Agent-ID", fmt.Sprintf("%d", agentID))
	req.Header.Set("X-Agent-Token", token)

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("heartbeat failed: %s", string(b))
	}

	return nil
}

func (c *Client) NextTask(agentID int64, token string) (*Task, int, error) {
	req, err := http.NewRequest(http.MethodGet, c.baseURL+"/agents/tasks/next", nil)
	if err != nil {
		return nil, 0, err
	}

	req.Header.Set("X-Agent-ID", fmt.Sprintf("%d", agentID))
	req.Header.Set("X-Agent-Token", token)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNoContent {
		return nil, http.StatusNoContent, nil
	}

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return nil, resp.StatusCode, fmt.Errorf("next task failed: %s", string(b))
	}

	var t Task
	if err := json.NewDecoder(resp.Body).Decode(&t); err != nil {
		return nil, resp.StatusCode, err
	}

	return &t, resp.StatusCode, nil
}

func (c *Client) SubmitResult(agentID int64, token string, res TaskResultRequest) error {
	body, _ := json.Marshal(res)

	req, err := http.NewRequest(http.MethodPost, c.baseURL+"/agents/tasks/result", bytes.NewReader(body))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Agent-ID", fmt.Sprintf("%d", agentID))
	req.Header.Set("X-Agent-Token", token)

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("submit result failed: %s", string(b))
	}

	return nil
}
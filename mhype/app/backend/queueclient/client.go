package queueclient

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
)

type Consumer struct {
	Action        string `json:"action"`
	Workers       int    `json:"workers"`
	MaxAttempts   int    `json:"maxAttempts"`
	BackoffMillis []int  `json:"backoffMillis"`
	BufferSize    int    `json:"bufferSize"`
}

type Client struct {
	httpClient *http.Client
	backendURL string
	pluginID   string
}

func New(httpClient *http.Client) (*Client, error) {
	backendURL := os.Getenv("PLUGIN_BACKEND_URL")
	if backendURL == "" {
		return nil, errors.New("PLUGIN_BACKEND_URL is not set")
	}
	pluginID := os.Getenv("PLUGIN_ID")
	if pluginID == "" {
		return nil, errors.New("PLUGIN_ID is not set")
	}
	return &Client{
		httpClient: httpClient,
		backendURL: backendURL,
		pluginID:   pluginID,
	}, nil
}

func (c *Client) Register(ctx context.Context, name, description string, consumer Consumer) error {
	body, err := json.Marshal(map[string]any{
		"name":        name,
		"description": description,
		"consumer":    consumer,
	})
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		fmt.Sprintf("%s/api/plugins/%s/queues", c.backendURL, c.pluginID),
		bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	res, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusConflict {
		return nil
	}
	if res.StatusCode >= 300 {
		var env struct {
			Message string `json:"message"`
		}
		json.NewDecoder(res.Body).Decode(&env)
		if env.Message != "" {
			return fmt.Errorf("register queue: %s", env.Message)
		}
		return fmt.Errorf("register queue: HTTP %d", res.StatusCode)
	}
	return nil
}

func (c *Client) Publish(ctx context.Context, queueName string, payload any) (taskID string, err error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		fmt.Sprintf("%s/api/plugins/%s/queues/%s/publish", c.backendURL, c.pluginID, queueName),
		bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	res, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	var env struct {
		Message json.RawMessage `json:"message"`
	}
	if err := json.NewDecoder(res.Body).Decode(&env); err != nil {
		return "", fmt.Errorf("publish decode: %w", err)
	}

	if res.StatusCode >= 300 {
		var msg string
		json.Unmarshal(env.Message, &msg)
		if msg != "" {
			return "", fmt.Errorf("publish queue: %s", msg)
		}
		return "", fmt.Errorf("publish queue: HTTP %d", res.StatusCode)
	}

	var result struct {
		TaskID string `json:"taskId"`
	}
	if err := json.Unmarshal(env.Message, &result); err != nil {
		return "", fmt.Errorf("publish result decode: %w", err)
	}
	return result.TaskID, nil
}

func (c *Client) Deregister(ctx context.Context, name string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete,
		fmt.Sprintf("%s/api/plugins/%s/queues/%s", c.backendURL, c.pluginID, name), nil)
	if err != nil {
		return err
	}

	res, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusNotFound {
		return nil
	}
	if res.StatusCode >= 300 {
		var env struct {
			Message string `json:"message"`
		}
		json.NewDecoder(res.Body).Decode(&env)
		if env.Message != "" {
			return fmt.Errorf("deregister queue: %s", env.Message)
		}
		return fmt.Errorf("deregister queue: HTTP %d", res.StatusCode)
	}
	return nil
}

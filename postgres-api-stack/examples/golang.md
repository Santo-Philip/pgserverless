# Go Examples

```go
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const apiBase = "http://localhost:8080"

type Client struct {
	baseURL    string
	token      string
	httpClient *http.Client
}

func NewClient() *Client {
	return &Client{
		baseURL: apiBase,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *Client) SetToken(token string) {
	c.token = token
}

func (c *Client) do(method, path string, body interface{}, params map[string]string) (*http.Response, error) {
	var reqBody io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshal body: %w", err)
		}
		reqBody = bytes.NewReader(b)
	}

	req, err := http.NewRequest(method, c.baseURL+path, reqBody)
	if err != nil {
		return nil, fmt.Errorf("new request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	q := req.URL.Query()
	for k, v := range params {
		q.Set(k, v)
	}
	req.URL.RawQuery = q.Encode()

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}

	return resp, nil
}

type LoginRequest struct {
	PEmail    string `json:"p_email"`
	PPassword string `json:"p_password"`
}

type LoginResponse struct {
	Token  string `json:"token"`
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Name   string `json:"name"`
}

func (c *Client) Login(email, password string) (*LoginResponse, error) {
	resp, err := c.do("POST", "/rpc/login", LoginRequest{
		PEmail:    email,
		PPassword: password,
	}, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("login failed: HTTP %d", resp.StatusCode)
	}

	var result LoginResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode: %w", err)
	}

	c.SetToken(result.Token)
	return &result, nil
}

type RegisterRequest struct {
	PEmail    string `json:"p_email"`
	PPassword string `json:"p_password"`
	PName     string `json:"p_name,omitempty"`
}

func (c *Client) Register(email, password, name string) error {
	resp, err := c.do("POST", "/rpc/register", RegisterRequest{
		PEmail:    email,
		PPassword: password,
		PName:     name,
	}, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("register failed: HTTP %d", resp.StatusCode)
	}
	return nil
}

func (c *Client) GetUsers(limit, offset int) ([]map[string]interface{}, error) {
	resp, err := c.do("GET", "/users", nil, map[string]string{
		"order":  "created_at.desc",
		"limit":  fmt.Sprintf("%d", limit),
		"offset": fmt.Sprintf("%d", offset),
	})
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var users []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&users); err != nil {
		return nil, fmt.Errorf("decode users: %w", err)
	}
	return users, nil
}

type Organization struct {
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Description string `json:"description,omitempty"`
}

func (c *Client) CreateOrg(org Organization) (map[string]interface{}, error) {
	resp, err := c.do("POST", "/organizations", org, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode: %w", err)
	}
	return result, nil
}

func (c *Client) UpdateUser(id string, updates map[string]interface{}) error {
	resp, err := c.do("PATCH", fmt.Sprintf("/users?id=eq.%s", id), updates, nil)
	if err != nil {
		return err
	}
	resp.Body.Close()

	if resp.StatusCode != 204 && resp.StatusCode != 200 {
		return fmt.Errorf("update failed: HTTP %d", resp.StatusCode)
	}
	return nil
}

func (c *Client) DeleteUser(id string) error {
	resp, err := c.do("DELETE", fmt.Sprintf("/users?id=eq.%s", id), nil, nil)
	if err != nil {
		return err
	}
	resp.Body.Close()

	if resp.StatusCode != 204 {
		return fmt.Errorf("delete failed: HTTP %d", resp.StatusCode)
	}
	return nil
}

func (c *Client) Healthcheck() (map[string]interface{}, error) {
	resp, err := c.do("GET", "/rpc/healthcheck", nil, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode: %w", err)
	}
	return result, nil
}

func main() {
	client := NewClient()

	// Health check
	health, _ := client.Healthcheck()
	fmt.Printf("Health: %v\n", health)

	// Register
	err := client.Register("test@example.com", "password123", "Test User")
	if err != nil {
		fmt.Printf("Register error: %v\n", err)
	}

	// Login
	loginResp, err := client.Login("test@example.com", "password123")
	if err != nil {
		fmt.Printf("Login error: %v\n", err)
		return
	}
	fmt.Printf("Logged in as: %s (%s)\n", loginResp.Name, loginResp.Email)

	// Get users (requires auth)
	users, _ := client.GetUsers(10, 0)
	for _, u := range users {
		fmt.Printf("User: %v\n", u["email"])
	}
}
```

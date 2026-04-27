package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
)

const (
	Model           = "deepseek-v4-pro"
	ReasoningEffort = "medium"
	Stream          = false
)

type requestPayload struct {
	Model           string            `json:"model"`
	Messages        []message         `json:"messages"`
	Thinking        map[string]string `json:"thinking"`
	ReasoningEffort string            `json:"reasoning_effort"`
	Stream          bool              `json:"stream"`
}

type message struct {
	Role    string `json:"role"` // check if this can be limited to a set of options
	Content string `json:"content"`
}

func newMessage(content string) (message, error) {
	if content == "" {
		return message{}, errors.New("Empty strings are not allowed")
	}
	return message{Role: "user", Content: content}, nil
}

func newRequestPayload(messages []message) (requestPayload, error) {
	if len(messages) == 0 {
		return requestPayload{}, errors.New("Empty list of Messages is not allowed")
	}

	thinking := map[string]string{"type": "enabled"}
	return requestPayload{
		Model: Model, Messages: messages, Thinking: thinking,
		ReasoningEffort: ReasoningEffort,
	}, nil
}

func main() {
	apiKey := os.Getenv("DEEPSEEK_API_KEY")
	baseUrl := os.Getenv("DEEPSEEK_BASE_URL")

	if apiKey == "" {
		fmt.Println("DEEPSEEK_API_KEY is not set")
		os.Exit(1)
	}
	if baseUrl == "" {
		fmt.Println("DEEPSEEK_BASE_URL is not set")
		os.Exit(1)
	}

	contentStrings := [2]string{"Planning on building a system design flash cards app in golang.", "Can you suggest a plan to build it?"}

	var messages []message
	for _, content := range contentStrings {
		message, err := newMessage(content)
		if err != nil {
			fmt.Println("message creation failed", err)
		}
		messages = append(messages, message)
	}

	payloadData, err := newRequestPayload(messages)
	if err != nil {
		fmt.Println("newRequestPayload creation failed", err)
		os.Exit(1)
	}
	/*
		payload := `{
			"model": "deepseek-v4-pro",
			"messages": [{"role": "user", "content": "Planning on building a system design flash cards app in golang."},
				     {"role": "user", "content": "Can you suggest a plan to build it?"}],
			"thinking": {"type": "enabled"},
			"reasoning_effort": "medium",
			"stream": false
		}`
	*/

	payload, err := json.Marshal(payloadData)
	if err != nil {
		fmt.Println("marshalling payload data failed", err)
	}

	req, _ := http.NewRequest("POST", baseUrl, bytes.NewReader(payload))
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("api request failed", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body", err)
		os.Exit(1)
	}

	fmt.Println(string(body))
}

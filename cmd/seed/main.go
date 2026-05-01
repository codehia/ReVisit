package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"slices"
	"sync"
	"time"

	"github.com/codehia/goflash/internal/types"
	"go.uber.org/zap"
)

var (
	logger       = zap.Must(zap.NewDevelopment())
	sugar        = logger.Sugar()
	systemPrompt string
	httpClient   = &http.Client{Timeout: 72 * time.Second}
)

const (
	Model          = "deepseek-v4-flash"
	Stream         = false
	fakeUserPrompt = `
		Generate flashcards for this concept. Fully decompose it into atomic ideas and generate one card per atomic idea.
		Use the tag and tag_path exactly as given - do not modify them or create new tags.

		Concept: TTL
		Tag path: ["Caching", "Cache Eviction Policies"]
		Anchor: TTL (time-to-live) eviction associates each cache entry with an expiration timestamp. 
			Once the timestamp passes, the entry is treated as invalid: either purged eagerly by a background sweeper or lazily on the next read.
			TTL is independent of access patterns, unlike LRU or LFU. Redis EXPIRE and SET with EX option. Memcached entries take an explicit TTL.
			CDN cache headers (Cache-Control: max-age) implement TTL at the HTTP layer.
	`
	assistantPrompt = `
	{
	   "entries":[
	    {"tag":"TTL","tag_path":["Caching","Cache Eviction Policies"],
	   "cards":[ 
		{"question":"What is TTL-based cache eviction and how does it work?",
		 "answer":"TTL eviction associates each cache entry with an expiration timestamp. Once the timestamp passes the entry is invalid, purged either eagerly by a background sweeper or lazily on the next read. TTL is independent of access patterns unlike LRU or LFU, meaning a hot entry will still be evicted once it expires.",
		 "examples":"Redis EXPIRE and SET with EX option. Memcached entries take an explicit TTL. CDN cache headers (Cache-Control: max-age) implement TTL at the HTTP layer.",
		 "tradeoffs":"TTL is simple and predictable but evicts entries even when still hot, and keeps cold entries until they expire. Combine with LRU when both freshness and access patterns matter.",
		 "card_type":"definition"}, 
		{"question":"How does Redis implement TTL expiry internally?",
		 "answer":"Redis uses two strategies combined. Lazy expiry: when a key is accessed Redis checks if it has expired and deletes it before returning. Active expiry: a periodic background job samples random keys with TTLs and deletes expired ones. This hybrid avoids scanning all keys while still reclaiming memory from unaccessed expired entries.","examples":"Redis commands: EXPIRE key seconds, PEXPIRE key milliseconds, TTL key to inspect remaining time.",
		 "tradeoffs":"The sampling approach means expired but unaccessed keys can linger and consume memory. Under heavy expiry load the background job can cause latency spikes.",
		 "card_type":"mechanism"}, 
		{"question":"When would you choose TTL eviction over LRU?",
		 "answer":"Use TTL when correctness depends on freshness: session tokens, rate limit counters, DNS records, or pricing data that must not be served stale. Also when downstream contracts dictate expiry via HTTP Cache-Control headers or CDN edge caches. Use LRU when the goal is keeping the working set in memory regardless of age and stale data is acceptable.",
		 "examples":"Session stores in Redis use TTL matching the session lifetime. Rate limiters use short TTL windows (1s, 1m). Application data caches typically prefer LRU.",
		 "tradeoffs":"","card_type":"tradeoff"}]}
	   ]
	}
	`
)

type config struct {
	APIKey  string
	BaseURL string
}

type LeafNode struct {
	types.Node
	Path string
}

type requestPayload struct {
	Model          string            `json:"model"`
	Temperature    float64           `json:"temperature"`
	MaxTokens      int               `json:"max_tokens"`
	ResponseFormat map[string]string `json:"response_format"`
	Messages       []message         `json:"messages"`
	Thinking       map[string]string `json:"thinking"`
	Stream         bool              `json:"stream"`
}

type message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type entriesWrapper struct {
	Entries []types.Response `json:"entries"`
}

type Choice struct {
	Index   int     `json:"index"`
	Message message `json:"message"`
}

type APIResponse struct {
	Choices []Choice `json:"choices"`
}

func newMessage(content string, role string) (message, error) {
	if content == "" {
		return message{}, errors.New("empty strings are not allowed")
	}
	if role == "" {
		role = "user"
	}
	return message{Role: role, Content: content}, nil
}

func newRequestPayload(messages []message) (requestPayload, error) {
	if len(messages) == 0 {
		return requestPayload{}, errors.New("empty list of messages is not allowed")
	}

	thinking := map[string]string{"type": "disabled"}

	responseFormat := map[string]string{"type": "json_object"}
	return requestPayload{
		Model:          Model,
		Temperature:    0.0,
		MaxTokens:      32000,
		ResponseFormat: responseFormat,
		Messages:       messages,
		Thinking:       thinking,
		Stream:         Stream,
	}, nil
}

func getRootNode(path string) (types.Node, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return types.Node{}, fmt.Errorf("getRootNode: read %q: %w", path, err)
	}

	var root types.Node
	if err := json.Unmarshal(data, &root); err != nil {
		return types.Node{}, fmt.Errorf("getRootNode: parse json: %w", err)
	}

	return root, nil
}

func getLeafNodes(node types.Node, path string) []LeafNode {
	if len(node.Children) == 0 {
		return []LeafNode{{Node: node, Path: path}}
	}
	var childPath string
	if path == "" {
		childPath = node.Name
	} else {
		childPath = path + " > " + node.Name
	}
	var leaves []LeafNode
	for _, child := range node.Children {
		leaves = append(leaves, getLeafNodes(child, childPath)...)
	}
	return leaves
}

func findChildrenNodes(node types.Node, childrenNames []string) []types.Node {
	var childrenNodes []types.Node
	for _, child := range node.Children {
		if slices.Contains(childrenNames, child.Name) {
			childrenNodes = append(childrenNodes, child)
		}
	}
	return childrenNodes
}

func createUserMessage(leaf LeafNode) (message, error) {
	content := fmt.Sprintf(`
		Generate flashcards for this concept. Fully decompose it into atomic ideas and generate one card per atomic idea. 
		Use the tag and tag_path exactly as given - do not modify them or create new tags.
		
		Concept: %s
		Tag path: %s
		Anchor: %s`, leaf.Name, leaf.Path, leaf.Notes)
	return newMessage(content, "user")
}

func createFakeUserMessage() (message, error) {
	return newMessage(fakeUserPrompt, "user")
}

func createAssistantMessage() (message, error) {
	return newMessage(assistantPrompt, "assistant")
}

func createSystemMessage() (message, error) {
	return newMessage(systemPrompt, "system")
}

func createPayload(node LeafNode) (requestPayload, error) {
	systemMessage, err := createSystemMessage()
	if err != nil {
		return requestPayload{}, err
	}
	userMessage, err := createUserMessage(node)
	if err != nil {
		return requestPayload{}, err
	}
	fakeUserMessage, err := createFakeUserMessage()
	if err != nil {
		return requestPayload{}, err
	}
	assistantMessage, err := createAssistantMessage()
	if err != nil {
		return requestPayload{}, err
	}

	messages := []message{systemMessage, fakeUserMessage, assistantMessage, userMessage}
	payload, err := newRequestPayload(messages)
	if err != nil {
		return requestPayload{}, err
	}

	return payload, nil
}

func makeRequest(payloadData requestPayload, cfg config, results chan<- []types.Response, retry chan<- requestPayload, wg *sync.WaitGroup) {
	payload, err := json.Marshal(payloadData)
	if err != nil {
		sugar.Errorw("marshalling payload failed", "error", err)
		wg.Done()
		return
	}

	sugar.Infow("sending request", "model", payloadData.Model)
	req, err := http.NewRequest("POST", cfg.BaseURL, bytes.NewReader(payload))
	if err != nil {
		sugar.Errorw("failed to create request", "error", err)
		wg.Add(1)
		retry <- payloadData
		wg.Done()
		return
	}
	req.Header.Set("Authorization", "Bearer "+cfg.APIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		sugar.Warnw("http request failed, retrying", "error", err)
		wg.Add(1)
		retry <- payloadData
		wg.Done()
		return
	}
	defer resp.Body.Close() //nolint:errcheck

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		sugar.Warnw("reading response body failed, retrying", "error", err)
		wg.Add(1)
		retry <- payloadData
		wg.Done()
		return
	}

	responses, err := validateResponse(body)
	if err != nil {
		sugar.Warnw("validation failed, retrying", "error", err)
		wg.Add(1)
		retry <- payloadData
		wg.Done()
		return
	}

	sugar.Infow("request successful", "responses", len(responses))
	results <- responses
	wg.Done()
}

func validateResponse(response []byte) ([]types.Response, error) {
	apiResponse := APIResponse{}
	if err := json.Unmarshal(response, &apiResponse); err != nil {
		return nil, err
	}
	if len(apiResponse.Choices) == 0 {
		return nil, fmt.Errorf("API returned no choices, body: %s", string(response))
	}

	var responses []types.Response

	for _, choice := range apiResponse.Choices {
		var wrapper entriesWrapper
		if err := json.Unmarshal([]byte(choice.Message.Content), &wrapper); err != nil {
			return nil, err
		}
		for _, entry := range wrapper.Entries {
			if err := entry.Validate(); err != nil {
				return nil, err
			}
		}
		responses = append(responses, wrapper.Entries...)

	}
	return responses, nil
}

func writeToResultJson(path string, results []types.Response) {
	sugar.Infow("writing results to file", "path", path, "count", len(results))
	jsonFileData, err := os.ReadFile(path)

	if err == nil && len(jsonFileData) > 0 {
		var responsesFromJson []types.Response
		err := json.Unmarshal(jsonFileData, &responsesFromJson)
		if err != nil {
			sugar.Errorw("error reading file", "path", path, "error", err)
			return
		}
		results = append(results, responsesFromJson...)

		seen := map[string]bool{}
		var distinct []types.Response
		for _, result := range results {
			key, err := json.Marshal(result)
			if err != nil {
				sugar.Errorw("failed to marshal response for dedup", "tag", result.Tag, "error", err)
				continue
			}
			if !seen[string(key)] {
				seen[string(key)] = true
				distinct = append(distinct, result)
			}
		}
		results = distinct

	}

	marshaledResultData, err := json.MarshalIndent(results, "", " ")
	if err != nil {
		sugar.Errorw("failed to marshal results", "error", err)
		return
	}

	err = os.WriteFile(path, marshaledResultData, 0o644)
	if err != nil {
		sugar.Errorw("failed to write file", "path", path, "error", err)
		return
	}
	sugar.Infow("results written successfully", "path", path)
}

func main() {
	defer logger.Sync() //nolint:errcheck
	cfg := config{
		APIKey:  os.Getenv("DEEPSEEK_API_KEY"),
		BaseURL: os.Getenv("DEEPSEEK_BASE_URL"),
	}
	if cfg.APIKey == "" {
		sugar.Errorw("DEEPSEEK_API_KEY is not set")
		os.Exit(1)
	}
	if cfg.BaseURL == "" {
		sugar.Errorw("DEEPSEEK_BASE_URL is not set")
		os.Exit(1)
	}

	systemPromptFileData, err := os.ReadFile("systemPrompt.txt")
	if err != nil {
		sugar.Errorw("failed to load the systemPrompt", "error", err)
		os.Exit(1)
	}
	systemPrompt = string(systemPromptFileData)

	children := os.Args[1:]
	if len(children) == 0 {
		sugar.Errorw("no children specified")
		os.Exit(1)
	}

	root, err := getRootNode("system-design-hierarchy.json")
	if err != nil {
		sugar.Errorw("failed to read root node", "error", err)
		os.Exit(1)
	}

	childrenNodes := findChildrenNodes(root, children)
	var leafNodes []LeafNode
	for _, node := range childrenNodes {
		leafNodes = append(leafNodes, getLeafNodes(node, "")...)
	}
	if len(leafNodes) == 0 {
		sugar.Errorw("no matching nodes found", "args", children)
		os.Exit(1)
	}
	sugar.Infow("found leaf nodes", "count", len(leafNodes))

	results := make(chan []types.Response)
	retry := make(chan requestPayload)

	var wg sync.WaitGroup
	for _, node := range leafNodes {
		payload, err := createPayload(node)
		if err != nil {
			sugar.Errorw("failed to create payload", "node", node.Name, "error", err)
			continue
		}
		sugar.Infow("launching request", "node", node.Name)
		wg.Add(1) // only count goroutines that actually launch
		go makeRequest(payload, cfg, results, retry, &wg)
	}

	go func() {
		wg.Wait()
		sugar.Infow("all requests done, closing channels")
		close(results)
		close(retry)
	}()
	go func() {
		for r := range retry {
			sugar.Infow("retrying request", "model", r.Model)
			// TODO: no retry limit — a permanently failing request retries forever.
			// Fix: track attempt count per payload and drop after N retries.
			go makeRequest(r, cfg, results, retry, &wg)
		}
	}()

	var collected []types.Response
	for r := range results {
		collected = append(collected, r...)
	}
	sugar.Infow("collection complete", "total_responses", len(collected))
	writeToResultJson("output.json", collected)
}

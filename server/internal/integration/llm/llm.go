// Package llm is a small client for the Gemini generateContent and Groq chat completions REST APIs.
package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"the-slot-finder/server/internal/config"
	"the-slot-finder/server/internal/constants"
	"the-slot-finder/server/internal/errors"
)

// Content is one turn of the conversation.
type Content struct {
	Role  string `json:"role,omitempty"`
	Parts []Part `json:"parts"`
}

// Part holds text, a tool call from the model, or a tool result. ThoughtSignature is sent back
// as received, so models that reason between tool calls keep their context.
type Part struct {
	Text             string            `json:"text,omitempty"`
	FunctionCall     *FunctionCall     `json:"functionCall,omitempty"`
	FunctionResponse *FunctionResponse `json:"functionResponse,omitempty"`
	ThoughtSignature string            `json:"thoughtSignature,omitempty"`
}

// FunctionCall is a tool the model wants to run.
type FunctionCall struct {
	Name string         `json:"name"`
	Args map[string]any `json:"args,omitempty"`
}

// FunctionResponse is a tool result for the model.
type FunctionResponse struct {
	Name     string         `json:"name"`
	Response map[string]any `json:"response"`
}

// FunctionDeclaration describes one tool; Parameters is a JSON schema subset.
type FunctionDeclaration struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Parameters  map[string]any `json:"parameters,omitempty"`
}

// Tool groups function declarations.
type Tool struct {
	FunctionDeclarations []FunctionDeclaration `json:"functionDeclarations"`
}

// Request is one generateContent call.
type Request struct {
	SystemInstruction *Content  `json:"systemInstruction,omitempty"`
	Contents          []Content `json:"contents"`
	Tools             []Tool    `json:"tools,omitempty"`
}

// Response is the first candidate's content; empty when the model returned none.
type Response struct {
	Content Content
}

// Client generates one model reply.
type Client interface {
	Generate(ctx context.Context, req Request) (*Response, error)
}

// RateLimitedError is returned when the model service answers HTTP 429. RetryAfter is in whole
// seconds and 0 when the service sent no usable Retry-After header.
type RateLimitedError struct {
	RetryAfter int
}

func (e *RateLimitedError) Error() string { return "model service rate limited" }

// ProvideClient creates the client for LLM_PROVIDER. A blank key is allowed: the service
// reports the feature as off, so the server still starts. An unknown provider fails here.
func ProvideClient(cfg *config.Config) (Client, error) {
	httpClient := &http.Client{Timeout: cfg.LLMTimeout}
	switch cfg.LLMProvider {
	case constants.LLMProviderGemini:
		return &geminiClient{
			http:   httpClient,
			apiKey: cfg.LLMAPIKey,
			url:    constants.LLMBaseURL + cfg.LLMModel + constants.LLMGenerateMethod,
		}, nil
	case constants.LLMProviderGroq:
		return &groqClient{http: httpClient, apiKey: cfg.LLMAPIKey, model: cfg.LLMModel}, nil
	default:
		return nil, errors.InternalError("Invalid LLM provider", fmt.Errorf("invalid LLM provider: %s", cfg.LLMProvider)).
			WithOperation("initialize_llm_client").
			WithResource("llm").
			WithContext("llm_provider", cfg.LLMProvider)
	}
}

// postJSON sends in as JSON with one auth header and decodes a 200 reply into out. The error
// never holds the auth value or the body sent.
func postJSON(ctx context.Context, hc *http.Client, url, authHeader, authValue string, in, out any) error {
	body, err := json.Marshal(in)
	if err != nil {
		return fmt.Errorf("encode request: %w", err)
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set(authHeader, authValue)
	resp, err := hc.Do(httpReq)
	if err != nil {
		// The URL error holds the request URL only, never the key header.
		return fmt.Errorf("call model: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode == http.StatusTooManyRequests {
		retryAfter, _ := strconv.Atoi(resp.Header.Get("Retry-After"))
		return &RateLimitedError{RetryAfter: max(retryAfter, 0)}
	}
	if resp.StatusCode != http.StatusOK {
		snippet, _ := io.ReadAll(io.LimitReader(resp.Body, constants.LLMMaxErrorBodyBytes))
		return fmt.Errorf("model returned status %d: %s", resp.StatusCode, snippet)
	}
	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		return fmt.Errorf("decode reply: %w", err)
	}
	return nil
}

type geminiClient struct {
	http   *http.Client
	apiKey string
	url    string
}

type generateReply struct {
	Candidates []struct {
		Content Content `json:"content"`
	} `json:"candidates"`
}

func (c *geminiClient) Generate(ctx context.Context, req Request) (*Response, error) {
	var reply generateReply
	if err := postJSON(ctx, c.http, c.url, constants.LLMAPIKeyHeader, c.apiKey, req, &reply); err != nil {
		return nil, err
	}
	if len(reply.Candidates) == 0 {
		return &Response{}, nil
	}
	return &Response{Content: reply.Candidates[0].Content}, nil
}

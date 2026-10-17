package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"the-slot-finder/server/internal/constants"
)

// Groq and OpenRouter speak the OpenAI chat completions shape. These types are only its wire format.
type openaiMessage struct {
	Role       string           `json:"role"`
	Content    string           `json:"content,omitempty"`
	ToolCalls  []openaiToolCall `json:"tool_calls,omitempty"`
	ToolCallID string           `json:"tool_call_id,omitempty"`
}

type openaiToolCall struct {
	ID       string             `json:"id"`
	Type     string             `json:"type"`
	Function openaiFunctionCall `json:"function"`
}

type openaiFunctionCall struct {
	Name string `json:"name"`
	// Arguments is a JSON object encoded as a string.
	Arguments string `json:"arguments"`
}

type openaiTool struct {
	Type     string              `json:"type"`
	Function FunctionDeclaration `json:"function"`
}

type openaiRequest struct {
	Model           string          `json:"model"`
	Messages        []openaiMessage `json:"messages"`
	Tools           []openaiTool    `json:"tools,omitempty"`
	ReasoningEffort string          `json:"reasoning_effort,omitempty"`
	Temperature     float64         `json:"temperature"`
}

type openaiReply struct {
	Choices []struct {
		Message openaiMessage `json:"message"`
	} `json:"choices"`
}

// openaiClient talks to one OpenAI-compatible endpoint. lowReasoning is Groq only: it sends
// reasoning_effort for the gpt-oss models, which OpenRouter does not take under that name.
type openaiClient struct {
	http         *http.Client
	url          string
	apiKey       string
	model        string
	temperature  float64
	lowReasoning bool
}

func (c *openaiClient) Generate(ctx context.Context, req Request) (*Response, error) {
	body, err := toOpenAIRequest(c, req)
	if err != nil {
		return nil, err
	}
	var reply openaiReply
	if err := postJSON(ctx, c.http, c.url, constants.OpenAIAuthHeader,
		constants.OpenAIAuthScheme+c.apiKey, body, &reply); err != nil {
		return nil, err
	}
	if len(reply.Choices) == 0 {
		return &Response{}, nil
	}
	return fromOpenAIMessage(reply.Choices[0].Message)
}

func toOpenAIRequest(c *openaiClient, req Request) (openaiRequest, error) {
	out := openaiRequest{Model: c.model, Temperature: c.temperature}
	if c.lowReasoning && strings.HasPrefix(c.model, constants.GroqReasoningModelPrefix) {
		out.ReasoningEffort = constants.GroqReasoningEffortLow
	}
	if req.SystemInstruction != nil {
		out.Messages = append(out.Messages, openaiMessage{Role: constants.OpenAIRoleSystem, Content: joinText(req.SystemInstruction.Parts)})
	}
	// pending holds the calls of the last assistant message that have no result yet.
	var pending []openaiToolCall
	for i, content := range req.Contents {
		if content.Role == constants.LLMRoleModel {
			msg, err := toOpenAIAssistant(i, content)
			if err != nil {
				return out, err
			}
			pending = msg.ToolCalls
			out.Messages = append(out.Messages, msg)
			continue
		}
		toolMsgs, err := toOpenAIToolMessages(content, &pending)
		if err != nil {
			return out, err
		}
		out.Messages = append(out.Messages, toolMsgs...)
		if text := joinText(content.Parts); text != "" {
			out.Messages = append(out.Messages, openaiMessage{Role: constants.OpenAIRoleUser, Content: text})
		}
	}
	for _, tool := range req.Tools {
		for _, decl := range tool.FunctionDeclarations {
			decl.Parameters = lowerTypes(decl.Parameters)
			out.Tools = append(out.Tools, openaiTool{Type: constants.OpenAIToolTypeFunc, Function: decl})
		}
	}
	return out, nil
}

// toOpenAIAssistant gives each tool call the id call_<content index>_<call index>, so the same
// history always yields the same ids.
func toOpenAIAssistant(contentIndex int, content Content) (openaiMessage, error) {
	msg := openaiMessage{Role: constants.OpenAIRoleAssistant, Content: joinText(content.Parts)}
	for _, part := range content.Parts {
		if part.FunctionCall == nil {
			continue
		}
		args, err := json.Marshal(part.FunctionCall.Args)
		if err != nil {
			return msg, fmt.Errorf("encode tool call arguments: %w", err)
		}
		msg.ToolCalls = append(msg.ToolCalls, openaiToolCall{
			ID:       fmt.Sprintf("call_%d_%d", contentIndex, len(msg.ToolCalls)),
			Type:     constants.OpenAIToolTypeFunc,
			Function: openaiFunctionCall{Name: part.FunctionCall.Name, Arguments: string(args)},
		})
	}
	return msg, nil
}

// toOpenAIToolMessages answers each function response with the id of the first unanswered call of
// the same name from the assistant message before it.
func toOpenAIToolMessages(content Content, pending *[]openaiToolCall) ([]openaiMessage, error) {
	var msgs []openaiMessage
	for _, part := range content.Parts {
		if part.FunctionResponse == nil {
			continue
		}
		id := ""
		for i, call := range *pending {
			if call.Function.Name == part.FunctionResponse.Name {
				id = call.ID
				*pending = append((*pending)[:i:i], (*pending)[i+1:]...)
				break
			}
		}
		if id == "" {
			return nil, fmt.Errorf("tool result for %q has no matching call", part.FunctionResponse.Name)
		}
		data, err := json.Marshal(part.FunctionResponse.Response)
		if err != nil {
			return nil, fmt.Errorf("encode tool result: %w", err)
		}
		msgs = append(msgs, openaiMessage{Role: constants.OpenAIRoleTool, ToolCallID: id, Content: string(data)})
	}
	return msgs, nil
}

func fromOpenAIMessage(msg openaiMessage) (*Response, error) {
	var parts []Part
	if msg.Content != "" {
		parts = append(parts, Part{Text: msg.Content})
	}
	for _, call := range msg.ToolCalls {
		args := map[string]any{}
		if strings.TrimSpace(call.Function.Arguments) != "" {
			if err := json.Unmarshal([]byte(call.Function.Arguments), &args); err != nil {
				return nil, fmt.Errorf("decode arguments of tool call %q: %w", call.Function.Name, err)
			}
		}
		parts = append(parts, Part{FunctionCall: &FunctionCall{Name: call.Function.Name, Args: args}})
	}
	return &Response{Content: Content{Parts: parts}}, nil
}

func joinText(parts []Part) string {
	var sb strings.Builder
	for _, part := range parts {
		sb.WriteString(part.Text)
	}
	return sb.String()
}

// lowerTypes copies a schema with every "type" value lower-cased: Gemini writes STRING and
// OBJECT, JSON schema wants string and object.
func lowerTypes(schema map[string]any) map[string]any {
	if schema == nil {
		return nil
	}
	out := make(map[string]any, len(schema))
	for key, value := range schema {
		switch v := value.(type) {
		case string:
			if key == "type" {
				v = strings.ToLower(v)
			}
			out[key] = v
		case map[string]any:
			out[key] = lowerTypes(v)
		default:
			out[key] = value
		}
	}
	return out
}

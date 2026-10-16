package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"the-slot-finder/server/internal/constants"
)

// Groq speaks the OpenAI chat completions shape. These types are only its wire format.
type groqMessage struct {
	Role       string         `json:"role"`
	Content    string         `json:"content,omitempty"`
	ToolCalls  []groqToolCall `json:"tool_calls,omitempty"`
	ToolCallID string         `json:"tool_call_id,omitempty"`
}

type groqToolCall struct {
	ID       string           `json:"id"`
	Type     string           `json:"type"`
	Function groqFunctionCall `json:"function"`
}

type groqFunctionCall struct {
	Name string `json:"name"`
	// Arguments is a JSON object encoded as a string.
	Arguments string `json:"arguments"`
}

type groqTool struct {
	Type     string              `json:"type"`
	Function FunctionDeclaration `json:"function"`
}

type groqRequest struct {
	Model           string        `json:"model"`
	Messages        []groqMessage `json:"messages"`
	Tools           []groqTool    `json:"tools,omitempty"`
	ReasoningEffort string        `json:"reasoning_effort,omitempty"`
}

type groqReply struct {
	Choices []struct {
		Message groqMessage `json:"message"`
	} `json:"choices"`
}

type groqClient struct {
	http   *http.Client
	apiKey string
	model  string
}

func (c *groqClient) Generate(ctx context.Context, req Request) (*Response, error) {
	body, err := toGroqRequest(c.model, req)
	if err != nil {
		return nil, err
	}
	var reply groqReply
	if err := postJSON(ctx, c.http, constants.GroqChatURL, constants.GroqAuthHeader,
		constants.GroqAuthScheme+c.apiKey, body, &reply); err != nil {
		return nil, err
	}
	if len(reply.Choices) == 0 {
		return &Response{}, nil
	}
	return fromGroqMessage(reply.Choices[0].Message)
}

func toGroqRequest(model string, req Request) (groqRequest, error) {
	out := groqRequest{Model: model}
	if strings.HasPrefix(model, constants.GroqReasoningModelPrefix) {
		out.ReasoningEffort = constants.GroqReasoningEffortLow
	}
	if req.SystemInstruction != nil {
		out.Messages = append(out.Messages, groqMessage{Role: constants.GroqRoleSystem, Content: joinText(req.SystemInstruction.Parts)})
	}
	// pending holds the calls of the last assistant message that have no result yet.
	var pending []groqToolCall
	for i, content := range req.Contents {
		if content.Role == constants.LLMRoleModel {
			msg, err := toGroqAssistant(i, content)
			if err != nil {
				return out, err
			}
			pending = msg.ToolCalls
			out.Messages = append(out.Messages, msg)
			continue
		}
		toolMsgs, err := toGroqToolMessages(content, &pending)
		if err != nil {
			return out, err
		}
		out.Messages = append(out.Messages, toolMsgs...)
		if text := joinText(content.Parts); text != "" {
			out.Messages = append(out.Messages, groqMessage{Role: constants.GroqRoleUser, Content: text})
		}
	}
	for _, tool := range req.Tools {
		for _, decl := range tool.FunctionDeclarations {
			decl.Parameters = lowerTypes(decl.Parameters)
			out.Tools = append(out.Tools, groqTool{Type: constants.GroqToolTypeFunc, Function: decl})
		}
	}
	return out, nil
}

// toGroqAssistant gives each tool call the id call_<content index>_<call index>, so the same
// history always yields the same ids.
func toGroqAssistant(contentIndex int, content Content) (groqMessage, error) {
	msg := groqMessage{Role: constants.GroqRoleAssistant, Content: joinText(content.Parts)}
	for _, part := range content.Parts {
		if part.FunctionCall == nil {
			continue
		}
		args, err := json.Marshal(part.FunctionCall.Args)
		if err != nil {
			return msg, fmt.Errorf("encode tool call arguments: %w", err)
		}
		msg.ToolCalls = append(msg.ToolCalls, groqToolCall{
			ID:       fmt.Sprintf("call_%d_%d", contentIndex, len(msg.ToolCalls)),
			Type:     constants.GroqToolTypeFunc,
			Function: groqFunctionCall{Name: part.FunctionCall.Name, Arguments: string(args)},
		})
	}
	return msg, nil
}

// toGroqToolMessages answers each function response with the id of the first unanswered call of
// the same name from the assistant message before it.
func toGroqToolMessages(content Content, pending *[]groqToolCall) ([]groqMessage, error) {
	var msgs []groqMessage
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
		msgs = append(msgs, groqMessage{Role: constants.GroqRoleTool, ToolCallID: id, Content: string(data)})
	}
	return msgs, nil
}

func fromGroqMessage(msg groqMessage) (*Response, error) {
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

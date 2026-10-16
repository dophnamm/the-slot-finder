package constants

// Ask AI (Google Gemini or Groq, chosen with LLM_PROVIDER). Each provider has its own key and model variables.
const (
	// DefaultGeminiModel is the Gemini default; model IDs change (older ones close to new accounts), so confirm it in AI Studio.
	DefaultGeminiModel = "gemini-3.5-flash"
	// DefaultGroqModel is the Groq default; model IDs change, so confirm it in the Groq console.
	DefaultGroqModel = "openai/gpt-oss-120b"
	// GroqChatURL is Groq's OpenAI-compatible chat completions endpoint.
	GroqChatURL    = "https://api.groq.com/openai/v1/chat/completions"
	GroqAuthHeader = "Authorization"
	GroqAuthScheme = "Bearer "
	// GroqReasoningModelPrefix marks the models that accept reasoning_effort; a low effort spends fewer tokens per answer.
	GroqReasoningModelPrefix = "openai/gpt-oss"
	GroqReasoningEffortLow   = "low"
	GroqRoleSystem           = "system"
	GroqRoleUser             = "user"
	GroqRoleAssistant        = "assistant"
	GroqRoleTool             = "tool"
	GroqToolTypeFunc         = "function"
	// LLMBaseURL is the Gemini REST endpoint; the model and :generateContent follow it.
	LLMBaseURL        = "https://generativelanguage.googleapis.com/v1beta/models/"
	LLMGenerateMethod = ":generateContent"
	LLMAPIKeyHeader   = "x-goog-api-key"

	LLMRoleUser  = "user"
	LLMRoleModel = "model"

	// LLMToolCutMarker ends a tool result that was cut at LLM_MAX_TOOL_RESULT_BYTES.
	LLMToolCutMarker = "...[data cut: the result was too long, ask for a narrower range or one program]"

	// LLMMaxErrorBodyBytes is how much of an error reply is kept in the error.
	LLMMaxErrorBodyBytes = 512
)

// Tools the model may call. Each one reads through an existing service, so roles apply.
const (
	AIToolHeatmap         = "get_heatmap"
	AIToolRisks           = "get_risks"
	AIToolOpenDrafts      = "list_open_drafts"
	AIToolProject         = "get_project"
	AIToolProgramOverview = "get_program_overview"
)

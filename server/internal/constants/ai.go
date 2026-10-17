package constants

// Ask AI (Google Gemini, Groq or OpenRouter, chosen with LLM_PROVIDER). Each provider has its own key and model variables.
const (
	// DefaultGeminiModel is the Gemini default; model IDs change (older ones close to new accounts), so confirm it in AI Studio.
	DefaultGeminiModel = "gemini-3.5-flash"
	// DefaultGroqModel is the Groq default; model IDs change, so confirm it in the Groq console.
	DefaultGroqModel = "openai/gpt-oss-120b"
	// GroqChatURL is Groq's OpenAI-compatible chat completions endpoint.
	GroqChatURL = "https://api.groq.com/openai/v1/chat/completions"
	// DefaultOpenRouterModel is the free router: it picks a free model that has the features the request needs, such as
	// tool calling. Model ids change, so confirm it on openrouter.ai/models.
	DefaultOpenRouterModel = "openrouter/free"
	// OpenRouterChatURL is OpenRouter's OpenAI-compatible chat completions endpoint.
	OpenRouterChatURL = "https://openrouter.ai/api/v1/chat/completions"
	// DefaultLLMTemperature is low so the same question gets a consistent answer.
	DefaultLLMTemperature = 0.2
	// Groq and OpenRouter both speak the OpenAI chat completions shape.
	OpenAIAuthHeader = "Authorization"
	OpenAIAuthScheme = "Bearer "
	// GroqReasoningModelPrefix marks the models that accept reasoning_effort; a low effort spends fewer tokens per answer.
	GroqReasoningModelPrefix = "openai/gpt-oss"
	GroqReasoningEffortLow   = "low"
	OpenAIRoleSystem         = "system"
	OpenAIRoleUser           = "user"
	OpenAIRoleAssistant      = "assistant"
	OpenAIRoleTool           = "tool"
	OpenAIToolTypeFunc       = "function"
	// LLMBaseURL is the Gemini REST endpoint; the model and :generateContent follow it.
	LLMBaseURL        = "https://generativelanguage.googleapis.com/v1beta/models/"
	LLMGenerateMethod = ":generateContent"
	LLMAPIKeyHeader   = "x-goog-api-key"

	LLMRoleUser  = "user"
	LLMRoleModel = "model"

	// Where a draft summary came from: the model, or the template written from the facts.
	SummarySourceAI       = "ai"
	SummarySourceTemplate = "template"

	// LLMToolCutMarker ends a tool result that was cut at LLM_MAX_TOOL_RESULT_BYTES.
	LLMToolCutMarker = "...[data cut: the result was too long, ask for a narrower range or one program]"

	// LLMMaxErrorBodyBytes is how much of an error reply is kept in the error.
	LLMMaxErrorBodyBytes = 512
)

// Tools the model may call. Each one reads through an existing service, so roles apply.
const (
	AIToolHeatmap          = "get_heatmap"
	AIToolRisks            = "get_risks"
	AIToolOpenDrafts       = "list_open_drafts"
	AIToolProject          = "get_project"
	AIToolProgramOverview  = "get_program_overview"
	AIToolCapacityForecast = "get_capacity_forecast"
	AIToolFindProjects     = "find_projects"
	AIToolFindPeople       = "find_people"
	AIToolCreateScenario   = "create_scenario"

	// AIFindMaxResults caps the matches a find tool returns.
	AIFindMaxResults = 5
	// AIScenarioMaxChanges caps the changes create_scenario returns, largest first.
	AIScenarioMaxChanges = 5
	// AIScenarioDemandSeparator joins the project and expertise names of a change.
	AIScenarioDemandSeparator = " - "
	// MaxAbsenceHoursPerDay is the most hours of capacity a day of leave can take.
	MaxAbsenceHoursPerDay = 24
	// AIScenarioNameMaxRunes is how much of the question goes into the scenario name.
	AIScenarioNameMaxRunes = 60
)

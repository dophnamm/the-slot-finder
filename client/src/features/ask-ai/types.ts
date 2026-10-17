export type AskAiMessageRole = 'user' | 'assistant' | 'error'

export interface AskAiMessage {
  role: AskAiMessageRole
  text: string
}

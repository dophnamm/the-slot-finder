export type AskAiMessageRole = 'user' | 'assistant'

export interface AskAiMessage {
  role: AskAiMessageRole
  text: string
}

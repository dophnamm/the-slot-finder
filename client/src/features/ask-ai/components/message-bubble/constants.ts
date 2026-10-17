import type { AskAiMessageRole } from '../../types'

export const BUBBLE_CLASS = 'max-w-[85%] rounded-lg px-3 py-2 text-sm has-[table]:w-full has-[table]:max-w-full'

/** An error looks like an answer in red, so a failed question reads as a reply to retry. */
export const BUBBLE_ROLE_CLASSES: Record<AskAiMessageRole, string> = {
  user: 'self-end bg-primary whitespace-pre-wrap text-primary-foreground',
  assistant: 'self-start bg-muted text-foreground',
  error: 'self-start border border-destructive/30 bg-destructive/10 whitespace-pre-wrap text-destructive',
}

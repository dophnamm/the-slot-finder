import type { ExternalAssistant, InternalAssistant } from './types'

// Static placeholder data until the feature is wired to GET /programs/:id/assistants.

export const INTERNAL_ASSISTANTS: InternalAssistant[] = [
  { person: 'Linh Tran', email: 'linh.tran@example.com' },
  { person: 'Minh Pham', email: 'minh.pham@example.com' },
]

export const EXTERNAL_ASSISTANTS: ExternalAssistant[] = [
  { name: 'Alex Morgan', email: 'alex@partner.example', description: 'Vendor liaison' },
]

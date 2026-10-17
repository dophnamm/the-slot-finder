export interface AskRequest {
  question: string
  page_context?: string
}

export interface AskAnswer {
  answer: string
}

export interface IntakeRequest {
  text: string
  mask_terms?: string[]
}

export interface IntakeDemand {
  expertise_id: string
  expertise_name: string
  total_hours: number
  min_level: number
  quote: string
}

export interface IntakeMasked {
  emails: number
  customers: number
  phones: number
}

/** What the AI read from a brief; a field it could not find is null. */
export interface IntakeSuggestion {
  name: string | null
  start_earliest: string | null
  deadline: string | null
  scope_points: number | null
  demands: IntakeDemand[]
  warnings: string[]
  masked: IntakeMasked
}

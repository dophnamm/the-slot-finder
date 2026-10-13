import type { ApiError } from '@/shared/api/types'

export interface CalendarFeedFormProps {
  personId: string
  /** Connects the trimmed link; the panel owns the request, so its result outlives this form. */
  onConnect: (link: string) => void
  pending: boolean
  error: ApiError | null
}

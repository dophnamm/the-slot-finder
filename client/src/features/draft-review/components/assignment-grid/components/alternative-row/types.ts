import type { Alternative } from '@/shared/staffing'

export interface AlternativeRowProps {
  alternative: Alternative
  busy: boolean
  onChoose: (personId: string) => void
}

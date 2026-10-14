import type { PlainKey } from '@/shared/i18n'

export interface LeadSelectProps {
  id: string
  /** The chosen person id; empty means no lead. */
  value: string
  /** The option for no one; "No lead" when left out. */
  emptyLabel?: PlainKey
  onChange: (personId: string) => void
}

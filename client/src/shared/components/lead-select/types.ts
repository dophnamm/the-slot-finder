export interface LeadSelectProps {
  id: string
  /** The chosen person id; empty means no lead. */
  value: string
  onChange: (personId: string) => void
}

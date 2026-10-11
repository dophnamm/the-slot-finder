import type { Phase } from '@/shared/programs'

export interface PhaseRowProps {
  phase: Phase
  editable: boolean
  disabled: boolean
  onEdit: (phase: Phase) => void
  onRemove: (phase: Phase) => void
}

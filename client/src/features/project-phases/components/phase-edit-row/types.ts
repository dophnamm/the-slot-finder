import type { PhaseDraft } from '../../types'

export interface PhaseEditRowProps {
  draft: PhaseDraft
  canSave: boolean
  busy: boolean
  onChange: (draft: PhaseDraft) => void
  onSave: () => void
  onCancel: () => void
}

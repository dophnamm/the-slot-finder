import type { ReactNode } from 'react'
import type { PhaseDraft } from '../../types'

export interface PhaseDialogProps {
  draft: PhaseDraft
  canSave: boolean
  busy: boolean
  /** The timeline and the slip and deadline notes, previewing the change before it is saved. */
  preview: ReactNode
  onChange: (draft: PhaseDraft) => void
  onSave: () => void
  onCancel: () => void
}

import type { ProgramAction } from '@/shared/programs'

export interface ProgramActionButtonProps {
  action: ProgramAction
  disabled: boolean
  onAction: (action: ProgramAction) => void
}

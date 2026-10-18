import type { Program } from '@/shared/programs'

export interface ProgramRowProps {
  program: Program
  /** Shows the arrow that opens the preview; it reads the overview and members, which a specialist may not. */
  canPreview: boolean
}

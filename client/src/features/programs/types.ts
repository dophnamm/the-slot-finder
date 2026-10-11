import type { Program } from '@/shared/programs'

export interface ProgramsTableProps {
  /** Called after a program is created, e.g. to open it. */
  onCreated: (program: Program) => void
}

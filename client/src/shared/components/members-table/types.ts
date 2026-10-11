import type { ProgramMember } from '@/shared/programs'

export interface MembersTableProps {
  members: ProgramMember[]
  isPending: boolean
  /** The program the members belong to, for their project links. */
  programId: string
}

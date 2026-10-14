import type { TeamMember } from '../../types'

export interface TeamTableProps {
  members: TeamMember[]
  isPending: boolean
  /** The program the team belongs to, for the members' project links. */
  programId: string
  /** Adds a column with Remove on each row. */
  canManage: boolean
  busy: boolean
  onRemove: (member: TeamMember) => void
}

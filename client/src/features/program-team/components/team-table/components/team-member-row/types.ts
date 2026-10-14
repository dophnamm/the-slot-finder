import type { TeamMember } from '../../../../types'

export interface TeamMemberRowProps {
  member: TeamMember
  programId: string
  canManage: boolean
  busy: boolean
  onRemove: (member: TeamMember) => void
}

import type { TeamMember } from '../../../../types'

export interface TeamMemberRowProps {
  member: TeamMember
  programId: string
  canManage: boolean
  removing: boolean
  onRemove: (member: TeamMember) => void
}

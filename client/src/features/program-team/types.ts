import type { Person } from '@/shared/people'
import type { MemberExpertise, ProjectRef } from '@/shared/programs'

export interface ProgramTeamProps {
  programId: string
  /** Planners add and remove members; a manager reads the team. */
  canManage: boolean
  /** Opens a person's page and shows their expertise and hours; a specialist sees names only. */
  canViewPeople: boolean
}

/** One person of a program's team (`GET /programs/:id/team`), with their hours in its baseline. */
export interface TeamMember {
  person: Person
  total_hours: number
  /** Null until a draft gives the member hours in this program. */
  from_week: string | null
  to_week: string | null
  expertises: MemberExpertise[]
  projects: ProjectRef[]
}

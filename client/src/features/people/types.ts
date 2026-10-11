import type { Role } from '@/resources/rbac'

export interface PersonSkill {
  expertise_id: string
  expertise: string
  level: number
  primary: boolean
}

/** A person as `GET /people` returns it. Dates use the server format `2026-10-12`. */
export interface Person {
  id: string
  name: string
  email: string
  role: Role
  weekly_hours: number
  non_project_hours: number
  start_date: string | null
  end_date: string | null
  active: boolean
  skills: PersonSkill[]
}

export type { Expertise } from '@/shared/expertises'

export interface SkillInput {
  expertise_id: string
  level: number
  primary: boolean
}

export interface CreatePersonInput {
  name: string
  email: string
  role: Role
  weekly_hours: number
  non_project_hours?: number
  start_date?: string
  end_date?: string
  skills: SkillInput[]
}

/** Only the fields that change; a null date clears it. */
export interface UpdatePersonInput {
  name?: string
  role?: Role
  weekly_hours?: number
  non_project_hours?: number
  start_date?: string | null
  end_date?: string | null
  active?: boolean
}

/** What the person form saves: a new person, or an existing one with its changes. */
export type SavePersonInput =
  | { mode: 'create'; person: CreatePersonInput }
  | { mode: 'update'; id: string; changes: UpdatePersonInput; skills: SkillInput[] | null }

export interface PeopleTableProps {
  /** Shows a person's absences; without it the row has no Absences button. */
  onShowAbsences?: (person: Person) => void
}

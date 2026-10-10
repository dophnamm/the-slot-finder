import type { Role } from '@/resources/rbac'
import type { CreatePersonInput, Person, SkillInput, UpdatePersonInput } from '../../types'
import { EMPTY_FORM } from './constants'
import type { PersonFormValues } from './types'

export function initialValues(person: Person | null): PersonFormValues {
  if (!person) return EMPTY_FORM
  return {
    name: person.name,
    email: person.email,
    role: person.role,
    weekly_hours: String(person.weekly_hours),
    non_project_hours: String(person.non_project_hours),
    start_date: person.start_date ?? '',
    end_date: person.end_date ?? '',
    active: person.active,
  }
}

export function initialSkills(person: Person | null): SkillInput[] {
  return (person?.skills ?? []).map(({ expertise_id, level, primary }) => ({ expertise_id, level, primary }))
}

export function toCreateInput(values: PersonFormValues, skills: SkillInput[]): CreatePersonInput {
  return {
    name: values.name.trim(),
    email: values.email.trim(),
    role: values.role,
    weekly_hours: Number(values.weekly_hours),
    non_project_hours: values.non_project_hours === '' ? undefined : Number(values.non_project_hours),
    start_date: values.start_date || undefined,
    end_date: values.end_date || undefined,
    skills,
  }
}

/** Only the fields that differ from the saved person; an emptied date becomes null. */
export function toChanges(person: Person, values: PersonFormValues): UpdatePersonInput {
  const changes: UpdatePersonInput = {}
  const name = values.name.trim()
  if (name !== person.name) changes.name = name
  if (values.role !== person.role) changes.role = values.role
  if (Number(values.weekly_hours) !== person.weekly_hours) changes.weekly_hours = Number(values.weekly_hours)
  if (values.non_project_hours !== '' && Number(values.non_project_hours) !== person.non_project_hours) {
    changes.non_project_hours = Number(values.non_project_hours)
  }
  if (values.start_date !== (person.start_date ?? '')) changes.start_date = values.start_date || null
  if (values.end_date !== (person.end_date ?? '')) changes.end_date = values.end_date || null
  if (values.active !== person.active) changes.active = values.active
  return changes
}

export function skillsChanged(before: SkillInput[], after: SkillInput[]): boolean {
  return JSON.stringify(before) !== JSON.stringify(after)
}

/** The form can be saved with a name, an email, hours, and exactly one primary skill. */
export function canSave(values: PersonFormValues, skills: SkillInput[]): boolean {
  const primaries = skills.filter((skill) => skill.primary).length
  return values.name.trim() !== '' && values.email.trim() !== '' && Number(values.weekly_hours) > 0 && primaries === 1
}

export function isRole(value: string, roles: readonly Role[]): value is Role {
  return (roles as readonly string[]).includes(value)
}

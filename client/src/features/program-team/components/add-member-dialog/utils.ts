import type { Person } from '@/shared/people'
import type { CandidateOption } from './types'

/** Active people outside the team, by name, each with their primary expertise. */
export function candidateOptions(people: Person[], teamIds: string[]): CandidateOption[] {
  return people
    .filter((person) => person.active && !teamIds.includes(person.id))
    .map((person) => ({
      id: person.id,
      name: person.name,
      expertise: person.skills.find((skill) => skill.primary)?.expertise ?? null,
    }))
    .sort((a, b) => a.name.localeCompare(b.name))
}

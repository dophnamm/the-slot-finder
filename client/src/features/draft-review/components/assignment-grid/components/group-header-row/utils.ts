import type { DraftDemand } from '@/shared/staffing'
import type { StaffablePerson } from '../../../../types'
import type { PersonOption } from './types'

/** People of the demand's program team not on the demand yet: those who fit first, then stretch, each by name. */
export function personOptions(people: StaffablePerson[], demand: DraftDemand, taken: string[]): PersonOption[] {
  const team = demand.team_ids
  return people
    .filter((person) => !taken.includes(person.id) && (!team || team.includes(person.id)))
    .map((person) => ({
      id: person.id,
      name: person.name,
      stretch: !person.skills.some((skill) => skill.expertise === demand.expertise && skill.level >= demand.min_level),
    }))
    .sort((a, b) => Number(a.stretch) - Number(b.stretch) || a.name.localeCompare(b.name))
}

import type { StaffablePerson } from '../../../../types'
import type { DemandGroup } from '../../types'

export interface GroupHeaderRowProps {
  group: DemandGroup
  /** Columns the header spans: the person column and every week. */
  weeks: string[]
  people: StaffablePerson[]
  /** The people to add are still loading. */
  peopleLoading: boolean
  editable: boolean
  busy: boolean
  onAdd: (demandId: string, personId: string) => void
}

/** A person in the add list; stretch when they lack the skill or the level. */
export interface PersonOption {
  id: string
  name: string
  stretch: boolean
}

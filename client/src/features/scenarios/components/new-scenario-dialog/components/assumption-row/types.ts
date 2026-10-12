import type { Project } from '@/shared/programs'
import type { PersonOption } from '../../../../types'
import type { AssumptionDraft } from '../../types'

export interface AssumptionRowProps {
  draft: AssumptionDraft
  projects: Project[]
  people: PersonOption[]
  /** False for the last row: a scenario needs at least one assumption. */
  removable: boolean
  onChange: (draft: AssumptionDraft) => void
  onRemove: (key: string) => void
}

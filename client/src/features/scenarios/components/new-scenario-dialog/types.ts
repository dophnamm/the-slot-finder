import type { AssumptionType } from '@/shared/scenarios'

export interface NewScenarioDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
}

/** One assumption while it is being typed; numbers stay text until sent. */
export interface AssumptionDraft {
  key: string
  type: AssumptionType
  project_id: string
  weeks: string
  person_id: string
  from_date: string
  to_date: string
  hours_per_day: string
}

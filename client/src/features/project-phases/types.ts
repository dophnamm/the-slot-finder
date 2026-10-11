export interface ProjectPhasesProps {
  projectId: string
}

/** A phase as the editor holds it; `id` is null for a phase being added. */
export interface PhaseDraft {
  id: string | null
  name: string
  start_week: string
  end_week: string
}

export type PhaseDraftField = 'name' | 'start_week' | 'end_week'

/** How far an unsaved end-week change moves the later phases. */
export interface Slip {
  weeks: number
  phases: number
}

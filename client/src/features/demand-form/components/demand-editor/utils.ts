import type { Expertise } from '@/shared/expertises'
import type { Demand, DemandInput, ProjectDetail } from '@/shared/programs'
import { DEMAND_LIMITS, NEW_DEMAND } from './constants'
import type { DemandDraft, WeekWindow } from './types'

export function toDrafts(demands: Demand[]): DemandDraft[] {
  return demands.map((d) => ({
    key: d.id,
    id: d.id,
    expertise_id: d.expertise_id,
    phase_id: d.phase_id ?? '',
    min_level: d.min_level,
    headcount: String(d.headcount),
    total_hours: String(d.total_hours),
  }))
}

export function newDraft(expertises: Expertise[]): DemandDraft {
  return {
    key: crypto.randomUUID(),
    expertise_id: expertises[0]?.id ?? '',
    phase_id: '',
    min_level: NEW_DEMAND.minLevel,
    headcount: NEW_DEMAND.headcount,
    total_hours: NEW_DEMAND.totalHours,
  }
}

export function toDemandInput(draft: DemandDraft): DemandInput {
  return {
    id: draft.id,
    expertise_id: draft.expertise_id,
    phase_id: draft.phase_id || null,
    min_level: draft.min_level,
    headcount: Number(draft.headcount),
    total_hours: Number(draft.total_hours),
  }
}

export function isValidDraft(draft: DemandDraft): boolean {
  const headcount = Number(draft.headcount)
  const hours = Number(draft.total_hours)
  return (
    draft.expertise_id !== '' &&
    Number.isInteger(headcount) &&
    headcount >= DEMAND_LIMITS.minHeadcount &&
    headcount <= DEMAND_LIMITS.maxHeadcount &&
    hours > 0 &&
    hours <= DEMAND_LIMITS.maxHours &&
    isHoursStep(draft.total_hours)
  )
}

/** Whether the hours are a whole multiple of the planning step (an empty or invalid number passes). */
export function isHoursStep(totalHours: string): boolean {
  const hours = Number(totalHours)
  return !Number.isFinite(hours) || hours % DEMAND_LIMITS.hoursStep === 0
}

/** Weeks a demand may be staffed in: its phase, or the project's start through its deadline. */
export function demandWindow(project: ProjectDetail, phaseId: string | null): WeekWindow {
  const phase = project.phases.find((p) => p.id === phaseId)
  return phase
    ? { from: phase.start_week, to: phase.end_week }
    : { from: project.start_week, to: project.deadline_week }
}

export function phaseName(project: ProjectDetail, phaseId: string | null): string | null {
  return project.phases.find((p) => p.id === phaseId)?.name ?? null
}

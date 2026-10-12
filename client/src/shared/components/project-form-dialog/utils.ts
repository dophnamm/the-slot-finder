import { DEFAULT_PRIORITY, type Project, type ProjectInput } from '@/shared/programs'
import type { ProjectFormValues } from './types'

export function initialValues(project: Project | null): ProjectFormValues {
  return {
    name: project?.name ?? '',
    priority: String(project?.priority ?? DEFAULT_PRIORITY),
    start_earliest: project?.start_earliest ?? '',
    deadline: project?.deadline ?? '',
    scope_points: String(project?.scope_points ?? 0),
    done_points: String(project?.done_points ?? 0),
    lead_id: project?.lead?.id ?? '',
  }
}

export function toInput(values: ProjectFormValues, canChooseLead: boolean): ProjectInput {
  return {
    name: values.name.trim(),
    priority: Number(values.priority),
    start_earliest: values.start_earliest,
    deadline: values.deadline,
    scope_points: Number(values.scope_points),
    done_points: Number(values.done_points),
    lead_id: canChooseLead ? values.lead_id || null : undefined,
  }
}

/** Only the fields that differ from the saved project. */
export function toChanges(project: Project, values: ProjectFormValues, canChooseLead: boolean): Partial<ProjectInput> {
  const next = toInput(values, canChooseLead)
  const changes: Partial<ProjectInput> = {}
  if (next.name !== project.name) changes.name = next.name
  if (next.priority !== project.priority) changes.priority = next.priority
  if (next.start_earliest !== project.start_earliest) changes.start_earliest = next.start_earliest
  if (next.deadline !== project.deadline) changes.deadline = next.deadline
  if (next.scope_points !== project.scope_points) changes.scope_points = next.scope_points
  if (next.done_points !== project.done_points) changes.done_points = next.done_points
  if (canChooseLead && (next.lead_id ?? null) !== (project.lead?.id ?? null)) changes.lead_id = next.lead_id
  return changes
}

/** A name and both dates, the deadline not before the start. */
export function canSave(values: ProjectFormValues): boolean {
  return (
    values.name.trim() !== '' &&
    values.start_earliest !== '' &&
    values.deadline !== '' &&
    values.deadline >= values.start_earliest
  )
}

import type { Program, ProgramInput } from '@/shared/programs'
import type { ProgramFormValues } from './types'

const LINE_BREAK = /\r?\n/

export function initialValues(program: Program | null): ProgramFormValues {
  return {
    name: program?.name ?? '',
    client: program?.client ?? '',
    start_date: program?.start_date ?? '',
    end_date: program?.end_date ?? '',
    goal: program?.goal ?? '',
    objectives: (program?.objectives ?? []).join('\n'),
    charter_url: program?.charter_url ?? '',
    lead_id: program?.lead?.id ?? '',
  }
}

function objectiveList(text: string): string[] {
  return text
    .split(LINE_BREAK)
    .map((line) => line.trim())
    .filter(Boolean)
}

/** The full new program; the lead is sent only when the user may choose it. */
export function toInput(values: ProgramFormValues, canChooseLead: boolean): ProgramInput {
  return {
    name: values.name.trim(),
    client: values.client.trim(),
    start_date: values.start_date,
    end_date: values.end_date,
    goal: values.goal.trim(),
    objectives: objectiveList(values.objectives),
    charter_url: values.charter_url.trim() || null,
    lead_id: canChooseLead ? values.lead_id || null : undefined,
  }
}

/** Only the fields that differ from the saved program. */
export function toChanges(program: Program, values: ProgramFormValues, canChooseLead: boolean): Partial<ProgramInput> {
  const next = toInput(values, canChooseLead)
  const changes: Partial<ProgramInput> = {}
  if (next.name !== program.name) changes.name = next.name
  if (next.client !== program.client) changes.client = next.client
  if (next.start_date !== program.start_date) changes.start_date = next.start_date
  if (next.end_date !== program.end_date) changes.end_date = next.end_date
  if (next.goal !== program.goal) changes.goal = next.goal
  if (JSON.stringify(next.objectives) !== JSON.stringify(program.objectives)) changes.objectives = next.objectives
  if (next.charter_url !== program.charter_url) changes.charter_url = next.charter_url
  if (canChooseLead && (next.lead_id ?? null) !== (program.lead?.id ?? null)) changes.lead_id = next.lead_id
  return changes
}

/** A name, a client and both dates, the end not before the start. */
export function canSave(values: ProgramFormValues): boolean {
  return (
    values.name.trim() !== '' &&
    values.client.trim() !== '' &&
    values.start_date !== '' &&
    values.end_date !== '' &&
    values.end_date >= values.start_date
  )
}

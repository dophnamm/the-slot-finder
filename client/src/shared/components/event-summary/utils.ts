import { CAPACITY_FIELD_LABELS, STATUS_KIND_PROGRAM } from '@/shared/events'
import type { PlainKey } from '@/shared/i18n'
import { PROGRAM_STATUS_LABELS, PROJECT_STATUS_LABELS } from '@/shared/programs'

/** The translation keys of the known capacity fields; an unknown field is skipped. */
export function capacityFieldKeys(fields: string[]): PlainKey[] {
  return fields.flatMap((field) => CAPACITY_FIELD_LABELS[field] ?? [])
}

/** The labels of the known capacity fields, translated; an unknown field is skipped. */
export function capacityFieldLabels(fields: string[], t: (key: PlainKey) => string): string {
  return capacityFieldKeys(fields)
    .map((key) => t(key))
    .join(', ')
}

/** The label key of a project or program status, or undefined for an unknown one. */
export function statusKey(kind: string | undefined, status: string | undefined): PlainKey | undefined {
  const labels: Record<string, PlainKey> = kind === STATUS_KIND_PROGRAM ? PROGRAM_STATUS_LABELS : PROJECT_STATUS_LABELS
  return status === undefined ? undefined : labels[status]
}

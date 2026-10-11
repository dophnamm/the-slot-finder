import { ROUTES } from '@/resources/routes'
import type { BreadcrumbKind, StaticCrumb } from './types'

/** Trails made only of fixed texts; program and project pages build theirs from server data. */
export const STATIC_TRAILS: Partial<Record<BreadcrumbKind, StaticCrumb[]>> = {
  dashboard: [{ label: 'Dashboard' }],
  programs: [{ label: 'Programs' }],
  intake: [{ label: 'Intake' }],
  staffingDraft: [{ label: 'Intake', to: ROUTES.intake }, { label: 'Staffing draft' }],
  scenarios: [{ label: 'Scenarios' }],
  people: [{ label: 'People' }],
  events: [{ label: 'Events' }],
  mySchedule: [{ label: 'My schedule' }],
}

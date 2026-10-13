import type { PlainKey } from '@/shared/i18n'

/** Which trail a route shows; set as the route's `handle` in app/router.tsx. */
export type BreadcrumbKind =
  | 'dashboard'
  | 'programs'
  | 'programDetail'
  | 'projectDetail'
  | 'intake'
  | 'staffingDraft'
  | 'scenarios'
  | 'scenarioDetail'
  | 'people'
  | 'personDetail'
  | 'events'
  | 'mySchedule'

export interface BreadcrumbHandle {
  breadcrumb: BreadcrumbKind
}

/** One step of a trail whose labels are fixed texts. */
export interface StaticCrumb {
  label: PlainKey
  to?: string
}

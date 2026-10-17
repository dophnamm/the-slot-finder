import type { PlainKey } from '@/shared/i18n'

export const TOUR_STORAGE_PREFIX = 'slot-finder:tour-seen:'
export const TOUR_SEEN_VALUE = '1'
export const TOUR_POPOVER_CLASS = 'slot-finder-tour'

/** What each sidebar screen is for, by its link label. */
export const NAV_DESCRIPTIONS: Record<string, PlainKey> = {
  Dashboard: 'See who is overloaded or idle each week, and which deadlines are at risk.',
  Programs: 'Open a program to see its team, its projects and how well they are staffed.',
  Intake: 'Incoming projects wait here. Propose a squad to turn one into a staffing draft.',
  Drafts: 'Review staffing drafts, edit the weekly hours, then accept one as the new baseline.',
  Scenarios: 'Try a what-if, such as a project moved or someone on leave, without touching the baseline.',
  People: 'Find people, their expertises, levels, capacity and absences.',
  Events: 'A history of changes, such as leave or slipped deadlines, that triggered a re-plan.',
  'My schedule': 'Your own month: hours per project each week, leave and holidays. Report an absence here.',
}

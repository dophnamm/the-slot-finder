import type { ProgramTimeline, ProgressItem, StaffingRow } from './types'

// Static placeholder data until the feature is wired to GET /programs/:id/overview.

export const PROGRESS: ProgressItem[] = [
  { label: 'Time', value: 62 },
  { label: 'Scope', value: 45 },
  { label: 'Cost', value: 41 },
  { label: 'Staffing', value: 88 },
]

export const MEMBER_INITIALS = ['AN', 'BT', 'CL']

export const TIMELINE: ProgramTimeline = {
  start: '2026-W40',
  end: '2026-W51',
  dates: 'Sep 28 – Dec 18',
  milestone: 'Beta release',
  milestoneWeek: '2026-W46',
}

export const GOAL = 'one place to plan who works on what, week by week.'

export const OBJECTIVES = [
  'Ship the self-service booking flow.',
  'Retire the legacy scheduling tool.',
  'Cut manual planning time by half.',
]

export const STAFFING: StaffingRow[] = [
  { expertise: 'PM', budget: 320, planned: 300, forecast: 310, gap: 10, staffing: 94, contribution: 12 },
  { expertise: 'Analyst', budget: 480, planned: 420, forecast: 460, gap: 20, staffing: 88, contribution: 18 },
  { expertise: 'Engineer', budget: 1600, planned: 1450, forecast: 1580, gap: 20, staffing: 91, contribution: 58 },
  { expertise: 'QA', budget: 360, planned: 260, forecast: 300, gap: 60, staffing: 72, contribution: 12 },
]

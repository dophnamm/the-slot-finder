import type { PlainKey } from '@/shared/i18n'

export interface ProgramOverviewProps {
  programId: string
}

export interface ProgressItem {
  label: PlainKey
  value: number
}

export interface ProgramTimeline {
  start: string
  end: string
  dates: string
  milestone: string
  milestoneWeek: string
}

export interface StaffingRow {
  expertise: string
  budget: number
  planned: number
  forecast: number
  gap: number
  staffing: number
  contribution: number
}

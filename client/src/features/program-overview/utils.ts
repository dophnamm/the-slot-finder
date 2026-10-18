import { PROGRAM_TABS } from '@/resources/routes'
import { DEFAULT_PROGRAM_TAB, NO_OVERVIEW_PROGRAM_TAB } from './constants'
import type { ProgramTab } from './types'

/**
 * The tab a link asks for, or the first one the viewer has when it names none or an unknown one. Without the
 * overview (a specialist) a link to it falls back to the projects too.
 */
export function programTab(value: string | null, hasOverview: boolean): ProgramTab {
  const known = (Object.values(PROGRAM_TABS) as string[]).includes(value ?? '')
  if (known && (hasOverview || value !== PROGRAM_TABS.overview)) return value as ProgramTab
  return hasOverview ? DEFAULT_PROGRAM_TAB : NO_OVERVIEW_PROGRAM_TAB
}

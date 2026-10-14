import { PROGRAM_TABS } from '@/resources/routes'
import { DEFAULT_PROGRAM_TAB } from './constants'
import type { ProgramTab } from './types'

/** The tab a link asks for, or the overview when it names none or an unknown one. */
export function programTab(value: string | null): ProgramTab {
  return (Object.values(PROGRAM_TABS) as string[]).includes(value ?? '') ? (value as ProgramTab) : DEFAULT_PROGRAM_TAB
}

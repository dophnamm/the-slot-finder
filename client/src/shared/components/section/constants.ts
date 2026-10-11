import type { SectionVariant } from './types'

/** The body's layout, with or without a card. */
export const SECTION_BODY_CLASS = 'flex min-h-0 flex-1 flex-col gap-4'

export const SECTION_TITLE_CLASS: Record<SectionVariant, string> = {
  section: 'flex flex-wrap items-center gap-2 font-heading text-base font-semibold',
  page: 'flex flex-wrap items-center gap-2 font-heading text-xl font-semibold',
}

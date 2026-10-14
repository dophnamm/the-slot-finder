import type { SectionVariant } from './types'

/** A white card with a hairline border on the gray page; no shadow. */
export const SECTION_CARD_CLASS = 'overflow-hidden rounded-xl border bg-background'

/** The body of a card that holds text, numbers or a form. */
export const SECTION_PADDED_BODY_CLASS = 'flex min-h-0 flex-1 flex-col gap-4 px-5 pb-5'

/**
 * The body of a card that holds a table: the table runs from edge to edge, everything else keeps the
 * card's inset, and a table at the bottom closes the card without a gap under it.
 */
export const SECTION_FLUSH_BODY_CLASS =
  'flex min-h-0 flex-1 flex-col gap-4 pb-5 *:not-data-[slot=table-frame]:not-data-[slot=loading-bar]:mx-5 has-[>[data-slot=table-frame]:last-child]:pb-0'

/** A page header's body sits in a card under the heading. */
export const PAGE_BODY_CLASS = 'flex min-h-0 flex-1 flex-col gap-4 p-5'

export const SECTION_DESCRIPTION_CLASS = 'max-w-[75ch] text-sm text-pretty text-muted-foreground'

export const SECTION_TITLE_CLASS: Record<SectionVariant, string> = {
  section: 'flex flex-wrap items-center gap-2 font-heading text-base font-semibold',
  page: 'flex flex-wrap items-center gap-2 font-heading text-lg font-semibold',
}

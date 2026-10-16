/** Server limits of POST /ai/ask. */
export const MAX_QUESTION_LENGTH = 500
export const MAX_PAGE_CONTEXT_LENGTH = 200

/** Same dimmed, blurred layer as the dialog overlay, so the page behind the open panel is covered. */
export const BACKDROP_CLASS =
  'fixed inset-0 z-40 bg-black/10 transition-opacity duration-200 ease-out supports-backdrop-filter:backdrop-blur-xs'
export const BACKDROP_OPEN_CLASS = 'opacity-100'
export const BACKDROP_CLOSED_CLASS = 'pointer-events-none opacity-0'

export const PANEL_CLASS =
  'flex max-w-[calc(100vw-2rem)] origin-bottom-right flex-col overflow-hidden rounded-xl border bg-popover text-popover-foreground shadow-lg transition-[opacity,transform,width,height,visibility] duration-200 ease-out'
export const PANEL_OPEN_CLASS = 'pointer-events-auto visible translate-y-0 scale-100 opacity-100'
export const PANEL_CLOSED_CLASS = 'pointer-events-none invisible translate-y-2 scale-95 opacity-0'

export const PANEL_SIZE_CLASS = {
  normal: 'h-[min(42rem,calc(100svh-8rem))] w-96',
  large: 'h-[min(56rem,calc(100svh-8rem))] w-[min(64rem,calc(100vw-2rem))]',
} as const

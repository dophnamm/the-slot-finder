/** Loading placeholders fade in after 300 ms, so an answer that comes quickly never flashes them. */
export const LOADING_REVEAL_CLASS = 'animate-in fade-in fill-mode-both delay-300 duration-200'

/** Bar lengths of skeleton text: uneven, so the rows read as text rather than stripes. */
export const SKELETON_TEXT_WIDTHS = ['w-3/5', 'w-2/5', 'w-1/2', 'w-1/3', 'w-2/3'] as const

/** Rows a table shows while its first page loads. */
export const LOADING_TABLE_ROWS = 5

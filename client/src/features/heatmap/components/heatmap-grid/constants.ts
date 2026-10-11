/** Week headers stay at the top while the rows scroll; opaque, so rows do not show through. */
export const HEADER_CELL_CLASS = 'sticky top-0 z-20 bg-muted'

/** The person column stays in view while the weeks scroll sideways. */
export const PERSON_COLUMN_CLASS = 'sticky left-0 z-10 min-w-48 bg-card'

/** The top-left header cell stays above both. */
export const CORNER_CELL_CLASS = 'sticky top-0 left-0 z-30 min-w-48 bg-muted'

/** An expertise divider row; its label stays in the person column. */
export const GROUP_ROW_CLASS = 'bg-muted hover:bg-muted'
export const GROUP_LABEL_CLASS = 'sticky left-0 z-10 bg-muted text-xs font-medium text-muted-foreground'

/** Every week column: room for the cell, which sits centred under its header. */
export const WEEK_COLUMN_CLASS = 'min-w-24 text-center'

/** This week's column: its body cells are outlined on both sides, so the week is easy to find. */
export const CURRENT_WEEK_CLASS = 'shadow-[inset_2px_0_0_var(--color-underused),inset_-2px_0_0_var(--color-underused)]'

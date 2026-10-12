/** The person column stays in view while the weeks scroll sideways. */
export const PERSON_COLUMN_CLASS = 'sticky left-0 z-10 w-80 min-w-80 bg-card'

/** Week columns share one width and a line between them, so hours sit right under their week. */
export const WEEK_COLUMN_CLASS = 'w-20 min-w-20 border-l text-center'

/** A person's row; the sticky person cell takes the same hover colour as the rest of the row. */
export const ASSIGNMENT_ROW_CLASS = 'group/row hover:bg-muted'
export const ASSIGNMENT_PERSON_CELL_CLASS = 'group-hover/row:bg-muted'

/** A demand's row above its people: a solid band, so it reads as a heading. */
export const GROUP_ROW_CLASS = 'bg-muted hover:bg-muted'
export const GROUP_LABEL_CELL_CLASS = 'bg-muted align-top whitespace-normal'

/** The demand's weeks, drawn as a bar under the week columns. */
export const WINDOW_BAR_CLASS = 'h-1.5 rounded-full bg-underused/40'

/** Columns of the table; the loading and empty rows span them all. */
export const LABEL_COLUMNS = 1

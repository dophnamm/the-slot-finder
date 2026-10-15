/** The week label column and the hours column on either side of the weekday columns. */
export const CALENDAR_LABEL_COLUMNS = 2

/** A faint line on the left of each weekday column, so the days line up down the month. */
export const DAY_COLUMN_CLASS = 'border-l'

/** Every ISO weekday, Monday (1) to Sunday (7): the calendar shows the whole week. */
export const CALENDAR_WEEKDAYS = [1, 2, 3, 4, 5, 6, 7] as const

/** Days off (outside the working weekdays) are shaded, so the working week still stands out. */
export const DAY_OFF_CLASS = 'bg-muted/50 text-muted-foreground'

/** The workload column is a little wider than its content (w-px shrinks it, min-w-88 sets the floor); the weekdays share the rest. */
export const WORKLOAD_COLUMN_CLASS = 'w-px min-w-88 whitespace-nowrap'

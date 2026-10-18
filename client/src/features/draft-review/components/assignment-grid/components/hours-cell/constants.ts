/** An empty week reads as zero hours. */
export const HOURS_CELL_PLACEHOLDER = '0'

/** An editable cell reads as a field: white box, shadow, darker border on hover, faint zero as the hint. */
export const HOURS_INPUT_CLASS =
  'mx-auto h-8 w-14 border-foreground/25 bg-background text-center shadow-xs placeholder:text-muted-foreground/60 hover:border-ring dark:bg-input/30'

/** The data a dragged cell carries; browsers need some to start a drag. */
export const DRAG_DATA_TYPE = 'text/plain'

/** A cell that may take the dragged hours: a ring in the theme colour. */
export const DROP_TARGET_CLASS = 'ring-2 ring-primary ring-inset'

/** A cell whose hours can be picked up. */
export const DRAGGABLE_CLASS = 'cursor-grab'

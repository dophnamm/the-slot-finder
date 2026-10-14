export interface TableLoadingProps {
  /** Columns of the table; each skeleton row has one cell per column. */
  columns: number
  /** How many skeleton rows to draw; a page of the table by default. */
  rows?: number
}

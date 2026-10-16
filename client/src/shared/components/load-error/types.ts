export interface LoadErrorProps {
  /** Why the request failed; its message says so under the title. */
  message: string
  onRetry: () => void
  /** Draws a bordered frame, for a message that stands in for a section. */
  framed?: boolean
  /** Renders as a table's only row spanning this many columns. */
  columns?: number
}

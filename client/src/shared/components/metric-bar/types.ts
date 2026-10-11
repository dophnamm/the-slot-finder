export interface MetricBarProps {
  /** Already translated. */
  label: string
  /** 0–100, or null when unknown; values over 100 fill the bar. */
  value: number | null
  /** The server raised a warning on this metric. */
  warning?: boolean
  /** Smaller text for table rows. */
  compact?: boolean
}

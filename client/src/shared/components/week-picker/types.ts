export interface WeekPickerProps {
  /** Ties the trigger to its `<Label htmlFor>`. */
  id?: string
  /** The form field the value belongs to; passed back to `onChange`. */
  name: string
  /** An ISO week such as `2026-W42`, or '' when empty. */
  value: string
  onChange: (value: string, name: string) => void
  /** Earliest ISO week that can be picked. */
  min?: string
  /** Latest ISO week that can be picked. */
  max?: string
  className?: string
  'aria-label'?: string
}

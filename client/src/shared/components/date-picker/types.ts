export interface DatePickerProps {
  /** Ties the trigger to its `<Label htmlFor>`. */
  id?: string
  /** The form field the value belongs to; passed back to `onChange`. */
  name: string
  /** `YYYY-MM-DD`, or '' when empty. */
  value: string
  onChange: (value: string, name: string) => void
  /** Earliest day that can be picked (`YYYY-MM-DD`). */
  min?: string
  /** Latest day that can be picked (`YYYY-MM-DD`). */
  max?: string
  disabled?: boolean
  'aria-label'?: string
}

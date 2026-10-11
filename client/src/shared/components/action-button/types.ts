export interface ActionButtonProps<A extends string> {
  action: A
  /** Already translated. */
  label: string
  destructive?: boolean
  /** When set, the user confirms this (translated) question first. */
  confirmText?: string
  disabled?: boolean
  onAction: (action: A) => void
}

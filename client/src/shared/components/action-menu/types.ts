export interface ActionMenuItem<A extends string> {
  action: A
  /** Already translated. */
  label: string
  /** Shown last, after a line, in red. */
  destructive?: boolean
  /** When set, the user confirms this (translated) question first. */
  confirmText?: string
}

export interface ActionMenuProps<A extends string> {
  items: readonly ActionMenuItem<A>[]
  disabled?: boolean
  onAction: (action: A) => void
}

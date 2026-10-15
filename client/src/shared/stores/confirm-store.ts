import { create } from 'zustand'

export interface ConfirmOptions {
  /** Already translated: the question. */
  title: string
  /** Already translated: what happens next. */
  description?: string
  /** Already translated: a risk the user should weigh, shown in red. */
  warning?: string
  /** Already translated: the verb on the confirming button, e.g. "Accept". */
  confirmLabel: string
  /** The action cannot be undone: the confirming button is red. */
  destructive?: boolean
}

interface ConfirmRequest extends ConfirmOptions {
  resolve: (confirmed: boolean) => void
}

interface ConfirmState {
  request: ConfirmRequest | null
  ask: (options: ConfirmOptions) => Promise<boolean>
  settle: (confirmed: boolean) => void
}

/** The question on screen, if any. `ConfirmDialog` shows it; callers use `confirm()`. */
export const useConfirmStore = create<ConfirmState>()((set, get) => ({
  request: null,
  ask: (options) =>
    new Promise<boolean>((resolve) => {
      get().request?.resolve(false)
      set({ request: { ...options, resolve } })
    }),
  settle: (confirmed) => {
    get().request?.resolve(confirmed)
    set({ request: null })
  },
}))

/** Asks the user in the app's own dialog; resolves true when they confirm, false when they cancel or close it. */
export function confirm(options: ConfirmOptions): Promise<boolean> {
  return useConfirmStore.getState().ask(options)
}

import type { PlainKey } from '@/shared/i18n'

export interface GuidedTourProps {
  /** Labels of the sidebar links the role can see, in sidebar order: one step each. */
  navLabels: PlainKey[]
  /** Whether the role can use Ask AI: adds a step for its button. */
  hasAskAi: boolean
}

/** Whether the browser's storage says the user has seen the tour. */
export type TourFlag = 'seen' | 'unseen' | 'unavailable'

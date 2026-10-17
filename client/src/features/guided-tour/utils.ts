import { driver, type DriveStep } from 'driver.js'
import type { PlainKey } from '@/shared/i18n'
import { ACCOUNT_TOUR_ANCHOR, ASK_AI_TOUR_ANCHOR } from '@/utils/constants'
import { navTourAnchor } from '@/utils/functions'
import { NAV_DESCRIPTIONS, TOUR_POPOVER_CLASS, TOUR_SEEN_VALUE, TOUR_STORAGE_PREFIX } from './constants'
import type { TourFlag } from './types'

type Translate = (key: PlainKey) => string

export function tourSelector(anchor: string) {
  return `[data-tour="${anchor}"]`
}

/** One step per sidebar link the role sees, the Ask AI button for the roles that can ask, then the account menu. */
export function buildTourSteps(navLabels: PlainKey[], hasAskAi: boolean, t: Translate): DriveStep[] {
  const navSteps = navLabels.map((label) => ({
    element: tourSelector(navTourAnchor(label)),
    popover: { title: t(label), description: t(NAV_DESCRIPTIONS[label] ?? label) },
  }))
  const askAiSteps: DriveStep[] = hasAskAi
    ? [
        {
          element: tourSelector(ASK_AI_TOUR_ANCHOR),
          popover: {
            title: t('Ask AI'),
            description: t('Ask a question about your data, such as who is overloaded or which project is at risk.'),
            side: 'left',
            align: 'end',
          },
        },
      ]
    : []
  return [
    ...navSteps,
    ...askAiSteps,
    {
      element: tourSelector(ACCOUNT_TOUR_ANCHOR),
      popover: { title: t('Account'), description: t('Take the tour again from here.') },
    },
  ]
}

/** Auto-start only when storage works, the tour is unseen and the first anchor is on screen. */
export function shouldAutoStart(flag: TourFlag, anchorVisible: boolean) {
  return flag === 'unseen' && anchorVisible
}

export function isVisible(element: Element | null) {
  return !!element && element.getClientRects().length > 0
}

export function readTourFlag(userId: string): TourFlag {
  try {
    return localStorage.getItem(TOUR_STORAGE_PREFIX + userId) === TOUR_SEEN_VALUE ? 'seen' : 'unseen'
  } catch {
    return 'unavailable'
  }
}

/** False when storage refuses the write. */
export function writeTourFlag(userId: string) {
  try {
    localStorage.setItem(TOUR_STORAGE_PREFIX + userId, TOUR_SEEN_VALUE)
    return true
  } catch {
    return false
  }
}

export function runTour(steps: DriveStep[], t: Translate) {
  const tour = driver({
    steps,
    showProgress: true,
    animate: !window.matchMedia('(prefers-reduced-motion: reduce)').matches,
    popoverClass: TOUR_POPOVER_CLASS,
    nextBtnText: t('Next'),
    prevBtnText: t('Back'),
    doneBtnText: t('Done'),
  })
  tour.drive()
  return tour
}

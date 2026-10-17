import 'driver.js/dist/driver.css'
import type { Driver } from 'driver.js'
import { useEffect, useRef } from 'react'
import { useCurrentUser } from '@/shared/auth'
import { useTranslation } from '@/shared/i18n'
import { useTourStore } from '@/shared/stores/tour-store'
import { navTourAnchor } from '@/utils/functions'
import type { GuidedTourProps } from './types'
import { buildTourSteps, isVisible, readTourFlag, runTour, shouldAutoStart, tourSelector, writeTourFlag } from './utils'

/** Renders nothing: starts the sidebar tour on a user's first visit and whenever the account menu asks. */
export function GuidedTour({ navLabels, hasAskAi }: GuidedTourProps) {
  const { t } = useTranslation()
  const { data: user } = useCurrentUser()
  const startRequests = useTourStore((state) => state.startRequests)
  const tourRef = useRef<Driver | null>(null)
  const labelsKey = navLabels.join('|')

  useEffect(() => {
    if (!user || navLabels.length === 0) return
    const anchor = document.querySelector(tourSelector(navTourAnchor(navLabels[0])))
    if (!shouldAutoStart(readTourFlag(user.id), isVisible(anchor))) return
    if (!writeTourFlag(user.id)) return
    tourRef.current = runTour(buildTourSteps(navLabels, hasAskAi, t), t)
    // Only a new user or role may auto-start; `t` and the label array identity must not.
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [user?.id, labelsKey])

  useEffect(() => {
    if (startRequests === 0) return
    tourRef.current?.destroy()
    tourRef.current = runTour(buildTourSteps(navLabels, hasAskAi, t), t)
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [startRequests])

  useEffect(() => () => tourRef.current?.destroy(), [])

  return null
}

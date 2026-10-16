import { useEffect, useState } from 'react'
import { retryAfterSeconds } from '@/shared/api/retry-after'
import type { ApiError } from '@/shared/api/types'
import { RETRY_WAIT_TICK_MS } from './constants'

/** Seconds left before a retry makes sense: counts down from the `retry_after` of a 429, else 0. */
export function useRetryWait(error: ApiError) {
  const seconds = retryAfterSeconds(error) ?? 0
  // Ticks belong to one error: a new failure starts its own count.
  const [elapsed, setElapsed] = useState({ error, ticks: 0 })

  useEffect(() => {
    if (seconds === 0) return undefined
    let ticks = 0
    const timer = setInterval(() => {
      ticks += 1
      setElapsed({ error, ticks })
      if (ticks >= seconds) clearInterval(timer)
    }, RETRY_WAIT_TICK_MS)
    return () => clearInterval(timer)
  }, [error, seconds])

  return Math.max(0, seconds - (elapsed.error === error ? elapsed.ticks : 0))
}

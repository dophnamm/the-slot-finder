import { dayjs } from '@/utils/functions'
import { AUTO_SIGN_IN_LIMIT, AUTO_SIGN_IN_STORAGE_KEY, AUTO_SIGN_IN_WINDOW_MS } from './constants'

function recentAttempts(now: number): number[] {
  try {
    const saved: unknown = JSON.parse(window.sessionStorage.getItem(AUTO_SIGN_IN_STORAGE_KEY) ?? '[]')
    if (!Array.isArray(saved)) return []
    return saved.filter((time): time is number => typeof time === 'number' && now - time < AUTO_SIGN_IN_WINDOW_MS)
  } catch {
    return []
  }
}

/**
 * Records an automatic sign-in in this tab and says whether it may go ahead: false once
 * AUTO_SIGN_IN_LIMIT of them happened within AUTO_SIGN_IN_WINDOW_MS without the server accepting
 * the session in between.
 */
export function recordAutoSignIn(now: number = dayjs().valueOf()): boolean {
  const attempts = recentAttempts(now)
  if (attempts.length >= AUTO_SIGN_IN_LIMIT) return false
  window.sessionStorage.setItem(AUTO_SIGN_IN_STORAGE_KEY, JSON.stringify([...attempts, now]))
  return true
}

/** Forgets the automatic sign-ins: the server accepted the session, or the user chose to try again. */
export function clearAutoSignIns() {
  window.sessionStorage.removeItem(AUTO_SIGN_IN_STORAGE_KEY)
}

import { ROUTES } from '@/resources/routes'
import type { LoginRedirectState } from './types'

/** Where to go after signing in: the page the user asked for, or home. */
export function returnPath(state: unknown): string {
  const from = (state as Partial<LoginRedirectState> | null)?.from
  return typeof from === 'string' && from.startsWith('/') ? from : ROUTES.dashboard
}

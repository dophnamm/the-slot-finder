import type { Location } from 'react-router'
import type { LoginRedirectState } from '@/shared/auth'

/** Remembers the page the user asked for, so sign-in can return to it. */
export function loginRedirectState({ pathname, search }: Location): LoginRedirectState {
  return { from: `${pathname}${search}` }
}

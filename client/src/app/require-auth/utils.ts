import type { Location } from 'react-router'

/** The page the user asked for, so sign-in can return to it. */
export function returnPath({ pathname, search }: Location): string {
  return `${pathname}${search}`
}

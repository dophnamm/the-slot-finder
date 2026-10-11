import type { UIMatch } from 'react-router'
import type { BreadcrumbHandle, BreadcrumbKind } from './types'

function isBreadcrumbHandle(handle: unknown): handle is BreadcrumbHandle {
  return typeof handle === 'object' && handle !== null && 'breadcrumb' in handle
}

/** The trail of the deepest matched route that declares one. */
export function currentBreadcrumb(matches: UIMatch[]): BreadcrumbKind | undefined {
  return matches.map((match) => match.handle).filter(isBreadcrumbHandle).at(-1)?.breadcrumb
}

import type { NavLinkRenderProps } from 'react-router'
import { can, type Role } from '@/resources/rbac'
import { cn } from '@/utils/functions'
import { NAV_ITEMS } from './constants'

export function navLinkClassName({ isActive }: NavLinkRenderProps) {
  return cn(
    'flex items-center gap-2 rounded-md px-2 py-1.5 text-sm text-muted-foreground transition-colors hover:bg-sidebar-accent hover:text-sidebar-accent-foreground',
    // The current page stands out: pastel blue, darker text, bold.
    isActive &&
      'bg-badge-blue font-semibold text-badge-blue-foreground hover:bg-badge-blue hover:text-badge-blue-foreground',
  )
}

/** The navigation links the role may open. */
export function visibleNavItems(role: Role | '' | undefined) {
  return NAV_ITEMS.filter((item) => can(role, item.permission))
}

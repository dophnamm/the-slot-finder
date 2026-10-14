import type { NavLinkRenderProps } from 'react-router'
import { can, type Role } from '@/resources/rbac'
import { cn } from '@/utils/functions'
import { NAV_ITEMS } from './constants'

/** A 40px row; hover is lighter than the current page, so moving the pointer never looks like a selection. */
export function navLinkClassName({ isActive }: NavLinkRenderProps) {
  return cn(
    'flex h-10 items-center gap-2.5 rounded-xl px-3 text-sm text-foreground/70 transition-colors outline-none focus-visible:ring-2 focus-visible:ring-ring',
    !isActive && 'hover:bg-sidebar-accent hover:text-foreground',
    isActive && 'bg-nav-active font-medium text-foreground',
  )
}

/** The navigation links the role may open. */
export function visibleNavItems(role: Role | '' | undefined) {
  return NAV_ITEMS.filter((item) => can(role, item.permission))
}

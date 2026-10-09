import type { NavLinkRenderProps } from 'react-router'
import { cn } from '@/utils/functions'

export function navLinkClassName({ isActive }: NavLinkRenderProps) {
  return cn(
    'flex items-center gap-2 rounded-md px-2 py-1.5 text-sm transition-colors hover:bg-sidebar-accent hover:text-sidebar-accent-foreground',
    isActive && 'bg-sidebar-accent font-medium text-sidebar-accent-foreground',
  )
}

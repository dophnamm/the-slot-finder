import type { LucideIcon } from 'lucide-react'
import type { Permission } from '@/resources/rbac'
import type { PlainKey } from '@/shared/i18n'

export interface NavItem {
  to: string
  label: PlainKey
  icon: LucideIcon
  /** The link shows only for roles with this permission. */
  permission: Permission
}

export interface SidebarProps {
  /** Called when a link is followed, so the mobile panel can close. */
  onNavigate?: () => void
}

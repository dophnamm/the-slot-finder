import type { LucideIcon } from 'lucide-react'
import type { PlainKey } from '@/shared/i18n'

export interface NavItem {
  to: string
  label: PlainKey
  icon: LucideIcon
}

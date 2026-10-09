import {
  CalendarDays,
  FolderKanban,
  GitCompareArrows,
  History,
  Inbox,
  LayoutDashboard,
  Users,
} from 'lucide-react'
import { ROUTES } from '@/resources/routes'
import type { NavItem } from './types'

export const NAV_ITEMS: NavItem[] = [
  { to: ROUTES.dashboard, label: 'Dashboard', icon: LayoutDashboard },
  { to: ROUTES.programs, label: 'Programs', icon: FolderKanban },
  { to: ROUTES.intake, label: 'Intake', icon: Inbox },
  { to: ROUTES.scenarios, label: 'Scenarios', icon: GitCompareArrows },
  { to: ROUTES.people, label: 'People', icon: Users },
  { to: ROUTES.events, label: 'Events', icon: History },
  { to: ROUTES.mySchedule, label: 'My schedule', icon: CalendarDays },
]

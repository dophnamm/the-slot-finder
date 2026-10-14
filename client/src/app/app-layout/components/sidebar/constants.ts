import { CalendarDays, FolderKanban, GitCompareArrows, History, Inbox, LayoutDashboard, Users } from 'lucide-react'
import { ROUTES } from '@/resources/routes'
import type { NavItem } from './types'

export const NAV_ITEMS: NavItem[] = [
  { to: ROUTES.dashboard, label: 'Dashboard', icon: LayoutDashboard, permission: 'dashboard:view' },
  { to: ROUTES.programs, label: 'Programs', icon: FolderKanban, permission: 'programs:view' },
  { to: ROUTES.intake, label: 'Intake', icon: Inbox, permission: 'intake:view' },
  { to: ROUTES.scenarios, label: 'Scenarios', icon: GitCompareArrows, permission: 'scenarios:view' },
  { to: ROUTES.people, label: 'People', icon: Users, permission: 'people:view' },
  { to: ROUTES.events, label: 'Events', icon: History, permission: 'events:view' },
  { to: ROUTES.mySchedule, label: 'My schedule', icon: CalendarDays, permission: 'my-schedule:view' },
]

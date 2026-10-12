import { CalendarDays, FolderKanban, GitCompareArrows, History, Inbox, LayoutDashboard, Users } from 'lucide-react'
import { ROUTES } from '@/resources/routes'
import type { NavItem } from './types'

export const NAV_ITEMS: NavItem[] = [
  { to: ROUTES.dashboard, label: 'Dashboard', icon: LayoutDashboard, permission: 'dashboard:view' },
  { to: ROUTES.programs, label: 'Programs', icon: FolderKanban, permission: 'programs:view' },
  { to: ROUTES.intake, label: 'Intake', icon: Inbox, permission: 'intake:view' },
  { to: ROUTES.scenarios, label: 'Scenarios', icon: GitCompareArrows, permission: 'scenarios:view' },
  { to: ROUTES.people, label: 'People', icon: Users, permission: 'people:manage' },
  { to: ROUTES.events, label: 'Events', icon: History, permission: 'events:view' },
  { to: ROUTES.mySchedule, label: 'My schedule', icon: CalendarDays, permission: 'my-schedule:view' },
]

/**
 * The page area grows with its content, so long pages scroll with the padding kept at the bottom. A
 * page that marks its root `data-fill-screen` (the dashboard) is held to the screen height from xl
 * up instead, so its own table scrolls.
 */
export const PAGE_AREA_CLASS = 'flex-1 xl:has-[[data-fill-screen]]:min-h-0'

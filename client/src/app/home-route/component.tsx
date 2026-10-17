import { Navigate } from 'react-router'
import { DashboardPage } from '@/pages/dashboard'
import { NoAccessPage } from '@/pages/no-access'
import { can } from '@/resources/rbac'
import { ROUTES } from '@/resources/routes'
import { useCurrentUser } from '@/shared/auth'
import { PageSkeleton } from '@/shared/components/page-skeleton'

/** Home depends on the role: planners and managers get the dashboard, specialists their schedule. */
export function HomeRoute() {
  const { data: user, isPending } = useCurrentUser()

  if (isPending) return <PageSkeleton />
  if (can(user?.role, 'dashboard:view')) return <DashboardPage />
  if (can(user?.role, 'my-schedule:view')) return <Navigate to={ROUTES.mySchedule} replace />
  return <NoAccessPage />
}

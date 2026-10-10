import { Outlet } from 'react-router'
import { NoAccessPage } from '@/pages/no-access'
import { can } from '@/resources/rbac'
import { useCurrentUser } from '@/shared/auth'
import type { RequirePermissionProps } from './types'

/** Renders the child routes when the user's role has the permission; the server checks again. */
export function RequirePermission({ permission }: RequirePermissionProps) {
  const { data: user, isPending } = useCurrentUser()

  if (isPending) return null
  if (!can(user?.role, permission)) return <NoAccessPage />
  return <Outlet />
}

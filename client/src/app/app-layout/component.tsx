import { LogOut } from 'lucide-react'
import { NavLink, Outlet } from 'react-router'
import { ROLE_LABELS } from '@/resources/rbac'
import { ROUTES } from '@/resources/routes'
import { signOut, useCurrentUser } from '@/shared/auth'
import { useTranslation } from '@/shared/i18n'
import { Button } from '@/shared/ui/button'
import { Separator } from '@/shared/ui/separator'
import { navLinkClassName, visibleNavItems } from './utils'

export function AppLayout() {
  const { t } = useTranslation()
  const { data: user } = useCurrentUser()
  const navItems = visibleNavItems(user?.role)

  const handleSignOut = () => {
    void signOut()
  }

  return (
    <div className="flex h-svh overflow-hidden">
      <aside className="flex w-56 shrink-0 flex-col gap-3 overflow-y-auto border-r bg-sidebar p-3 text-sidebar-foreground">
        <div className="px-2 pt-1 font-heading text-base font-semibold">{t('The Slot Finder')}</div>
        <Separator />
        <nav className="flex flex-col gap-1">
          {navItems.map(({ to, label, icon: Icon }) => (
            <NavLink key={to} to={to} end={to === ROUTES.dashboard} className={navLinkClassName}>
              <Icon className="size-4" />
              {t(label)}
            </NavLink>
          ))}
        </nav>
        <div className="mt-auto flex flex-col gap-2">
          <Separator />
          {user && (
            <div className="flex flex-col px-2 text-sm">
              <span className="truncate font-medium">{t('Signed in as {name}', { name: user.name || user.username })}</span>
              {user.role && <span className="text-muted-foreground">{t(ROLE_LABELS[user.role])}</span>}
            </div>
          )}
          <Button variant="ghost" className="justify-start" onClick={handleSignOut}>
            <LogOut />
            {t('Sign out')}
          </Button>
        </div>
      </aside>
      <main className="min-w-0 flex-1 overflow-y-auto p-6">
        <Outlet />
      </main>
    </div>
  )
}

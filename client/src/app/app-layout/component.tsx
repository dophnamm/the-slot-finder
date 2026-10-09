import { NavLink, Outlet } from 'react-router'
import { ROUTES } from '@/resources/routes'
import { useTranslation } from '@/shared/i18n'
import { Separator } from '@/shared/ui/separator'
import { NAV_ITEMS } from './constants'
import { navLinkClassName } from './utils'

export function AppLayout() {
  const { t } = useTranslation()

  return (
    <div className="flex min-h-svh">
      <aside className="flex w-56 shrink-0 flex-col gap-3 border-r bg-sidebar p-3 text-sidebar-foreground">
        <div className="px-2 pt-1 font-heading text-base font-semibold">{t('The Slot Finder')}</div>
        <Separator />
        <nav className="flex flex-col gap-1">
          {NAV_ITEMS.map(({ to, label, icon: Icon }) => (
            <NavLink key={to} to={to} end={to === ROUTES.dashboard} className={navLinkClassName}>
              <Icon className="size-4" />
              {t(label)}
            </NavLink>
          ))}
        </nav>
      </aside>
      <main className="min-w-0 flex-1 p-6">
        <Outlet />
      </main>
    </div>
  )
}

import { NavLink } from 'react-router'
import { ROUTES } from '@/resources/routes'
import { useCurrentUser } from '@/shared/auth'
import { useTranslation } from '@/shared/i18n'
import { AccountMenu } from './components/account-menu'
import { Brand } from './components/brand'
import type { SidebarProps } from './types'
import { navLinkClassName, visibleNavItems } from './utils'

/** The white column: the product, the pages the role may open, and the signed-in account at the bottom. */
export function Sidebar({ onNavigate }: SidebarProps) {
  const { t } = useTranslation()
  const { data: user } = useCurrentUser()
  const navItems = visibleNavItems(user?.role)

  return (
    <div className="flex h-full w-full flex-col bg-background text-foreground">
      <div className="flex h-16 shrink-0 items-center border-b px-5">
        <Brand />
      </div>
      <nav aria-label={t('Navigation')} className="flex min-h-0 flex-1 flex-col gap-1 overflow-y-auto p-3">
        {navItems.map(({ to, label, icon: Icon }) => (
          <NavLink key={to} to={to} end={to === ROUTES.dashboard} className={navLinkClassName} onClick={onNavigate}>
            <Icon className="size-4 shrink-0" />
            {t(label)}
          </NavLink>
        ))}
      </nav>
      <div className="shrink-0 p-3">
        <AccountMenu />
      </div>
    </div>
  )
}

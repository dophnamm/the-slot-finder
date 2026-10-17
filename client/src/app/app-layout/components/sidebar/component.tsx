import { NavLink } from 'react-router'
import { can } from '@/resources/rbac'
import { ROUTES } from '@/resources/routes'
import { useCurrentUser } from '@/shared/auth'
import { useTranslation } from '@/shared/i18n'
import { useOpenDrafts } from '@/shared/staffing'
import { Badge } from '@/shared/ui/badge'
import { Skeleton } from '@/shared/ui/skeleton'
import { LOADING_REVEAL_CLASS } from '@/utils/constants'
import { cn, navTourAnchor, range } from '@/utils/functions'
import { AccountMenu } from './components/account-menu'
import { Brand } from './components/brand'
import { NAV_LOADING_ITEMS } from './constants'
import type { SidebarProps } from './types'
import { navLinkClassName, visibleNavItems } from './utils'

/** The white column: the product, the pages the role may open, and the signed-in account at the bottom. */
export function Sidebar({ onNavigate }: SidebarProps) {
  const { t } = useTranslation()
  const { data: user, isPending } = useCurrentUser()
  const navItems = visibleNavItems(user?.role)
  const { data: openDrafts = [] } = useOpenDrafts(can(user?.role, 'staffing-drafts:view'))

  return (
    <div className="flex h-full w-full flex-col bg-background text-foreground">
      <div className="flex h-16 shrink-0 items-center border-b px-5">
        <Brand />
      </div>
      <nav aria-label={t('Navigation')} className="flex min-h-0 flex-1 flex-col gap-1 overflow-y-auto p-3">
        {isPending &&
          range(NAV_LOADING_ITEMS).map((item) => (
            <div key={item} aria-hidden className={cn('flex h-10 items-center gap-2.5 px-3', LOADING_REVEAL_CLASS)}>
              <Skeleton className="size-4 shrink-0" />
              <Skeleton className="h-3.5 w-24" />
            </div>
          ))}
        {navItems.map(({ to, label, icon: Icon }) => (
          <NavLink
            key={to}
            to={to}
            data-tour={navTourAnchor(label)}
            end={to === ROUTES.dashboard}
            className={navLinkClassName}
            onClick={onNavigate}
          >
            <Icon className="size-4 shrink-0" />
            {t(label)}
            {to === ROUTES.drafts && openDrafts.length > 0 && (
              <Badge variant="secondary" className="ml-auto tabular-nums">
                {openDrafts.length}
              </Badge>
            )}
          </NavLink>
        ))}
      </nav>
      <div className="shrink-0 p-3">
        <AccountMenu />
      </div>
    </div>
  )
}

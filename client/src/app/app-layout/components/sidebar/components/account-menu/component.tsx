import { ChevronsUpDown, LogOut } from 'lucide-react'
import { ROLE_LABELS } from '@/resources/rbac'
import { signOut, useCurrentUser } from '@/shared/auth'
import { useTranslation } from '@/shared/i18n'
import { Button } from '@/shared/ui/button'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuGroup,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/shared/ui/dropdown-menu'
import { Skeleton } from '@/shared/ui/skeleton'
import { LOADING_REVEAL_CLASS } from '@/utils/constants'
import { cn } from '@/utils/functions'
import { initials } from './utils'

/** The signed-in person as one row; it opens a menu with the role and Sign out. */
export function AccountMenu() {
  const { t } = useTranslation()
  const { data: user, isPending } = useCurrentUser()

  const handleSignOut = () => {
    void signOut()
  }

  if (isPending)
    return (
      <div aria-hidden className={cn('flex h-10 items-center gap-2.5 px-1.5', LOADING_REVEAL_CLASS)}>
        <Skeleton className="size-7 shrink-0 rounded-full" />
        <Skeleton className="h-3.5 w-28" />
      </div>
    )
  if (!user) return null
  const name = user.name || user.username

  return (
    <DropdownMenu>
      <DropdownMenuTrigger
        render={<Button variant="ghost" className="h-10 w-full justify-start gap-2.5 rounded-xl px-1.5" />}
      >
        <span className="grid size-7 shrink-0 place-items-center rounded-full bg-secondary text-xs font-medium">
          {initials(name)}
        </span>
        <span className="min-w-0 flex-1 truncate text-left text-sm font-medium">{name}</span>
        <ChevronsUpDown className="size-4 shrink-0 text-muted-foreground" />
      </DropdownMenuTrigger>
      <DropdownMenuContent side="top" align="start" className="w-(--anchor-width) rounded-xl p-1">
        <DropdownMenuGroup>
          <DropdownMenuLabel className="flex flex-col gap-0.5 px-3 py-2">
            <span className="truncate text-sm font-medium text-foreground">{t('Signed in as {name}', { name })}</span>
            {user.role && <span>{t(ROLE_LABELS[user.role])}</span>}
          </DropdownMenuLabel>
        </DropdownMenuGroup>
        <DropdownMenuSeparator />
        <DropdownMenuItem className="h-9 rounded-lg px-3" onClick={handleSignOut}>
          <LogOut />
          {t('Sign out')}
        </DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  )
}

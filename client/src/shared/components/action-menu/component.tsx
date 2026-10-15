import { MoreHorizontal } from 'lucide-react'
import type { MouseEvent } from 'react'
import { useTranslation } from '@/shared/i18n'
import { confirm } from '@/shared/stores/confirm-store'
import { Button } from '@/shared/ui/button'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/shared/ui/dropdown-menu'
import type { ActionMenuProps } from './types'

/**
 * The record's less common actions (hold, close, cancel) behind one ⋯ button, so a page header keeps a
 * single primary action. Destructive actions come last, after a line; irreversible ones ask first.
 */
export function ActionMenu<A extends string>({ items, disabled, onAction }: ActionMenuProps<A>) {
  const { t } = useTranslation()
  const regular = items.filter((item) => !item.destructive)
  const destructive = items.filter((item) => item.destructive)

  const handleSelect = async (event: MouseEvent<HTMLElement>) => {
    const item = items.find((candidate) => candidate.action === event.currentTarget.dataset.action)
    if (!item) return
    const confirmed =
      !item.confirmText ||
      (await confirm({
        title: item.label,
        description: item.confirmText,
        confirmLabel: item.label,
        destructive: item.destructive,
      }))
    if (confirmed) onAction(item.action)
  }

  if (items.length === 0) return null

  return (
    <DropdownMenu>
      <DropdownMenuTrigger
        render={<Button variant="outline" size="icon-sm" aria-label={t('More actions')} disabled={disabled} />}
      >
        <MoreHorizontal />
      </DropdownMenuTrigger>
      <DropdownMenuContent align="end" className="w-auto min-w-44 rounded-xl p-1">
        {regular.map((item) => (
          <DropdownMenuItem
            key={item.action}
            data-action={item.action}
            className="h-9 rounded-lg px-3"
            onClick={handleSelect}
          >
            {item.label}
          </DropdownMenuItem>
        ))}
        {regular.length > 0 && destructive.length > 0 && <DropdownMenuSeparator />}
        {destructive.map((item) => (
          <DropdownMenuItem
            key={item.action}
            data-action={item.action}
            variant="destructive"
            className="h-9 rounded-lg px-3"
            onClick={handleSelect}
          >
            {item.label}
          </DropdownMenuItem>
        ))}
      </DropdownMenuContent>
    </DropdownMenu>
  )
}

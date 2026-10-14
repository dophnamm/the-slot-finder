import { LoadingBar } from '@/shared/components/loading-bar'
import { TableFrameProvider } from '@/shared/ui/table'
import { cn } from '@/utils/functions'
import {
  PAGE_BODY_CLASS,
  SECTION_CARD_CLASS,
  SECTION_DESCRIPTION_CLASS,
  SECTION_FLUSH_BODY_CLASS,
  SECTION_PADDED_BODY_CLASS,
  SECTION_TITLE_CLASS,
} from './constants'
import type { SectionProps } from './types'

/**
 * A card on the gray page: the heading, its description and actions sit inside the card, above the body.
 * The `page` variant is the page's own header instead: the h1 with its actions, and the body in a card
 * underneath.
 */
export function Section({
  title,
  description,
  action,
  children,
  footer,
  card = false,
  variant = 'section',
  className,
  contentClassName,
  refreshing = false,
}: SectionProps) {
  const hasBody = Boolean(children || footer)

  if (variant === 'page') {
    return (
      <section className={cn('flex min-h-0 flex-col gap-4', className)}>
        <div className="flex flex-wrap items-start justify-between gap-x-4 gap-y-3">
          <div className="flex min-w-0 flex-1 flex-col gap-1">
            <h1 className={SECTION_TITLE_CLASS.page}>{title}</h1>
            {description && <p className={SECTION_DESCRIPTION_CLASS}>{description}</p>}
          </div>
          {action && <div className="flex shrink-0 flex-wrap items-center gap-2">{action}</div>}
        </div>
        {hasBody && (
          <div className={cn(SECTION_CARD_CLASS, PAGE_BODY_CLASS, 'relative', contentClassName)}>
            {refreshing && <LoadingBar />}
            {children}
            {footer && <div className="flex items-center justify-end gap-2">{footer}</div>}
          </div>
        )}
      </section>
    )
  }

  return (
    <section className={cn(SECTION_CARD_CLASS, 'flex min-h-0 flex-col', className)}>
      <div className="flex flex-wrap items-start justify-between gap-x-4 gap-y-2 px-5 py-4">
        <div className="flex min-w-0 flex-col gap-1">
          <h2 className={SECTION_TITLE_CLASS.section}>{title}</h2>
          {description && <p className={SECTION_DESCRIPTION_CLASS}>{description}</p>}
        </div>
        {action && <div className="flex flex-wrap items-center gap-2">{action}</div>}
      </div>
      {hasBody && (
        <TableFrameProvider frame={card ? 'framed' : 'flush'}>
          <div
            className={cn(card ? SECTION_PADDED_BODY_CLASS : SECTION_FLUSH_BODY_CLASS, 'relative', contentClassName)}
          >
            {refreshing && <LoadingBar />}
            {children}
            {footer && <div className="flex items-center justify-end gap-2">{footer}</div>}
          </div>
        </TableFrameProvider>
      )}
    </section>
  )
}

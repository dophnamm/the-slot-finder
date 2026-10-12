import { Card, CardContent, CardFooter } from '@/shared/ui/card'
import { cn } from '@/utils/functions'
import { SECTION_BODY_CLASS, SECTION_TITLE_CLASS } from './constants'
import type { SectionProps } from './types'

/** A heading with its description and actions, above the body: a table, or other content in a card. */
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
}: SectionProps) {
  const hasBody = Boolean(children || footer)

  return (
    <section className={cn('flex min-h-0 flex-col gap-3', className)}>
      <div className="flex flex-wrap items-start justify-between gap-x-4 gap-y-2">
        <div className="flex min-w-0 flex-col gap-1">
          <h2 className={SECTION_TITLE_CLASS[variant]}>{title}</h2>
          {description && <p className="text-sm text-muted-foreground">{description}</p>}
        </div>
        {action && <div className="flex flex-wrap items-center gap-2">{action}</div>}
      </div>
      {hasBody && card && (
        <Card className="min-h-0 flex-1">
          {children && <CardContent className={cn(SECTION_BODY_CLASS, contentClassName)}>{children}</CardContent>}
          {footer && <CardFooter className="justify-end gap-2">{footer}</CardFooter>}
        </Card>
      )}
      {hasBody && !card && (
        <div className={cn(SECTION_BODY_CLASS, contentClassName)}>
          {children}
          {footer && <div className="flex items-center justify-end gap-2">{footer}</div>}
        </div>
      )}
    </section>
  )
}

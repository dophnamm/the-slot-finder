import {
  Card,
  CardAction,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/shared/ui/card'
import type { PagePlaceholderProps } from './types'

/** Stand-in card for a screen that is not built yet. */
export function PagePlaceholder({ title, description, action, children }: PagePlaceholderProps) {
  return (
    <Card>
      <CardHeader>
        <CardTitle>{title}</CardTitle>
        <CardDescription>{description}</CardDescription>
        {action && <CardAction>{action}</CardAction>}
      </CardHeader>
      {children && <CardContent>{children}</CardContent>}
    </Card>
  )
}

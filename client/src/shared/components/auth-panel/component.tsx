import { PagePlaceholder } from '@/shared/components/page-placeholder'
import type { AuthPanelProps } from './types'

/** A small centred card for the screens around sign-in, shown outside the app layout. */
export function AuthPanel({ title, description, children }: AuthPanelProps) {
  return (
    <div className="flex min-h-svh items-center justify-center p-6">
      <div className="w-full max-w-sm">
        <PagePlaceholder title={title} description={description}>
          {children}
        </PagePlaceholder>
      </div>
    </div>
  )
}

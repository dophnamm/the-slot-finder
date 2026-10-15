import { PagePlaceholder } from '@/shared/components/page-placeholder'
import type { AuthPanelProps } from './types'

/** A small centred card on the gray page, like the Keycloak sign-in page, for the screens around sign-in. */
export function AuthPanel({ title, description, children }: AuthPanelProps) {
  return (
    <div className="flex min-h-svh items-center justify-center bg-canvas p-6">
      <div className="w-full max-w-sm">
        <PagePlaceholder title={title} description={description}>
          {children}
        </PagePlaceholder>
      </div>
    </div>
  )
}

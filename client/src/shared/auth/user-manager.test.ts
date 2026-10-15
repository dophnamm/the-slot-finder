import { describe, expect, it } from 'vitest'
import { ROUTES } from '@/resources/routes'
import { userManager } from './user-manager'

describe('userManager', () => {
  it('comes back to the dashboard after sign-out, which goes straight on to the sign-in page', () => {
    expect(userManager.settings.post_logout_redirect_uri).toBe(
      new URL(ROUTES.dashboard, window.location.origin).toString(),
    )
  })
})

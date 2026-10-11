import type { PlainKey } from '@/shared/i18n'
import type { SignInFailure } from './types'

export const SIGN_IN_FAILURE_MESSAGES: Record<SignInFailure, PlainKey> = {
  unreachable: 'Could not reach the sign-in service.',
  repeated: 'The server keeps rejecting your sign-in. Try again, or ask an admin to check the sign-in settings.',
}

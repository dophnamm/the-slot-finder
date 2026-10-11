import type { ProgramAction } from '@/shared/programs'

/** Actions that cannot be undone ask for confirmation first. */
export const CONFIRMED_PROGRAM_ACTIONS: readonly ProgramAction[] = ['cancel', 'close']

import type { ProjectAction } from '@/shared/programs'

/** Actions that end the project ask for confirmation first. */
export const CONFIRMED_PROJECT_ACTIONS: readonly ProjectAction[] = ['complete', 'cancel']

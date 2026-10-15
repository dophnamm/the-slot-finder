import type { PlainKey } from '@/shared/i18n'

/** One-sentence definitions of the words a newcomer would not know (root README §6.5 and §7). */
export const TERM_DEFINITIONS = {
  baseline: 'The plan in force: the staffing everyone works to until a newer draft is accepted.',
  staffingDraft: 'A plan the engine proposes for some projects; nothing changes until it is accepted.',
  stretch:
    'A person below the level the demand needs, or without the skill; added by hand, locked and confirmed on accept.',
  crossExpertise: 'A person working in one of their other skills, at or above the level the demand needs.',
  gapHours: 'Budget hours minus forecast hours; above 0 means the project is understaffed.',
  staffingPercent: 'Planned hours as a share of budget hours.',
  plannedBurnPercent: 'Planned hours in the weeks up to now as a share of budget hours.',
  eac: 'Estimate at completion: the hours used so far divided by Scope %.',
  atRisk: 'Understaffed with the deadline close, or the last phase ends after the deadline.',
  scenario: 'A what-if: the baseline planned again under changes you try; nothing real changes.',
} as const satisfies Record<string, PlainKey>

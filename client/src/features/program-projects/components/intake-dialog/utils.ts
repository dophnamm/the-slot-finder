import { translate } from '@/shared/i18n'
import type { IntakeMasked } from '@/shared/ai'
import { INTAKE_TERM_MAX_LENGTH, INTAKE_TERMS_MAX, INTAKE_TEXT_MAX } from './constants'

/** The names typed in the field: split on commas, trimmed, empties dropped. */
export function parseTerms(raw: string): string[] {
  return raw
    .split(',')
    .map((term) => term.trim())
    .filter(Boolean)
}

/** Whether the terms fit the server limits. */
export function validTerms(terms: string[]): boolean {
  return terms.length <= INTAKE_TERMS_MAX && terms.every((term) => term.length <= INTAKE_TERM_MAX_LENGTH)
}

/** Whether the brief can be sent: not blank and within the limit. */
export function validText(text: string): boolean {
  return text.trim().length > 0 && text.length <= INTAKE_TEXT_MAX
}

/** "Hidden before sending: 2 emails, 1 customer name, 0 phones". */
export function maskedSummary({ emails, customers, phones }: IntakeMasked): string {
  return translate('Hidden before sending: {emails}, {customers}, {phones}', {
    emails:
      emails === 1 ? translate('{count} email', { count: emails }) : translate('{count} emails', { count: emails }),
    customers:
      customers === 1
        ? translate('{count} customer name', { count: customers })
        : translate('{count} customer names', { count: customers }),
    phones:
      phones === 1 ? translate('{count} phone', { count: phones }) : translate('{count} phones', { count: phones }),
  })
}

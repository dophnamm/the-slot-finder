import type { LanguageKey } from '../language-key'

/** A translation table. Any key it leaves out falls back to English (the key itself). */
export type Locale = Partial<LanguageKey>

export const LANGUAGES = ['en'] as const

export type Language = (typeof LANGUAGES)[number]

export const DEFAULT_LANGUAGE: Language = 'en'

/** English is empty on purpose: every key is already its English text. */
export const locales: Record<Language, Locale> = {
  en: {},
}

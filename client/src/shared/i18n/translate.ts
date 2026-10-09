import { useMemo } from 'react'
import type { LanguageKey } from './language-key'
import { useLanguageStore } from './language-store'
import { locales, type Language } from './locales'

export type TranslationKey = keyof LanguageKey

/** Placeholder names in a key: `'Hello {name}'` → `'name'`. */
type ParamNames<K extends string> = K extends `${string}{${infer Name}}${infer Rest}`
  ? Name | ParamNames<Rest>
  : never

export type TranslationParams<K extends TranslationKey> = Record<ParamNames<K>, string | number>

/** Keys without placeholders. Use it to type labels kept in constants. */
export type PlainKey = {
  [K in TranslationKey]: [ParamNames<K>] extends [never] ? K : never
}[TranslationKey]

/** Keys with at least one placeholder. */
export type ParamKey = Exclude<TranslationKey, PlainKey>

/** `t('Dashboard')`, `t('Program {id}', { id })`: params are required exactly when the key has placeholders. */
export interface Translate {
  (key: PlainKey): string
  <K extends ParamKey>(key: K, params: TranslationParams<K>): string
}

type RawParams = Record<string, string | number>

function interpolate(template: string, params?: RawParams): string {
  if (!params) return template
  return template.replace(/\{(\w+)\}/g, (match, name: string) =>
    name in params ? String(params[name]) : match,
  )
}

function resolve(language: Language, key: TranslationKey, params?: RawParams): string {
  return interpolate(locales[language][key] ?? key, params)
}

export function translateIn(language: Language): Translate {
  return ((key: TranslationKey, params?: RawParams) => resolve(language, key, params)) as Translate
}

/** Translates with the current language. For code outside React, such as the HTTP client. */
export const translate = ((key: TranslationKey, params?: RawParams) =>
  resolve(useLanguageStore.getState().language, key, params)) as Translate

/** Translates with the current language and re-renders when it changes. */
export function useTranslation() {
  const language = useLanguageStore((state) => state.language)
  const t = useMemo(() => translateIn(language), [language])
  return { t, language }
}

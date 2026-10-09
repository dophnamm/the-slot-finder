import { create } from 'zustand'
import { DEFAULT_LANGUAGE, type Language } from './locales'

interface LanguageState {
  language: Language
  setLanguage: (language: Language) => void
}

export const useLanguageStore = create<LanguageState>()((set) => ({
  language: DEFAULT_LANGUAGE,
  setLanguage: (language) => set({ language }),
}))

// Keep <html lang> in sync for screen readers and browser features.
useLanguageStore.subscribe(({ language }) => {
  document.documentElement.lang = language
})

---
name: translating-client-text
description: Use when adding or changing any user-facing text in the Slot Finder client (titles, labels, buttons, table headers, toasts, aria-label, placeholder), when adding a language, or when tsc says a string "is not assignable to parameter of type 'PlainKey'" or "'ParamKey'".
---

# Translating client text

## Overview

The key **is** the English text. Every key is declared once in the `LanguageKey` type, and `t()` only accepts declared keys. English needs no locale file, because a missing translation falls back to the key itself. Other languages are optional `Partial<LanguageKey>` tables.

## Quick reference

| Task | Do |
|---|---|
| Show text in a component | `const { t } = useTranslation()` then `t('Team capacity')` |
| Text with a value | Key `'Hello, {name}'` and call `t('Hello, {name}', { name })`. Params are type-checked |
| Text outside React (HTTP client, stores) | `translate('The server is unreachable.')` |
| Label stored in a constant | Type the field as `PlainKey` and call `t(item.label)` when rendering |
| New text | Add `'Exact English text': string` to `src/shared/i18n/language-key.ts`, sorted alphabetically (case-insensitive) |
| Change wording | Rename the key in `LanguageKey`, then fix every place tsc reports, including locales |
| Add a language | See "Adding a language" below |

Import everything from `@/shared/i18n`.

## Core pattern

```tsx
// ❌ dotted keys, raw text, string building
<CardTitle>{t('capacity.title')}</CardTitle>
<p>Hello, {name}</p>
<p>{t('Hello') + ', ' + name}</p>

// ✅ English-text keys with placeholders
<CardTitle>{t('Team capacity')}</CardTitle>
<p>{t('Hello, {name}', { name })}</p>
```

```ts
// language-key.ts: add each key in its alphabetical place
'Hello, {name}': string
'Team capacity': string
```

## Rules

- **Translate UI copy only.** Server data such as people's names, emails, project names and objectives is shown as-is. So are numbers and units like `%`.
- **One whole sentence per key.** Use placeholders instead of joining fragments, because word order changes between languages.
- **Keys are string literals.** `t(\`Program ${id}\`)` cannot be type-checked. Write `t('Program {id}', { id })`.
- **Plurals:** use two keys and choose one in code: `count === 1 ? t('{count} project', { count }) : t('{count} projects', { count })`.
- **Never create an English locale file.** `locales.en` stays `{}`.

## Words to use

One word per idea, the same on every screen:

| Say | Not | Because |
|---|---|---|
| expertise (PM, Analyst, Engineer, QA), primary expertise | skill, primary skill | The demand, cross-expertise and staffing screens all say expertise; `PERSON_SKILL` is only the table name |
| staffing draft | proposal | A proposal is a business document |

## Adding a language

1. Add the code to `LANGUAGES` in `src/shared/i18n/locales/index.ts`, e.g. `['en', 'vi'] as const`.
2. Create `src/shared/i18n/locales/vi.ts` with `export const vi: Locale = { Dashboard: 'Bảng điều khiển', 'Program {id}': 'Chương trình {id}' }`. tsc rejects keys that are not in `LanguageKey`.
3. Register it in the `locales` map. Missing keys fall back to English.
4. Switch the language with `useLanguageStore.getState().setLanguage('vi')`. `<html lang>` follows automatically.

## Placeholders

**Every `Input` and `Textarea` has a placeholder** that shows a real example: an `'e.g. …'` key such as `t('e.g. 8')`, `t('e.g. Checkout Revamp')`, or a sentence for free text (`t('One objective per line, e.g. Ship the new checkout.')`). A number cell in a grid uses a named constant (`'0'`). File, hidden, checkbox and radio inputs are exempt; `DatePicker` and `WeekPicker` already show their own prompt.

## Checks

`npm run lint` rejects raw text in JSX and literal `alt`, `aria-label`, `placeholder` and `title` attributes, and an `Input` or `Textarea` without a placeholder. `npx tsc -b` rejects unknown keys and wrong or missing params.

| tsc error | Cause |
|---|---|
| `... not assignable to parameter of type 'PlainKey'` | The key is not in `LanguageKey` (typo or not added), or it has `{placeholders}` and you passed no params |
| `... not assignable to parameter of type 'ParamKey'` | You passed params to a key that has no placeholders |
| `'x' does not exist in type 'TranslationParams<...>'` | A param name does not match the placeholder |

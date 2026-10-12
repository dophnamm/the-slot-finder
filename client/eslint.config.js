import js from '@eslint/js'
import globals from 'globals'
import reactHooks from 'eslint-plugin-react-hooks'
import reactRefresh from 'eslint-plugin-react-refresh'
import tseslint from 'typescript-eslint'
import { defineConfig, globalIgnores } from 'eslint/config'

// Project conventions (see .claude/skills at the repo root). shadcn/ui files are exempt.
const APP_CODE = ['src/{app,pages,features}/**/*.tsx', 'src/shared/components/**/*.tsx']

const NO_RAW_TEXT = [
  {
    selector: 'JSXText[value=/[A-Za-z]/]',
    message: "User-facing text goes through t() from '@/shared/i18n'.",
  },
  {
    selector: 'JSXAttribute[name.name=/^(alt|aria-label|placeholder|title)$/] > Literal',
    message: "User-facing text goes through t() from '@/shared/i18n'.",
  },
]

const NO_INLINE_FUNCTION_PROPS = [
  {
    selector: 'JSXAttribute > JSXExpressionContainer > :matches(ArrowFunctionExpression, FunctionExpression)',
    message:
      'Declare the function first (a handler in the component body, or a module-level function) and pass it by name.',
  },
]

const NO_NATIVE_DATE_INPUTS = [
  {
    selector: 'JSXAttribute[name.name="type"] > Literal[value=/^(date|week|month|datetime-local)$/]',
    message:
      "The browser's date pickers do not match the app. Use DatePicker or WeekPicker from '@/shared/components'.",
  },
]

const NO_STYLED_LINKS = [
  {
    selector: 'JSXOpeningElement[name.name=/^(Link|a)$/] > JSXAttribute[name.name="className"]',
    message:
      "Text links use TextLink from '@/shared/components/text-link', so every link looks clickable the same way.",
  },
]

const NO_INLINE_STYLE = [
  {
    selector: 'JSXAttribute[name.name="style"]',
    message: 'No inline styles. Use Tailwind classes (and cn() for conditions).',
  },
]

const NO_TYPES_IN_COMPONENT = [
  {
    selector: ':matches(TSInterfaceDeclaration, TSTypeAliasDeclaration, TSEnumDeclaration, TSTypeLiteral)',
    message: 'Declare types in the types.ts next to this component, never in component.tsx.',
  },
]

// PascalCase functions are components; constants such as NAV_ITEMS do not match.
const COMPONENT_NAME = '/^[A-Z][a-z]/'
const ONE_COMPONENT_PER_FILE = [
  {
    selector: `Program > FunctionDeclaration[id.name=${COMPONENT_NAME}]`,
    message: 'component.tsx holds exactly one component, the exported one. Move this one to components/<name>/.',
  },
  {
    selector: `Program > VariableDeclaration > VariableDeclarator[id.name=${COMPONENT_NAME}] > :matches(ArrowFunctionExpression, FunctionExpression)`,
    message: 'component.tsx holds exactly one component, the exported one. Move this one to components/<name>/.',
  },
  {
    selector: `:function :matches(FunctionDeclaration[id.name=${COMPONENT_NAME}], VariableDeclarator[id.name=${COMPONENT_NAME}] > :matches(ArrowFunctionExpression, FunctionExpression))`,
    message: 'Never define a component inside another component. Move it to components/<name>/.',
  },
]

// component.tsx holds imports and one exported component; everything else has its own file.
const ONLY_IMPORTS_AND_COMPONENT = [
  {
    selector: 'Program > :not(ImportDeclaration, ExportNamedDeclaration)',
    message:
      'component.tsx holds only imports and the exported component. Constants go in constants.ts, helpers in utils.ts, types in types.ts.',
  },
  {
    selector: 'Program > ExportNamedDeclaration > VariableDeclaration',
    message: 'Export the component as `export function Name()`. Constants go in constants.ts.',
  },
]

export default defineConfig([
  globalIgnores(['dist']),
  {
    files: ['**/*.{ts,tsx}'],
    extends: [
      js.configs.recommended,
      tseslint.configs.recommended,
      reactHooks.configs.flat.recommended,
      reactRefresh.configs.vite,
    ],
    languageOptions: {
      globals: globals.browser,
    },
  },
  {
    // shadcn/ui components also export their variant helpers (e.g. buttonVariants).
    files: ['src/shared/ui/**/*.tsx'],
    rules: {
      'react-refresh/only-export-components': 'off',
    },
  },
  {
    // dayjs is only loaded with its plugins in one place.
    files: ['src/**/*.{ts,tsx}'],
    ignores: ['src/utils/functions/date.ts'],
    rules: {
      '@typescript-eslint/no-restricted-imports': [
        'error',
        {
          paths: [
            {
              name: 'dayjs',
              message: "Import dayjs and the date helpers from '@/utils/functions'.",
              allowTypeImports: true,
            },
          ],
        },
      ],
    },
  },
  {
    files: APP_CODE,
    rules: {
      'no-restricted-syntax': [
        'error',
        ...NO_RAW_TEXT,
        ...NO_INLINE_STYLE,
        ...NO_NATIVE_DATE_INPUTS,
        ...NO_STYLED_LINKS,
        ...NO_INLINE_FUNCTION_PROPS,
      ],
    },
  },
  {
    // Later configs replace earlier ones for the same rule, so this one repeats the shared selectors.
    files: ['src/**/component.tsx'],
    ignores: ['src/shared/ui/**'],
    rules: {
      'no-restricted-syntax': [
        'error',
        ...NO_RAW_TEXT,
        ...NO_INLINE_STYLE,
        ...NO_NATIVE_DATE_INPUTS,
        ...NO_STYLED_LINKS,
        ...NO_INLINE_FUNCTION_PROPS,
        ...NO_TYPES_IN_COMPONENT,
        ...ONE_COMPONENT_PER_FILE,
        ...ONLY_IMPORTS_AND_COMPONENT,
      ],
    },
  },
  {
    // TextLink is the one place that styles a link; it keeps every other rule.
    files: ['src/shared/components/text-link/component.tsx'],
    rules: {
      'no-restricted-syntax': [
        'error',
        ...NO_RAW_TEXT,
        ...NO_INLINE_STYLE,
        ...NO_NATIVE_DATE_INPUTS,
        ...NO_INLINE_FUNCTION_PROPS,
        ...NO_TYPES_IN_COMPONENT,
        ...ONE_COMPONENT_PER_FILE,
        ...ONLY_IMPORTS_AND_COMPONENT,
      ],
    },
  },
])

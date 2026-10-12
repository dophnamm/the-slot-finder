---
name: structuring-client-components
description: Use when creating, moving, splitting or renaming a page, feature or component in the Slot Finder client, when a component file starts to hold a second component, a constant, a helper or an inline function prop, or when deciding where a type, constant, mock data, hook or child component should live.
---

# Structuring client components

## Overview

Every component lives in its own kebab-case folder, and its exported name comes from the folder name. A folder holds exactly one component. Its types live next to it in `types.ts`, never inside `component.tsx`.

## Folder contract

```text
program-header/
├── index.ts        ← re-exports only: export { ProgramHeader } from './component'
├── component.tsx   ← exactly one component, exported, named after the folder
├── types.ts        ← every type and interface the component uses (props included)
├── constants.ts    ← constants only this component uses (optional)
├── utils.ts        ← helper functions only this component uses (optional)
└── components/     ← optional: child components, each with this same contract
```

- **Name:** PascalCase of the folder. Inside `pages/`, add `Page`: `pages/program-detail/` exports `ProgramDetailPage`. Everywhere else there is no suffix: `program-header/` exports `ProgramHeader`.
- **`types.ts`** exists whenever the component has props or any type. Export public types from `index.ts` too, e.g. `export type { ProgramHeaderProps } from './types'`.
- **Other allowed files:** `mock-data.ts` (placeholder data until the API is wired), `api.ts`, `use-<name>.ts`, and tests next to what they test: `component.test.tsx`, `utils.test.ts`, `use-<name>.test.ts` (Vitest + Testing Library, `npm test`). Nothing else.
- Constants or helpers used by **more than one** component go to `src/utils/`. Endpoints, routes and RBAC go to `src/resources/`. **REQUIRED:** use organizing-client-shared-code for those.

## What component.tsx contains

Only `import` statements and one `export function Name()`. Inside the component, define every function **before** the JSX and pass it by name. JSX props never get an inline function.

```tsx
// ❌ constant, helper and inline function prop in component.tsx
const ITEMS = [...]
function format(x) { ... }
export function TeamList() {
  return <Button onClick={() => setOpen(true)}>...</Button>
}

// ✅ ITEMS in constants.ts, format in utils.ts, handler declared before use
export function TeamList() {
  const [open, setOpen] = useState(false)
  const handleOpen = () => setOpen(true)
  return <Button onClick={handleOpen}>...</Button>
}
```

- A handler that needs state or props lives in the component body. One that does not lives in `utils.ts`, e.g. a `NavLink` `className` function.
- `.map()` callbacks that render list items are not props and are allowed.

## Never a component inside a component

`component.tsx` defines exactly one component, the exported one. A second component, whether at the top level or inside the exported one, moves to `components/<child>/component.tsx`, even if it is small.

```tsx
// ❌ program-detail/component.tsx
function Stat() { ... }
export function ProgramDetailPage() {
  const Row = () => <li />
  ...
}

// ✅ program-detail/components/stat/component.tsx exports Stat
//    program-detail/components/row/component.tsx exports Row
```

## Where a component goes

| Used by | Put it in |
|---|---|
| One page only | `pages/<page>/components/<child>/` |
| One feature only | `features/<feature>/components/<child>/` |
| Two or more pages or features | `shared/components/<name>/` |
| A generic UI primitive (button, dialog, ...) | `shared/ui/` via `npx shadcn@latest add <name>` |

- **Pages** read route params and compose features and their own child components. Server data and business UI live in **features**.
- **Features never import each other.** If two features need the same thing, move it to `shared/`.

## Nesting: three layers at most

`owner/` (1) → `owner/components/child/` (2) → `child/components/grandchild/` (3). A fourth level is not allowed. Flatten the tree, or move the piece to `shared/components/`.

## Checks

Format with `npm run format` (Prettier) before committing; `npm run lint` fails on unformatted files.

`npm run lint` runs ESLint, the Prettier check and then `scripts/check-structure.mjs`. Together they reject:
- types declared in `component.tsx`, including inline object types
- a second or nested component
- constants, helpers or other statements outside the exported component
- inline functions in JSX props
- unexpected file names
- folders that are not kebab-case
- a component folder without `index.ts`
- a fourth nesting level

## Common mistakes

| Mistake | Fix |
|---|---|
| `CapacityBar.tsx` or `index.tsx` holding the component | `capacity-bar/component.tsx` + `capacity-bar/index.ts` |
| `interface Props` at the top of `component.tsx` | Move it to `types.ts` and import it with `import type` |
| A private helper component next to the exported one | `components/<helper>/` |
| A feature importing another feature | Move the shared part to `shared/components/` |
| `const NAV_ITEMS = [...]` above the component | `constants.ts` next to it |
| `onClick={() => ...}` | `const handleX = () => ...` before the JSX, then `onClick={handleX}` |

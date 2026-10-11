---
name: styling-client-components
description: Use when styling anything in the Slot Finder client (layout, spacing, colors, conditional or state-based looks, dynamic sizes such as bar widths) or when a shadcn/ui component needs to look different, including when tempted to write style={...} or a deep selector into a component's internals.
---

# Styling client components

## Overview

Styles are Tailwind classes, never `style={...}`. Colors come from theme tokens. When a shadcn/ui component needs a new look, give it a **variant**. Do not reach into its internals from outside.

## Quick reference

| Need | Do |
|---|---|
| Layout and spacing | `flex` / `grid` with `gap-*`, rather than margins between siblings |
| Conditional classes | `cn('base classes', isOverloaded && 'text-destructive')` |
| Colors | Theme tokens: `bg-primary`, `text-muted-foreground`, `bg-destructive`, `border`. No hex values or `bg-[#...]` |
| A color the theme lacks (e.g. success, warning) | Add a token to `src/index.css`: `--success` in both `:root` and `.dark`, plus `--color-success: var(--success)` in `@theme inline`. Then use `bg-success`. No raw palette classes like `bg-green-500` |
| A new look for a shadcn component | Add a `variant` with `cva` in `src/shared/ui/<component>.tsx`, then use `<Progress variant="destructive" />` |
| A size from data (bar width, fill %) | A component that takes a value prop (`<Progress value={load} />`), or a fixed set of complete class names chosen in code |
| Responsive | `sm:` / `md:` / `lg:` prefixes |

## Core pattern

```tsx
// ❌ inline style, dynamic class strings, deep selector into shadcn internals
<div style={{ width: `${load}%`, background: 'red' }} />
<div className={`w-[${load}%]`} />
<Progress className="**:data-[slot=progress-indicator]:bg-destructive" value={load} />

// ✅ component with a value prop + a variant defined once in shared/ui/progress.tsx
<Progress value={load} variant={load > 100 ? 'destructive' : 'default'} />
```

## Rules

- **`shared/ui` is project code.** shadcn copies its components into the repo so that they can be edited. Add variants there instead of overriding from outside. Keep the shadcn structure and `data-slot` attributes intact.
- **Tailwind only sees complete class names.** It cannot generate `` `w-[${x}%]` `` at runtime. Pick from complete strings in code, e.g. `const WIDTH = { half: 'w-1/2', full: 'w-full' }`.
- **No `!important`** (`!` prefix) to win against a component. Use a variant or `cn()` order instead.
- **No new CSS files.** Theme changes go in the CSS variables in `src/index.css`.

## Page layout, tables and badges

- **Blocks are `Section`s** (`@/shared/components/section`): the title, description and actions sit above the body, never inside a card header. Use `variant="page"` for the header at the top of a page.
- **A table is never inside a card.** Put it straight in the `Section`; the bordered table is its own frame. Pass `card` only for other content (figures, lists, text).
- **Tables are bordered** by `shared/ui/table`: frame, cell lines, grey header. Do not add your own borders or header colours. To scroll a table in place (fixed header, fixed first column), use `<Table scroll="both">` in a flex column, and give sticky cells an opaque background (`bg-card`, `bg-muted`) so rows do not show through.
- **Badges are pastel and pick their colour by meaning**: `blue`, `green`, `amber`, `red`, `violet`, `teal`, `pink`, `gray` (tokens `--badge-*` in `index.css`). Map a status or kind to a colour in a `Record<…, BadgeVariant>` constant next to its labels; do not use `secondary`, `outline` or `default` for statuses.
- Group rows and other bands use `bg-muted`. A highlighted column (e.g. this week) is an inset `shadow-[inset_2px_0_0_…]` on its body cells, which moves with sticky cells, not a background.

## Checks

`npm run lint` rejects `style={...}` everywhere outside `src/shared/ui`. shadcn's own files may set CSS variables through `style`.

## Common mistakes

| Mistake | Fix |
|---|---|
| `style={{ width: pct + '%' }}` for a bar | `<Progress value={pct} />` |
| Arbitrary color `text-[#d00]` | `text-destructive` |
| Overriding a shadcn component's inside with `**:` or `[&_...]` selectors | Add a variant to the component in `shared/ui` |
| Margins on siblings to space a list | `flex flex-col gap-4` on the parent |
| `<Card><CardHeader><CardTitle>…` around a block | `<Section title=… description=… action=…>` |
| A `<Table>` inside a `Card` | The table directly in a `Section` (no `card`) |
| `<Badge variant="secondary">` for a status | A pastel variant chosen from a `BadgeVariant` map |

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

- **The page is gray, blocks are white cards.** The app shell puts pages on `bg-canvas`; the sidebar and the header bar are white. Each block is a `Section` (`@/shared/components/section`): a white card with a hairline border and no shadow, its title, description and actions inside the card header.
- **`variant="page"` is the page's own header**: the `<h1>` with its badges, one meta line, one primary button at most, `Edit` as an outline button, and every other record action (hold, close, cancel) in `ActionMenu` (`@/shared/components/action-menu`, the ⋯ button). Its body, if any, sits in a card under it.
- **A table runs from edge to edge of its card.** Put it straight in the `Section` (without `card`): the section tells the table to drop its own frame, so there is one frame, not a frame in a frame. Pass `card` only for other content (figures, lists, text). A table outside a section keeps a rounded frame.
- **Tables have lines between rows only** (`shared/ui/table`): no cell lines, a white sticky header with small muted text. Do not add your own borders or header colours; a month calendar adds a faint `border-l` per day column. To scroll a table in place, use `<Table scroll="both">` in a flex column, and give sticky cells an opaque background so rows do not show through.
- **A page made of one block fills the screen**: wrap it in `PageFill` (`@/shared/components/page-fill`) and give its `Section` `className="min-h-0 flex-1"` with `<Table scroll="both">`. From lg up the page does not scroll; the table scrolls under its sticky header.
- **A detail page splits its parts into tabs** (`Tabs` with `TabsList variant="line"`), the first tab opening on what people come for; each tab holds one or more cards. A list of the same kind of items inside a card is one list with lines between items, not a box per item.
- **An empty or loading table keeps its frame.** Put `<TableLoading columns={X} />` and `<TableEmpty columns={X} icon={…} title={…} description={…} />` as the last rows of `<TableBody>` (`@/shared/components/table-loading`, `table-empty`), with the column count as a named constant. Say what is missing and, when the user can act, what to do. Content that is not a table uses `<EmptyState framed />`. Never a grey `<p>` under the table.
- **Badges are pastel and pick their colour by meaning**: `blue`, `green`, `amber`, `red`, `violet`, `teal`, `pink`, `gray` (tokens `--badge-*` in `index.css`). Map a status or kind to a colour in a `Record<…, BadgeVariant>` constant next to its labels; do not use `secondary`, `outline` or `default` for statuses.
- **Text that opens a page is a `TextLink`** (`@/shared/components/text-link`): blue (`--link`) with a light underline, darker on hover; `external` opens another site in a new tab. Lint rejects `<Link className>` and `<a className>` elsewhere. Buttons that navigate stay buttons (`render={<Link />}`).
- **A row that opens a detail page is a `LinkRow`** (`@/shared/components/link-row`) when its other cells only show data: the whole row is clickable, its buttons keep their own action. Keep the name as a `<Link>` for keyboard users.
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
| `<p>No items yet.</p>` under an empty table | `<TableEmpty>` as a row inside `<TableBody>` |
| `<Badge variant="secondary">` for a status | A pastel variant chosen from a `BadgeVariant` map |

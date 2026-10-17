# Guided tour

**Goal:** a first-time user gets a short tour of the left sidebar: one step per link the role can open, a step on the Ask AI button for the roles that can use it, then a step on the account menu. The account menu can start the tour again.

**Roles:** everyone who signs in (planner, lead, specialist). Each role gets steps only for the links it can see.
**Target:** 17 Oct · **Depends on:** [sign-in](sign-in.md) · **Branch:** `feat/guided-tour` · **Status:** Done, not yet verified in a browser (see Still to verify)

## In scope

- **Steps.** One step per sidebar link the signed-in role can see, in sidebar order, then one step on the Ask AI button when the role has the `ai:ask` permission (planner and Manager; the button does not exist for a specialist), then one step on the account menu. The steps come from the same nav items the sidebar renders (`visibleNavItems`), so permissions decide them. Each step has the link's label as its title and a one-sentence description from `NAV_DESCRIPTIONS`.
- **Auto-start once per user.** On the first sign-in in a browser, the tour starts by itself on the page the user lands on. The browser remembers it per user id (`slot-finder:tour-seen:<user id>` in localStorage). The flag is written when the tour starts, so a reload never repeats it.
- **No auto-start** when:
  - storage is unavailable, on read or on write: the tour never starts on its own;
  - the first sidebar link is not on screen, for example below `lg`, where the sidebar is a sheet that is closed. The flag stays unwritten, so a later visit on a wide screen still gets the tour.
- **Take the tour.** The account menu has a **Take the tour** item. It starts the tour again on any screen with the same steps, and it does not touch the flag.
- **Closing.** The close button and Esc end the tour. Back and Next move between steps; the last button reads Done.
- **Motion.** With `prefers-reduced-motion: reduce`, the tour runs without animation.
- **Look.** The popover uses the app's tokens (popover, border, muted text, primary button), in light and dark, from one CSS block.
- **Text.** Every string goes through `t()` and `LanguageKey`.
- Client only. No server change.

## Not now

- Tours for other pages, such as the heatmap, a project page or draft review.
- Showing the tour again after new features ship.
- A mobile layout for the tour. Below `lg` the tour does not open the sidebar sheet.
- Translating driver.js's own "N of M" progress text. It stays in English.
- Syncing the seen flag across devices. It is per browser.

## Rules

- The flag is per user id and per browser. Signing out does not clear it, and another user on the same browser gets their own tour.
- A link the role cannot see is never a step, because the steps are built from the sidebar's own nav items.
- A nav item without a `NAV_DESCRIPTIONS` entry still gets a step, but its description is its label. Add the entry with the item.
- The step anchors are `data-tour` attributes: `nav-<label>` on each sidebar link (`navTourAnchor`) `ask-ai` on the Ask AI button (`ASK_AI_TOUR_ANCHOR`) and `account-menu` on the account menu trigger (`ACCOUNT_TOUR_ANCHOR`). Both names are defined once, in `utils/`.
- The account menu does not import the tour. It bumps a counter in `shared/stores/tour-store`, and `features/guided-tour` starts the tour when the counter changes.

## Server

- No API or server change. The seen flag lives only in the browser.

## Client

- `features/guided-tour`:
  - `component.tsx` (`GuidedTour`, renders nothing, two effects: auto-start and start on request);
  - `utils.ts` (steps, the auto-start check, the storage reads and writes, `runTour`);
  - `constants.ts` (storage prefix, popover class, nav descriptions), `types.ts`, `utils.test.ts`.
- `shared/stores/tour-store.ts`: `startRequests` counter and `requestStart()`.
- `utils/constants/tour.ts`: `ACCOUNT_TOUR_ANCHOR`. `utils/functions/tour.ts`: `navTourAnchor`.
- `app/app-layout`: mounts `<GuidedTour navLabels>` beside the Ask AI widget. The sidebar adds `data-tour` to each link. The account menu adds the anchor to its trigger and the **Take the tour** item.
- `index.css`: the `.driver-popover.slot-finder-tour` rules restyle the driver.js popover with the app tokens.
- Dependency: `driver.js` 1.10.1 (MIT), in `client/package.json`.

## Acceptance criteria

- [ ] A specialist's tour has fewer steps than a planner's, and each step names a link that role can open.
- [ ] Every step has a title and a description.
- [ ] A new user signs in and sees the tour once on the first page; a reload or a second sign-in does not start it.
- [ ] A planner's or Manager's tour has a step on the Ask AI button; a specialist's tour has none.
- [ ] **Take the tour** in the account menu starts the tour again, from any page.
- [ ] Esc ends the tour, and so does the close button.
- [ ] With storage blocked, the tour does not start on its own, and **Take the tour** still starts it.
- [ ] Below `lg`, with the sidebar closed, the tour does not start, and the flag stays unset.
- [ ] With reduced motion on, the popover appears without animation.
- [ ] The popover uses the app's colours in light and dark.
- [ ] Every tour string is a `LanguageKey`; none is typed inline.

## Tasks

**Client**
- [x] Steps from the sidebar's nav items, with descriptions in `NAV_DESCRIPTIONS`
- [x] Auto-start with the per-user flag, written before the tour starts; storage and first-link checks
- [x] The Ask AI step, only for roles with `ai:ask`
- [x] **Take the tour** in the account menu, through `tour-store`
- [x] `data-tour` anchors on the sidebar links and the account menu trigger
- [x] Popover styles in `index.css`, light and dark, from the app tokens
- [x] Texts in `LanguageKey`
- [x] `driver.js` added to `client/package.json`

**Tests**
- [x] `buildTourSteps`: step count per role, anchors, titles and descriptions (`utils.test.ts`)
- [x] `shouldAutoStart`: each flag and anchor combination (`utils.test.ts`)
- [ ] The two effects in `component.tsx`: auto-start once, and **Take the tour** starts it again

**Still to verify**
- [ ] Popover look in light and dark. The class names match the installed driver.js 1.10.1 stylesheet, but the CSS was written from memory and has not been seen in a browser
- [ ] First-run auto-start in a browser: sign in as a new user, see the tour once, reload and see no repeat
- [ ] The **Take the tour** item in the account menu, and that the menu closes before the account step is shown
- [ ] The `eslint-disable-next-line react-hooks/exhaustive-deps` comments on the two effects. The first leaves out `t` and the label array on purpose; the second leaves out `navLabels` and `t`
- [ ] No tests of the effects: only the pure functions in `utils.ts` are tested

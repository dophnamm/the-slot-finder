/**
 * The page area grows with its content, so long pages scroll with the padding kept at the bottom. A
 * page that marks its root `data-fill-screen` (the dashboard, and one-block pages through PageFill) is
 * held to the screen height from lg up instead, so its own table scrolls.
 */
export const PAGE_AREA_CLASS = 'flex-1 lg:has-[[data-fill-screen]]:min-h-0'

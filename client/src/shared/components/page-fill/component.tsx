import type { PageFillProps } from './types'

/**
 * A page made of one block (a list or a table): from lg up it is held to the screen height, so the page
 * does not scroll and the block's own table scrolls under a sticky header. Narrower screens scroll the page.
 */
export function PageFill({ children }: PageFillProps) {
  return (
    <div data-fill-screen className="flex flex-col gap-5 lg:h-full">
      {children}
    </div>
  )
}

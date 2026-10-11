import { Heatmap } from '@/features/heatmap'
import { Risks } from '@/features/risks'

/**
 * From xl up the dashboard fills the screen: the heatmap scrolls inside its own table and the risks
 * list sits beside it. Narrower screens stack them and the page scrolls.
 */
export function DashboardPage() {
  return (
    <div data-fill-screen className="flex flex-col gap-4 xl:h-full xl:flex-row">
      <div className="min-w-0 xl:min-h-0 xl:flex-1">
        <Heatmap />
      </div>
      <div className="order-first xl:order-last xl:min-h-0 xl:w-80 xl:shrink-0 2xl:w-96">
        <Risks />
      </div>
    </div>
  )
}

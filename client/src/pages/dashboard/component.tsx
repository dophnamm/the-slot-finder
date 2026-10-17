import { useState } from 'react'
import { CapacityForecast } from '@/features/capacity-forecast'
import { Heatmap } from '@/features/heatmap'
import { Risks } from '@/features/risks'
import { ViewToggle } from './components/view-toggle'
import { DEFAULT_DASHBOARD_VIEW } from './constants'

/**
 * From xl up the dashboard fills the screen: the heatmap scrolls inside its own table and the risks
 * list sits beside it. Narrower screens stack them and the page scrolls. The heatmap card shows people
 * or expertises; the week range is shared, so switching keeps it.
 */
export function DashboardPage() {
  const [view, setView] = useState(DEFAULT_DASHBOARD_VIEW)
  const toggle = <ViewToggle value={view} onChange={setView} />

  return (
    <div data-fill-screen className="flex flex-col gap-4 xl:h-full xl:flex-row">
      <div className="min-w-0 xl:min-h-0 xl:flex-1">
        {view === 'person' ? <Heatmap action={toggle} /> : <CapacityForecast action={toggle} />}
      </div>
      <div className="order-first xl:order-last xl:min-h-0 xl:w-80 xl:shrink-0 2xl:w-96">
        <Risks />
      </div>
    </div>
  )
}

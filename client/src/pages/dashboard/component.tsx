import { Heatmap } from '@/features/heatmap'
import { Risks } from '@/features/risks'

export function DashboardPage() {
  return (
    <div className="grid gap-4 2xl:grid-cols-4">
      <div className="min-w-0 2xl:col-span-3">
        <Heatmap />
      </div>
      <div className="order-first 2xl:order-last">
        <Risks />
      </div>
    </div>
  )
}

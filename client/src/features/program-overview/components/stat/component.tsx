import type { StatProps } from './types'

export function Stat({ label, value }: StatProps) {
  return (
    <div>
      <div className="text-xs text-muted-foreground">{label}</div>
      <div className="text-lg font-medium tabular-nums">{value}</div>
    </div>
  )
}

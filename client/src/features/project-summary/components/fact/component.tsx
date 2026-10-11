import type { FactProps } from './types'

/** One labelled value of the project summary. */
export function Fact({ label, children }: FactProps) {
  return (
    <div className="flex flex-col gap-1">
      <dt className="text-xs text-muted-foreground">{label}</dt>
      <dd className="text-sm">{children}</dd>
    </div>
  )
}

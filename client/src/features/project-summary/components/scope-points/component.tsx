import { useTranslation } from '@/shared/i18n'
import { Input } from '@/shared/ui/input'
import { Slider } from '@/shared/ui/slider'
import { DONE_POINTS_STEP } from './constants'
import type { ScopePointsProps } from './types'
import { useDonePoints } from './use-done-points'

/** Done points of a project: a slider and a field for whoever may change it, plain text otherwise. */
export function ScopePoints({ project, editable }: ScopePointsProps) {
  const { t } = useTranslation()
  const points = useDonePoints(project)

  if (!editable || project.scope_points <= 0)
    return t('{done} of {scope} points done', { done: project.done_points, scope: project.scope_points })

  return (
    <form className="flex items-center gap-3" onSubmit={points.handleSubmit}>
      <Slider
        className="min-w-24 flex-1"
        value={points.slider}
        min={0}
        max={project.scope_points}
        step={DONE_POINTS_STEP}
        thumbLabel={t('Done points')}
        onValueChange={points.handleSlide}
        onValueCommitted={points.handleSlideEnd}
      />
      <Input
        className="w-20 tabular-nums"
        type="number"
        min={0}
        max={project.scope_points}
        step={DONE_POINTS_STEP}
        aria-label={t('Done points')}
        placeholder={t('e.g. 0')}
        value={points.text}
        onChange={points.handleChange}
        onBlur={points.handleBlur}
      />
      <span className="whitespace-nowrap text-muted-foreground">
        {t('of {scope} points', { scope: project.scope_points })}
      </span>
    </form>
  )
}

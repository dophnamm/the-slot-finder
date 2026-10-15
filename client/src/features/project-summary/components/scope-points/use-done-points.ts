import { useState, type ChangeEvent, type FormEvent } from 'react'
import { toast } from 'sonner'
import { useTranslation } from '@/shared/i18n'
import { useSaveProject, type ProjectDetail } from '@/shared/programs'
import { firstValue, parsePoints, sliderPoints } from './utils'

/**
 * The done points being changed: the dragged or typed value until it is saved (or refused),
 * then the project's own value again. The slider saves when it is let go, the field when it is
 * left or Enter is pressed.
 */
export function useDonePoints(project: ProjectDetail) {
  const { t } = useTranslation()
  const save = useSaveProject(project.program_id, project.id)
  const [draft, setDraft] = useState<string | null>(null)
  const text = draft ?? String(project.done_points)
  const scope = project.scope_points

  const clearDraft = () => setDraft(null)
  const handleSaved = () => toast.success(t('Project updated.'))

  const commit = (points: number | null) => {
    if (save.isPending && save.variables?.done_points === points) return
    if (points === null || points === project.done_points) {
      clearDraft()
      return
    }
    setDraft(String(points))
    save.mutate({ done_points: points }, { onSuccess: handleSaved, onSettled: clearDraft })
  }

  const handleSlide = (value: number | readonly number[]) => setDraft(String(firstValue(value)))
  const handleSlideEnd = (value: number | readonly number[]) => commit(firstValue(value))
  const handleChange = (event: ChangeEvent<HTMLInputElement>) => setDraft(event.target.value)
  const handleBlur = () => commit(parsePoints(text, scope))
  const handleSubmit = (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault()
    commit(parsePoints(text, scope))
  }

  return {
    text,
    slider: [sliderPoints(text, project.done_points, scope)],
    handleSlide,
    handleSlideEnd,
    handleChange,
    handleBlur,
    handleSubmit,
  }
}

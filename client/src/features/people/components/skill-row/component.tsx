import { Trash2 } from 'lucide-react'
import { useTranslation } from '@/shared/i18n'
import { Button } from '@/shared/ui/button'
import { Select } from '@/shared/ui/select'
import { SKILL_LEVELS } from '@/utils/constants'
import { PRIMARY_RADIO_NAME } from './constants'
import type { SkillRowProps } from './types'

export function SkillRow({ index, skill, expertises, expertisesLoading, onChange, onRemove }: SkillRowProps) {
  const { t } = useTranslation()

  const expertiseOptions = expertises.map((expertise) => ({ value: expertise.id, label: expertise.name }))
  const levelOptions = SKILL_LEVELS.map((level) => ({ value: String(level), label: t('Level {level}', { level }) }))
  const handleExpertiseChange = (value: string) => onChange(index, { ...skill, expertise_id: value })
  const handleLevelChange = (value: string) => onChange(index, { ...skill, level: Number(value) })
  const handlePrimaryChange = () => onChange(index, { ...skill, primary: true })
  const handleRemove = () => onRemove(index)

  return (
    <div className="flex items-center gap-2">
      <Select
        className="flex-1"
        aria-label={t('Expertise')}
        options={expertiseOptions}
        value={skill.expertise_id}
        isLoading={expertisesLoading}
        onValueChange={handleExpertiseChange}
      />
      <Select
        aria-label={t('Level')}
        options={levelOptions}
        value={String(skill.level)}
        onValueChange={handleLevelChange}
      />
      <label className="flex items-center gap-2 text-sm">
        <input
          type="radio"
          className="size-4 accent-primary"
          name={PRIMARY_RADIO_NAME}
          checked={skill.primary}
          onChange={handlePrimaryChange}
        />
        {t('Primary')}
      </label>
      <Button type="button" variant="ghost" size="icon-sm" aria-label={t('Remove expertise')} onClick={handleRemove}>
        <Trash2 />
      </Button>
    </div>
  )
}

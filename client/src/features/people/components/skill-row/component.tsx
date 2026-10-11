import { Trash2 } from 'lucide-react'
import type { ChangeEvent } from 'react'
import { useTranslation } from '@/shared/i18n'
import { Button } from '@/shared/ui/button'
import { NativeSelect, NativeSelectOption } from '@/shared/ui/native-select'
import { SKILL_LEVELS } from '@/utils/constants'
import { PRIMARY_RADIO_NAME } from './constants'
import type { SkillRowProps } from './types'

export function SkillRow({ index, skill, expertises, onChange, onRemove }: SkillRowProps) {
  const { t } = useTranslation()

  const handleExpertiseChange = (event: ChangeEvent<HTMLSelectElement>) =>
    onChange(index, { ...skill, expertise_id: event.target.value })
  const handleLevelChange = (event: ChangeEvent<HTMLSelectElement>) =>
    onChange(index, { ...skill, level: Number(event.target.value) })
  const handlePrimaryChange = () => onChange(index, { ...skill, primary: true })
  const handleRemove = () => onRemove(index)

  return (
    <div className="flex items-center gap-2">
      <NativeSelect className="flex-1" aria-label={t('Expertise')} value={skill.expertise_id} onChange={handleExpertiseChange}>
        {expertises.map((expertise) => (
          <NativeSelectOption key={expertise.id} value={expertise.id}>
            {expertise.name}
          </NativeSelectOption>
        ))}
      </NativeSelect>
      <NativeSelect aria-label={t('Level')} value={String(skill.level)} onChange={handleLevelChange}>
        {SKILL_LEVELS.map((level) => (
          <NativeSelectOption key={level} value={String(level)}>
            {t('Level {level}', { level })}
          </NativeSelectOption>
        ))}
      </NativeSelect>
      <label className="flex items-center gap-2 text-sm">
        <input type="radio" className="size-4 accent-primary" name={PRIMARY_RADIO_NAME} checked={skill.primary}
          onChange={handlePrimaryChange} />
        {t('Primary')}
      </label>
      <Button type="button" variant="ghost" size="icon-sm" aria-label={t('Remove skill')} onClick={handleRemove}>
        <Trash2 />
      </Button>
    </div>
  )
}

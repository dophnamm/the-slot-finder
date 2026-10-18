import { Trash2 } from 'lucide-react'
import type { ChangeEvent } from 'react'
import { useTranslation } from '@/shared/i18n'
import { Button } from '@/shared/ui/button'
import { Input } from '@/shared/ui/input'
import { NativeSelect, NativeSelectOption } from '@/shared/ui/native-select'
import { SKILL_LEVELS } from '@/utils/constants'
import { INTAKE_HOURS_MAX } from '../../constants'
import type { DemandRowField, DemandRowProps } from './types'

/** One editable demand of the suggestion, with the brief's quote that supports it. */
export function DemandRow({ demand, expertises, expertisesLoading, onChange, onRemove }: DemandRowProps) {
  const { t } = useTranslation()
  const known = expertises.some((expertise) => expertise.id === demand.expertise_id)

  const handleChange = (event: ChangeEvent<HTMLInputElement | HTMLSelectElement>) =>
    onChange(demand.key, event.target.name as DemandRowField, event.target.value)
  const handleRemove = () => onRemove(demand.key)

  return (
    <li className="flex flex-col gap-1">
      <div className="flex items-center gap-2">
        <NativeSelect
          name="expertise_id"
          aria-label={t('Expertise')}
          className="flex-1"
          value={known ? demand.expertise_id : ''}
          isLoading={expertisesLoading}
          onChange={handleChange}
        >
          <NativeSelectOption value="">{t('Select an expertise')}</NativeSelectOption>
          {expertises.map((expertise) => (
            <NativeSelectOption key={expertise.id} value={expertise.id}>
              {expertise.name}
            </NativeSelectOption>
          ))}
        </NativeSelect>
        <Input
          name="total_hours"
          placeholder={t('e.g. 320')}
          type="number"
          aria-label={t('Budget hours')}
          className="w-28"
          min={1}
          max={INTAKE_HOURS_MAX}
          step="1"
          value={demand.total_hours}
          onChange={handleChange}
        />
        <Input
          name="min_level"
          placeholder={t('e.g. 2')}
          type="number"
          aria-label={t('Minimum level')}
          className="w-20"
          min={SKILL_LEVELS[0]}
          max={SKILL_LEVELS[SKILL_LEVELS.length - 1]}
          step="1"
          value={demand.min_level}
          onChange={handleChange}
        />
        <Button type="button" variant="ghost" size="icon" aria-label={t('Remove demand')} onClick={handleRemove}>
          <Trash2 />
        </Button>
      </div>
      {demand.quote && <p className="text-xs text-muted-foreground italic">“{demand.quote}”</p>}
    </li>
  )
}

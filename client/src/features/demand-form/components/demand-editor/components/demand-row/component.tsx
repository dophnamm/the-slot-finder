import { Trash2 } from 'lucide-react'
import type { ChangeEvent } from 'react'
import { useTranslation } from '@/shared/i18n'
import { Button } from '@/shared/ui/button'
import { Input } from '@/shared/ui/input'
import { NativeSelect, NativeSelectOption } from '@/shared/ui/native-select'
import { TableCell, TableRow } from '@/shared/ui/table'
import { SKILL_LEVELS } from '@/utils/constants'
import { DEMAND_LIMITS } from '../../constants'
import type { DemandRowProps, DemandTextField } from './types'

export function DemandRow({ index, draft, expertises, phases, window, onChange, onRemove }: DemandRowProps) {
  const { t } = useTranslation()

  const handleChange = (event: ChangeEvent<HTMLInputElement | HTMLSelectElement>) => {
    const field = event.target.name as DemandTextField
    onChange(index, { ...draft, [field]: event.target.value })
  }
  const handleLevelChange = (event: ChangeEvent<HTMLSelectElement>) => onChange(index, { ...draft, min_level: Number(event.target.value) })
  const handleRemove = () => onRemove(index)

  return (
    <TableRow>
      <TableCell>
        <NativeSelect name="expertise_id" aria-label={t('Expertise')} value={draft.expertise_id} onChange={handleChange}>
          {expertises.map((expertise) => (
            <NativeSelectOption key={expertise.id} value={expertise.id}>
              {expertise.name}
            </NativeSelectOption>
          ))}
        </NativeSelect>
      </TableCell>
      <TableCell>
        <NativeSelect name="phase_id" aria-label={t('Phase')} value={draft.phase_id} onChange={handleChange}>
          <NativeSelectOption value="">{t('Whole project')}</NativeSelectOption>
          {phases.map((phase) => (
            <NativeSelectOption key={phase.id} value={phase.id}>
              {phase.name}
            </NativeSelectOption>
          ))}
        </NativeSelect>
      </TableCell>
      <TableCell>
        <NativeSelect aria-label={t('Minimum level')} value={String(draft.min_level)} onChange={handleLevelChange}>
          {SKILL_LEVELS.map((level) => (
            <NativeSelectOption key={level} value={String(level)}>
              {t('Level {level}', { level })}
            </NativeSelectOption>
          ))}
        </NativeSelect>
      </TableCell>
      <TableCell>
        <Input name="headcount" type="number" className="w-20" aria-label={t('Headcount')} min={DEMAND_LIMITS.minHeadcount}
          max={DEMAND_LIMITS.maxHeadcount} step="1" value={draft.headcount} onChange={handleChange} />
      </TableCell>
      <TableCell>
        <Input name="total_hours" type="number" className="w-28" aria-label={t('Budget hours')} min={0} max={DEMAND_LIMITS.maxHours}
          step="any" value={draft.total_hours} onChange={handleChange} />
      </TableCell>
      <TableCell className="text-sm text-muted-foreground">{t('{start} – {end}', { start: window.from, end: window.to })}</TableCell>
      <TableCell>
        <Button type="button" variant="ghost" size="icon-sm" aria-label={t('Remove demand')} onClick={handleRemove}>
          <Trash2 />
        </Button>
      </TableCell>
    </TableRow>
  )
}

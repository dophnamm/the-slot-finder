import type { ChangeEvent } from 'react'
import { useLeadOptions } from '@/shared/programs'
import { useTranslation } from '@/shared/i18n'
import { NativeSelect, NativeSelectOption } from '@/shared/ui/native-select'
import type { LeadSelectProps } from './types'

/** Picks a planner or lead as the lead of a program or project. Only planners see it. */
export function LeadSelect({ id, value, onChange }: LeadSelectProps) {
  const { t } = useTranslation()
  const { data: options = [] } = useLeadOptions(true)

  const handleChange = (event: ChangeEvent<HTMLSelectElement>) => onChange(event.target.value)

  return (
    <NativeSelect id={id} className="w-full" value={value} onChange={handleChange}>
      <NativeSelectOption value="">{t('No lead')}</NativeSelectOption>
      {options.map((option) => (
        <NativeSelectOption key={option.id} value={option.id}>
          {option.name}
        </NativeSelectOption>
      ))}
    </NativeSelect>
  )
}

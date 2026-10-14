import type { ChangeEvent } from 'react'
import { useLeadOptions } from '@/shared/programs'
import { useTranslation } from '@/shared/i18n'
import { NativeSelect, NativeSelectOption } from '@/shared/ui/native-select'
import type { LeadSelectProps } from './types'

/** Picks a planner or manager to manage a program or lead a project. Only planners see it. */
export function LeadSelect({ id, value, emptyLabel = 'No lead', onChange }: LeadSelectProps) {
  const { t } = useTranslation()
  const { data: options = [], isPending } = useLeadOptions(true)

  const handleChange = (event: ChangeEvent<HTMLSelectElement>) => onChange(event.target.value)

  return (
    <NativeSelect id={id} className="w-full" value={value} isLoading={isPending} onChange={handleChange}>
      <NativeSelectOption value="">{t(emptyLabel)}</NativeSelectOption>
      {options.map((option) => (
        <NativeSelectOption key={option.id} value={option.id}>
          {option.name}
        </NativeSelectOption>
      ))}
    </NativeSelect>
  )
}

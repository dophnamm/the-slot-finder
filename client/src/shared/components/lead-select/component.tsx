import { useLeadOptions } from '@/shared/programs'
import { useTranslation } from '@/shared/i18n'
import { Select } from '@/shared/ui/select'
import type { LeadSelectProps } from './types'

/** Picks a planner or manager to manage a program or lead a project. Only planners see it. */
export function LeadSelect({ id, value, emptyLabel = 'No lead', onChange }: LeadSelectProps) {
  const { t } = useTranslation()
  const { data: options = [], isPending } = useLeadOptions(true)

  const selectOptions = [
    { value: '', label: t(emptyLabel) },
    ...options.map((option) => ({ value: option.id, label: option.name })),
  ]

  return (
    <Select
      id={id}
      className="w-full"
      options={selectOptions}
      value={value}
      isLoading={isPending}
      onValueChange={onChange}
    />
  )
}

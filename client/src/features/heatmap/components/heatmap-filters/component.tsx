import { Search } from 'lucide-react'
import type { ChangeEvent } from 'react'
import { WeekRangeFilters } from '@/shared/components/week-range-filters'
import { useExpertises } from '@/shared/expertises'
import { useTranslation } from '@/shared/i18n'
import { useProgramOptions } from '@/shared/programs'
import { Input } from '@/shared/ui/input'
import { Label } from '@/shared/ui/label'
import { NativeSelect, NativeSelectOption } from '@/shared/ui/native-select'
import { useHeatmapFilters } from '../../use-heatmap-filters'

/** Week range, expertise and program of the heatmap. */
export function HeatmapFilters() {
  const { t } = useTranslation()
  const { data: expertises = [], isPending: expertisesLoading } = useExpertises()
  const { data: programs = [], isPending: programsLoading } = useProgramOptions()
  const { expertiseId, programId, search, setExpertiseId, setProgramId, setSearch } = useHeatmapFilters()

  const handleExpertiseChange = (event: ChangeEvent<HTMLSelectElement>) => setExpertiseId(event.target.value)
  const handleProgramChange = (event: ChangeEvent<HTMLSelectElement>) => setProgramId(event.target.value)
  const handleSearchChange = (event: ChangeEvent<HTMLInputElement>) => setSearch(event.target.value)

  return (
    <div className="flex flex-wrap items-end gap-4">
      <div className="flex flex-col gap-2">
        <Label htmlFor="heatmap-search">{t('Search')}</Label>
        <div className="relative">
          <Search className="pointer-events-none absolute top-1/2 left-2.5 size-4 -translate-y-1/2 text-muted-foreground" />
          <Input
            id="heatmap-search"
            type="search"
            className="w-56 pl-8"
            placeholder={t('Name or expertise')}
            value={search}
            onChange={handleSearchChange}
          />
        </div>
      </div>
      <WeekRangeFilters />
      <div className="flex flex-col gap-2">
        <Label htmlFor="heatmap-expertise">{t('Expertise')}</Label>
        <NativeSelect
          id="heatmap-expertise"
          value={expertiseId}
          isLoading={expertisesLoading}
          onChange={handleExpertiseChange}
        >
          <NativeSelectOption value="">{t('All expertises')}</NativeSelectOption>
          {expertises.map((expertise) => (
            <NativeSelectOption key={expertise.id} value={expertise.id}>
              {expertise.name}
            </NativeSelectOption>
          ))}
        </NativeSelect>
      </div>
      <div className="flex flex-col gap-2">
        <Label htmlFor="heatmap-program">{t('Program')}</Label>
        <NativeSelect id="heatmap-program" value={programId} isLoading={programsLoading} onChange={handleProgramChange}>
          <NativeSelectOption value="">{t('All programs')}</NativeSelectOption>
          {programs.map((program) => (
            <NativeSelectOption key={program.id} value={program.id}>
              {program.name}
            </NativeSelectOption>
          ))}
        </NativeSelect>
      </div>
    </div>
  )
}

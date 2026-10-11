import { create } from 'zustand'
import { dayjs, toIsoWeek } from '@/utils/functions'
import { DEFAULT_WEEK_COUNT } from './constants'
import type { HeatmapFiltersState } from './types'

/** Week range, expertise and search text of the heatmap; UI state, not server data. */
export const useHeatmapFilters = create<HeatmapFiltersState>()((set) => ({
  from: toIsoWeek(dayjs()),
  weeks: DEFAULT_WEEK_COUNT,
  expertiseId: '',
  search: '',
  setFrom: (from) => set({ from }),
  setWeeks: (weeks) => set({ weeks }),
  setExpertiseId: (expertiseId) => set({ expertiseId }),
  setSearch: (search) => set({ search }),
}))

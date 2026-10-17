import { create } from 'zustand'
import type { HeatmapFiltersState } from './types'

/** Expertise, program and search text of the heatmap; UI state, not server data. */
export const useHeatmapFilters = create<HeatmapFiltersState>()((set) => ({
  expertiseId: '',
  programId: '',
  search: '',
  setExpertiseId: (expertiseId) => set({ expertiseId }),
  setProgramId: (programId) => set({ programId }),
  setSearch: (search) => set({ search }),
}))

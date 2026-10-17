import { create } from 'zustand'
import { DEFAULT_WEEK_COUNT } from '@/utils/constants'
import { dayjs, toIsoWeek } from '@/utils/functions'
import type { WeekRangeState } from './types'

/** One range for the heatmap and the capacity forecast, so switching views keeps it. */
export const useWeekRange = create<WeekRangeState>()((set) => ({
  from: toIsoWeek(dayjs()),
  weeks: DEFAULT_WEEK_COUNT,
  setFrom: (from) => set({ from }),
  setWeeks: (weeks) => set({ weeks }),
}))

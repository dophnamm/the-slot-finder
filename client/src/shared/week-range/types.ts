/** First week and number of weeks shared by the dashboard's views; UI state, not server data. */
export interface WeekRangeState {
  from: string
  weeks: number
  setFrom: (from: string) => void
  setWeeks: (weeks: number) => void
}

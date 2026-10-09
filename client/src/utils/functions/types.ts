import type { Dayjs } from 'dayjs'

/** First day (Monday) and last day (Sunday) of an ISO week. */
export interface IsoWeekRange {
  start: Dayjs
  end: Dayjs
}

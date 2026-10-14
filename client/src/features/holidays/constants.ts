export const HOLIDAYS_QUERY_KEY = ['holidays'] as const

/** Columns of the table; the loading and empty rows span them all. */
export const HOLIDAY_TABLE_COLUMNS = 4

/** Multipart fields of POST /holidays/import; they match the server's form. */
export const HOLIDAY_IMPORT_FIELDS = {
  file: 'file',
  url: 'url',
  from: 'from_date',
  to: 'to_date',
  hours: 'hours_per_day',
} as const

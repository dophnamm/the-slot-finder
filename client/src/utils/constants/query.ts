/** Retries for a failed query before the shared error handler shows it. */
export const QUERY_RETRY = 1

/** How long fetched server data counts as fresh. */
export const QUERY_STALE_TIME_MS = 30_000

/** The server's largest page size, used to fetch short lists such as people for a picker. */
export const MAX_PAGE_SIZE = 100

/** Prefix of every insights query (heatmap, a person's week), refreshed when the baseline changes. */
export const INSIGHTS_QUERY_KEY = ['insights'] as const

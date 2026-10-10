/** Number of pages for a total, at least one so the pager never shows "of 0". */
export function pageCount(total: number | undefined, pageSize: number): number {
  return Math.max(1, Math.ceil((total ?? 0) / pageSize))
}

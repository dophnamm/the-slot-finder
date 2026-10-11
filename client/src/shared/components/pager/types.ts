export interface PagerProps {
  /** The current page, starting at 1. */
  page: number
  /** Items across all pages, from the response's `meta.total`. */
  total: number | undefined
  pageSize: number
  onPageChange: (page: number) => void
}

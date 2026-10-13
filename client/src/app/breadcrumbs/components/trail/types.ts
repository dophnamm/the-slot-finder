/** A step with its text ready to show. The last step is the current page; a step before it without `to` is plain text. */
export interface Crumb {
  label: string
  to?: string
}

export interface TrailProps {
  crumbs: Crumb[]
}

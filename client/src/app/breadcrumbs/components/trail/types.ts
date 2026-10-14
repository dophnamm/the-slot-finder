/** A step with its text ready to show. The last step is the current page; a step before it without `to` is plain text. */
export interface Crumb {
  label: string
  to?: string
  /** Its name is still loading: a skeleton bar shows, and the label is read out to screen readers. */
  pending?: boolean
}

export interface TrailProps {
  crumbs: Crumb[]
}

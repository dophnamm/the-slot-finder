/** A step with its text ready to show; the last step is the current page and is not a link. */
export interface Crumb {
  label: string
  to?: string
}

export interface TrailProps {
  crumbs: Crumb[]
}

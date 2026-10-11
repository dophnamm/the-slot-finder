export interface ReplanBannerProps {
  programId: string
}

/** The open re-plan draft for a program and how many changes led to it. */
export interface OpenReplan {
  draftId: string
  changes: number
}

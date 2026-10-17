export interface DraftReviewProps {
  draftId: string
}

/** A person the Planner can add to a demand by hand. */
export interface StaffablePerson {
  id: string
  name: string
  active: boolean
  skills: { expertise: string; level: number; primary: boolean }[]
}

/** The written summary of a staffing draft; `template` means the AI was unavailable. */
export interface DraftSummary {
  summary: string
  source: 'ai' | 'template'
}

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

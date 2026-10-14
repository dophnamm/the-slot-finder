export interface AddMemberDialogProps {
  open: boolean
  programId: string
  /** People already in the team; they are left out of the list. */
  teamIds: string[]
  onOpenChange: (open: boolean) => void
}

/** A person the planner may add, with their primary expertise when they have one. */
export interface CandidateOption {
  id: string
  name: string
  expertise: string | null
}

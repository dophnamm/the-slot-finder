export interface IntakeDialogProps {
  open: boolean
  /** The program the project is created in. */
  programId: string
  onOpenChange: (open: boolean) => void
}

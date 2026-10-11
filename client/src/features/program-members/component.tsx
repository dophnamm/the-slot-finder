import { MembersTable } from '@/shared/components/members-table'
import { Section } from '@/shared/components/section'
import { useTranslation } from '@/shared/i18n'
import { useProgramMembers } from '@/shared/programs'
import type { ProgramMembersProps } from './types'

/** Everyone with baseline hours on the program: expertise, hours, weeks and projects. */
export function ProgramMembers({ programId }: ProgramMembersProps) {
  const { t } = useTranslation()
  const { data: members = [], isPending } = useProgramMembers(programId)

  return (
    <Section title={t('Members')} description={t('People with hours in the baseline on this program’s projects, most hours first.')}>
      <MembersTable members={members} isPending={isPending} programId={programId} />
    </Section>
  )
}

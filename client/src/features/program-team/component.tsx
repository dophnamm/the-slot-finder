import { UserPlus } from 'lucide-react'
import { useState } from 'react'
import { toast } from 'sonner'
import { Section } from '@/shared/components/section'
import { useTranslation } from '@/shared/i18n'
import { confirm } from '@/shared/stores/confirm-store'
import { Button } from '@/shared/ui/button'
import { AddMemberDialog } from './components/add-member-dialog'
import { TeamTable } from './components/team-table'
import type { ProgramTeamProps, TeamMember } from './types'
import { useProgramTeam, useRemoveTeamMember } from './use-program-team'

/** Team tab of the program page: the people its projects are staffed from; planners add and remove them. */
export function ProgramTeam({ programId, canManage, canViewPeople }: ProgramTeamProps) {
  const { t } = useTranslation()
  const [adding, setAdding] = useState(false)
  const { data: members = [], isPending } = useProgramTeam(programId)
  const remove = useRemoveTeamMember(programId)
  const teamIds = members.map((member) => member.person.id)

  const handleAdd = () => setAdding(true)
  const handleRemoved = () => toast.success(t('Member removed.'))
  const handleRemove = async (member: TeamMember) => {
    const confirmed = await confirm({
      title: t('Remove {name} from the team?', { name: member.person.name }),
      description: t('Their baseline hours stay; the next draft plans without them.'),
      confirmLabel: t('Remove'),
      destructive: true,
    })
    if (!confirmed) return
    remove.mutate(member.person.id, { onSuccess: handleRemoved })
  }

  return (
    <Section
      title={t('Team')}
      description={t('The people this program’s projects are staffed from. A draft proposes only them.')}
      action={
        canManage && (
          <Button onClick={handleAdd}>
            <UserPlus />
            {t('Add member')}
          </Button>
        )
      }
    >
      <TeamTable
        members={members}
        isPending={isPending}
        programId={programId}
        canManage={canManage}
        canViewPeople={canViewPeople}
        removingId={remove.isPending ? remove.variables : undefined}
        onRemove={handleRemove}
      />
      {canManage && <AddMemberDialog open={adding} programId={programId} teamIds={teamIds} onOpenChange={setAdding} />}
    </Section>
  )
}

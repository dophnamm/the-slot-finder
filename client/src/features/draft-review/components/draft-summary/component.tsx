import { useState } from 'react'
import { Section } from '@/shared/components/section'
import { useTranslation } from '@/shared/i18n'
import { Button } from '@/shared/ui/button'
import { useSummarizeDraft } from '../../use-summarize-draft'
import type { DraftSummaryProps, SummaryAnswer } from './types'

/** A short written summary of the draft, requested on click. */
export function DraftSummary({ draft }: DraftSummaryProps) {
  const { t } = useTranslation()
  const summarize = useSummarizeDraft(draft.id)
  const [answer, setAnswer] = useState<SummaryAnswer | null>(null)
  // Only an answer for this exact draft shows: any change to the draft drops it.
  const result = answer?.draft === draft ? answer.result : null

  const handleSummarize = () => summarize.mutate(undefined, { onSuccess: (data) => setAnswer({ draft, result: data }) })

  return (
    <Section
      card
      title={t('Summary')}
      action={
        <Button variant="outline" size="sm" isLoading={summarize.isPending} onClick={handleSummarize}>
          {t('Summarize this draft')}
        </Button>
      }
      contentClassName="flex flex-col gap-2"
    >
      {result && (
        <>
          <p className="text-sm whitespace-pre-line">{result.summary}</p>
          <p className="text-xs text-muted-foreground">
            {result.source === 'ai' ? t('Written by AI') : t('Template summary')}
          </p>
        </>
      )}
    </Section>
  )
}

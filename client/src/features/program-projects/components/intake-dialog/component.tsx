import { type ChangeEvent, type FormEvent, useState } from 'react'
import { type IntakeSuggestion, useExtractIntake } from '@/shared/ai'
import { useTranslation } from '@/shared/i18n'
import { Button } from '@/shared/ui/button'
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/shared/ui/dialog'
import { Input } from '@/shared/ui/input'
import { Label } from '@/shared/ui/label'
import { Textarea } from '@/shared/ui/textarea'
import { IntakeReview } from '../intake-review'
import { INTAKE_TEXT_MAX } from './constants'
import type { IntakeDialogProps } from './types'
import { parseTerms, validTerms, validText } from './utils'

/** Reads a customer brief with the AI and suggests a project; the person reviews it before anything is saved. */
export function IntakeDialog({ open, programId, onOpenChange }: IntakeDialogProps) {
  const { t } = useTranslation()
  const [text, setText] = useState('')
  const [termsText, setTermsText] = useState('')
  const [suggestion, setSuggestion] = useState<IntakeSuggestion | null>(null)
  const extract = useExtractIntake()
  const terms = parseTerms(termsText)
  const termsOk = validTerms(terms)
  const canSuggest = validText(text) && termsOk

  const handleClose = () => onOpenChange(false)
  const handleText = (event: ChangeEvent<HTMLTextAreaElement>) => setText(event.target.value)
  const handleTerms = (event: ChangeEvent<HTMLInputElement>) => setTermsText(event.target.value)
  const handleSubmit = (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault()
    if (!canSuggest || extract.isPending) return
    extract.mutate({ text, ...(terms.length > 0 && { mask_terms: terms }) }, { onSuccess: setSuggestion })
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-xl">
        <DialogHeader>
          <DialogTitle>{t('New project from a brief')}</DialogTitle>
          <DialogDescription>
            {suggestion
              ? t('Review the suggestion and fix anything wrong. Nothing is saved until you create the project.')
              : t('Customer names, emails and phone numbers are hidden before the text goes to the AI.')}
          </DialogDescription>
        </DialogHeader>
        {suggestion ? (
          <IntakeReview programId={programId} suggestion={suggestion} onClose={handleClose} />
        ) : (
          <form className="flex flex-col gap-4" onSubmit={handleSubmit}>
            <div className="flex flex-col gap-2">
              <Label htmlFor="intake-text">{t('Brief')}</Label>
              <Textarea
                id="intake-text"
                className="max-h-64 min-h-32"
                placeholder={t('Paste the customer’s brief or email here.')}
                maxLength={INTAKE_TEXT_MAX}
                value={text}
                onChange={handleText}
              />
              <p className="text-right text-xs text-muted-foreground">
                {t('{count} / {max} characters', { count: text.length, max: INTAKE_TEXT_MAX })}
              </p>
            </div>
            <div className="flex flex-col gap-2">
              <Label htmlFor="intake-terms">{t('Customer or contact names to hide (comma separated)')}</Label>
              <Input
                id="intake-terms"
                placeholder={t('e.g. Acme, Jane Doe')}
                aria-invalid={!termsOk || undefined}
                value={termsText}
                onChange={handleTerms}
              />
              {!termsOk && (
                <p className="text-xs text-destructive">{t('Use at most 10 names of 100 characters each.')}</p>
              )}
            </div>
            <DialogFooter>
              <Button type="submit" disabled={!canSuggest} isLoading={extract.isPending}>
                {t('Suggest project')}
              </Button>
            </DialogFooter>
          </form>
        )}
      </DialogContent>
    </Dialog>
  )
}

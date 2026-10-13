import { Upload } from 'lucide-react'
import { type ChangeEvent, type FormEvent, useState } from 'react'
import { toast } from 'sonner'
import { WeekPicker } from '@/shared/components/week-picker'
import { useTranslation } from '@/shared/i18n'
import { Button } from '@/shared/ui/button'
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/shared/ui/dialog'
import { Input } from '@/shared/ui/input'
import { Label } from '@/shared/ui/label'
import { ProblemTable } from './components/problem-table'
import { FIELD_IDS, RANGE_FIELDS, TIMESHEET_FILE_ACCEPT } from './constants'
import type { ImportResult, WeekRange } from './types'
import { useDownloadTemplate, useImportTimesheet } from './use-timesheet-import'
import { defaultRange, rowProblems } from './utils'

/** Import timesheet: download the template for past weeks, then upload it with actual hours. */
export function TimesheetImport() {
  const { t } = useTranslation()
  const [open, setOpen] = useState(false)
  const [range, setRange] = useState<WeekRange>(defaultRange)
  const [file, setFile] = useState<File | null>(null)
  const download = useDownloadTemplate()
  const upload = useImportTimesheet()
  const problems = rowProblems(upload.error)

  const handleOpenChange = (next: boolean) => {
    setOpen(next)
    setFile(null)
    upload.reset()
  }
  const handleOpen = () => handleOpenChange(true)
  const handleWeekChange = (value: string, name: string) => setRange((current) => ({ ...current, [name]: value }))
  const handleDownload = () => download.mutate(range)
  const handleFileChange = (event: ChangeEvent<HTMLInputElement>) => {
    setFile(event.target.files?.[0] ?? null)
    upload.reset()
  }
  const handleImported = ({ created, updated, skipped }: ImportResult) => {
    toast.success(t('{created} rows added, {updated} updated, {skipped} skipped.', { created, updated, skipped }))
    handleOpenChange(false)
  }
  const handleSubmit = (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault()
    if (file) upload.mutate(file, { onSuccess: handleImported })
  }

  return (
    <>
      <Button variant="outline" onClick={handleOpen}>
        <Upload />
        {t('Import timesheet')}
      </Button>
      <Dialog open={open} onOpenChange={handleOpenChange}>
        <DialogContent className="sm:max-w-2xl">
          <form className="flex min-h-0 flex-col gap-5" onSubmit={handleSubmit}>
            <DialogHeader>
              <DialogTitle>{t('Import timesheet')}</DialogTitle>
              <DialogDescription>
                {t(
                  'Download the template for past weeks, fill in actual_hours, then import it. One row with a problem saves nothing.',
                )}
              </DialogDescription>
            </DialogHeader>
            <section className="flex flex-col gap-2">
              <h3 className="text-sm font-medium">{t('Download the template')}</h3>
              <div className="flex flex-wrap items-end gap-3">
                <div className="flex flex-col gap-1">
                  <Label htmlFor={FIELD_IDS.from}>{t('From week')}</Label>
                  <WeekPicker
                    id={FIELD_IDS.from}
                    name={RANGE_FIELDS.from}
                    value={range.from}
                    onChange={handleWeekChange}
                  />
                </div>
                <div className="flex flex-col gap-1">
                  <Label htmlFor={FIELD_IDS.to}>{t('To week')}</Label>
                  <WeekPicker id={FIELD_IDS.to} name={RANGE_FIELDS.to} value={range.to} onChange={handleWeekChange} />
                </div>
                <Button type="button" variant="outline" disabled={download.isPending} onClick={handleDownload}>
                  {t('Download template')}
                </Button>
              </div>
            </section>
            <section className="flex flex-col gap-2">
              <h3 className="text-sm font-medium">{t('Import the filled-in file')}</h3>
              <Label htmlFor={FIELD_IDS.file}>{t('Timesheet file')}</Label>
              <Input id={FIELD_IDS.file} type="file" accept={TIMESHEET_FILE_ACCEPT} onChange={handleFileChange} />
            </section>
            {problems.length > 0 && (
              <div className="flex max-h-64 min-h-0 flex-col gap-2">
                <p className="text-sm text-destructive">{t('Nothing was saved. Fix these rows and import again.')}</p>
                <ProblemTable problems={problems} />
              </div>
            )}
            <DialogFooter>
              <Button type="submit" disabled={!file || upload.isPending}>
                {upload.isPending ? t('Importing…') : t('Import')}
              </Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>
    </>
  )
}

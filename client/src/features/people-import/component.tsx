import { Upload } from 'lucide-react'
import { type ChangeEvent, type FormEvent, useState } from 'react'
import { toast } from 'sonner'
import { ProblemTable } from '@/shared/components/problem-table'
import { useTranslation } from '@/shared/i18n'
import { Button } from '@/shared/ui/button'
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/shared/ui/dialog'
import { Input } from '@/shared/ui/input'
import { Label } from '@/shared/ui/label'
import { FILE_FIELD_ID, PEOPLE_FILE_ACCEPT, ROW_PROBLEM_LABELS } from './constants'
import type { PeopleImportResult } from './types'
import { useDownloadPeopleTemplate, useImportPeople } from './use-people-import'
import { rowProblems } from './utils'

/** Import people: download the template, then upload it filled in to add or update people. */
export function PeopleImport() {
  const { t } = useTranslation()
  const [open, setOpen] = useState(false)
  const [file, setFile] = useState<File | null>(null)
  // A new key gives an empty file field, so picking the same (fixed) file again fires a change.
  const [fileFieldKey, setFileFieldKey] = useState(0)
  const download = useDownloadPeopleTemplate()
  const upload = useImportPeople()
  const problems = rowProblems(upload.error)

  const handleOpenChange = (next: boolean) => {
    setOpen(next)
    setFile(null)
    upload.reset()
  }
  const handleOpen = () => handleOpenChange(true)
  const handleDownload = () => download.mutate()
  const handleFileChange = (event: ChangeEvent<HTMLInputElement>) => {
    setFile(event.target.files?.[0] ?? null)
    upload.reset()
  }
  const handleImported = ({ created, updated, unchanged }: PeopleImportResult) => {
    toast.success(t('Created {created}, updated {updated}, unchanged {unchanged}', { created, updated, unchanged }))
    handleOpenChange(false)
  }
  const handleRefused = () => {
    setFile(null)
    setFileFieldKey((key) => key + 1)
  }
  const handleSubmit = (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault()
    if (file) upload.mutate(file, { onSuccess: handleImported, onError: handleRefused })
  }

  return (
    <>
      <Button variant="outline" onClick={handleOpen}>
        <Upload />
        {t('Import people')}
      </Button>
      <Dialog open={open} onOpenChange={handleOpenChange}>
        <DialogContent className="sm:max-w-2xl">
          <form className="flex min-h-0 flex-col gap-5" onSubmit={handleSubmit}>
            <DialogHeader>
              <DialogTitle>{t('Import people')}</DialogTitle>
              <DialogDescription>
                {t(
                  'Download the template, add one row per person, then import the file. A person with a known email is updated, a new email adds a person. One row with a problem saves nothing.',
                )}
              </DialogDescription>
            </DialogHeader>
            <section className="flex flex-col gap-2">
              <h3 className="text-sm font-medium">{t('Download the template')}</h3>
              <div>
                <Button type="button" variant="outline" isLoading={download.isPending} onClick={handleDownload}>
                  {t('Download template')}
                </Button>
              </div>
            </section>
            <section className="flex flex-col gap-2">
              <h3 className="text-sm font-medium">{t('Import the filled-in file')}</h3>
              <Label htmlFor={FILE_FIELD_ID}>{t('People file')}</Label>
              <Input
                key={fileFieldKey}
                id={FILE_FIELD_ID}
                type="file"
                accept={PEOPLE_FILE_ACCEPT}
                onChange={handleFileChange}
              />
            </section>
            {problems.length > 0 && (
              <div className="flex max-h-64 min-h-0 flex-col gap-2">
                <p className="text-sm text-destructive">{t('Nothing was saved. Fix these rows and import again.')}</p>
                <ProblemTable problems={problems} labels={ROW_PROBLEM_LABELS} />
              </div>
            )}
            <DialogFooter>
              <Button type="submit" disabled={!file} isLoading={upload.isPending}>
                {t('Import')}
              </Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>
    </>
  )
}

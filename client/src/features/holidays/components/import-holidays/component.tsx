import { CalendarArrowDown } from 'lucide-react'
import { type ChangeEvent, type FormEvent, useState } from 'react'
import { toast } from 'sonner'
import { DatePicker } from '@/shared/components/date-picker'
import { useTranslation } from '@/shared/i18n'
import { Button } from '@/shared/ui/button'
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/shared/ui/dialog'
import { Input } from '@/shared/ui/input'
import { Label } from '@/shared/ui/label'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/shared/ui/tabs'
import { MAX_HOURS_PER_DAY } from '@/utils/constants'
import type { HolidayImportResult, HolidaySource } from '../../types'
import { useImportHolidays } from '../../use-holidays'
import { HOLIDAY_FILE_ACCEPT, IMPORT_FIELD_IDS } from './constants'
import type { ImportHolidaysValues, ImportTextField } from './types'
import { canImport, initialValues, toInput } from './utils'

/** The "Import holidays" button: a dialog that reads a holiday calendar from an ICS file or a link. */
export function ImportHolidays() {
  const { t } = useTranslation()
  const [open, setOpen] = useState(false)
  const [values, setValues] = useState<ImportHolidaysValues>(initialValues)
  const importing = useImportHolidays()

  const handleOpen = () => setOpen(true)
  const handleOpenChange = (next: boolean) => {
    setOpen(next)
    if (!next) setValues(initialValues())
  }
  const handleSourceChange = (source: HolidaySource) => setValues((current) => ({ ...current, source }))
  const handleFileChange = (event: ChangeEvent<HTMLInputElement>) => {
    const file = event.target.files?.[0] ?? null
    setValues((current) => ({ ...current, file }))
  }
  const handleChange = (event: ChangeEvent<HTMLInputElement>) => {
    const field = event.target.name as ImportTextField
    const { value } = event.target
    setValues((current) => ({ ...current, [field]: value }))
  }
  const handleDateChange = (value: string, name: string) =>
    setValues((current) => ({ ...current, [name as ImportTextField]: value }))
  const handleImported = (result: HolidayImportResult) => {
    toast.success(t('Holidays imported.'), {
      description: t('{added} added, {skipped} already there.', { added: result.added, skipped: result.skipped }),
    })
    handleOpenChange(false)
  }
  const handleSubmit = (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault()
    importing.mutate(toInput(values), { onSuccess: handleImported })
  }

  return (
    <>
      <Button variant="outline" onClick={handleOpen}>
        <CalendarArrowDown />
        {t('Import holidays')}
      </Button>
      <Dialog open={open} onOpenChange={handleOpenChange}>
        <DialogContent className="sm:max-w-lg">
          <form className="flex flex-col gap-5" onSubmit={handleSubmit}>
            <DialogHeader>
              <DialogTitle>{t('Import holidays')}</DialogTitle>
              <DialogDescription>
                {t(
                  'Each event of a holiday calendar between the two dates becomes a company holiday. Holidays imported before are skipped.',
                )}
              </DialogDescription>
            </DialogHeader>
            <Tabs value={values.source} onValueChange={handleSourceChange} className="gap-3">
              <TabsList>
                <TabsTrigger value="file">{t('File')}</TabsTrigger>
                <TabsTrigger value="link">{t('Link')}</TabsTrigger>
              </TabsList>
              <TabsContent value="file" className="flex flex-col gap-2">
                <Label htmlFor={IMPORT_FIELD_IDS.file}>{t('ICS file')}</Label>
                <Input
                  id={IMPORT_FIELD_IDS.file}
                  type="file"
                  accept={HOLIDAY_FILE_ACCEPT}
                  onChange={handleFileChange}
                />
                <p className="text-sm text-muted-foreground">
                  {t('Export a holiday calendar from Google or Outlook as an .ics file.')}
                </p>
              </TabsContent>
              <TabsContent value="link" className="flex flex-col gap-2">
                <Label htmlFor={IMPORT_FIELD_IDS.url}>{t('Calendar link')}</Label>
                <Input
                  placeholder={t('e.g. https://calendar.google.com/calendar/ical/…/public/basic.ics')}
                  id={IMPORT_FIELD_IDS.url}
                  name="url"
                  type="url"
                  value={values.url}
                  onChange={handleChange}
                />
                <p className="text-sm text-muted-foreground">
                  {t(
                    'A public holiday calendar from Google or Outlook, such as Holidays in Vietnam. The link is not saved.',
                  )}
                </p>
              </TabsContent>
            </Tabs>
            <div className="grid gap-4 sm:grid-cols-3">
              <div className="flex flex-col gap-2">
                <Label htmlFor={IMPORT_FIELD_IDS.from}>{t('From')}</Label>
                <DatePicker
                  id={IMPORT_FIELD_IDS.from}
                  name="from_date"
                  value={values.from_date}
                  onChange={handleDateChange}
                />
              </div>
              <div className="flex flex-col gap-2">
                <Label htmlFor={IMPORT_FIELD_IDS.to}>{t('To')}</Label>
                <DatePicker
                  id={IMPORT_FIELD_IDS.to}
                  name="to_date"
                  min={values.from_date}
                  value={values.to_date}
                  onChange={handleDateChange}
                />
              </div>
              <div className="flex flex-col gap-2">
                <Label htmlFor={IMPORT_FIELD_IDS.hours}>{t('Hours per day')}</Label>
                <Input
                  placeholder={t('e.g. 8')}
                  id={IMPORT_FIELD_IDS.hours}
                  name="hours_per_day"
                  type="number"
                  min={0.5}
                  max={MAX_HOURS_PER_DAY}
                  step="0.5"
                  value={values.hours_per_day}
                  onChange={handleChange}
                  required
                />
              </div>
            </div>
            <DialogFooter>
              <Button type="submit" disabled={!canImport(values)} isLoading={importing.isPending}>
                {t('Import holidays')}
              </Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>
    </>
  )
}

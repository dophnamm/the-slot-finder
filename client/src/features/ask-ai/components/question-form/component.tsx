import { Send } from 'lucide-react'
import { useState, type SyntheticEvent } from 'react'
import { useTranslation } from '@/shared/i18n'
import { Button } from '@/shared/ui/button'
import { Input } from '@/shared/ui/input'
import { MAX_QUESTION_LENGTH } from '../../constants'
import type { QuestionFormProps } from './types'

export function QuestionForm({ isPending, onAsk }: QuestionFormProps) {
  const { t } = useTranslation()
  const [question, setQuestion] = useState('')
  const trimmed = question.trim()

  const handleChange = (event: React.ChangeEvent<HTMLInputElement>) => setQuestion(event.target.value)
  const handleSubmit = (event: SyntheticEvent<HTMLFormElement>) => {
    event.preventDefault()
    if (!trimmed || isPending) return
    onAsk(trimmed)
    setQuestion('')
  }

  return (
    <form className="flex gap-2 border-t p-3" onSubmit={handleSubmit}>
      <Input
        value={question}
        maxLength={MAX_QUESTION_LENGTH}
        placeholder={t('e.g. Who is overloaded next week?')}
        aria-label={t('Your question')}
        onChange={handleChange}
      />
      <Button type="submit" size="icon" aria-label={t('Send')} disabled={isPending || !trimmed}>
        <Send />
      </Button>
    </form>
  )
}

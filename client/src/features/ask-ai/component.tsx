import { MessageCircle, Maximize2, Minimize2, X } from 'lucide-react'
import { useState } from 'react'
import { useLocation } from 'react-router'
import { can } from '@/resources/rbac'
import { useAskAi, type AskAnswer } from '@/shared/ai'
import type { ApiError } from '@/shared/api/types'
import { useCurrentUser } from '@/shared/auth'
import { useTranslation } from '@/shared/i18n'
import { Button } from '@/shared/ui/button'
import { cn } from '@/utils/functions'
import { MessageList } from './components/message-list'
import { QuestionForm } from './components/question-form'
import {
  BACKDROP_CLASS,
  BACKDROP_CLOSED_CLASS,
  BACKDROP_OPEN_CLASS,
  MAX_PAGE_CONTEXT_LENGTH,
  PANEL_CLASS,
  PANEL_CLOSED_CLASS,
  PANEL_OPEN_CLASS,
  PANEL_SIZE_CLASS,
} from './constants'
import type { AskAiMessage } from './types'
import { askErrorMessage } from './utils'

/** Floating assistant at the bottom-right corner; messages live in this component only. */
export function AskAi() {
  const { t } = useTranslation()
  const { data: user } = useCurrentUser()
  const { pathname } = useLocation()
  const ask = useAskAi()
  const [isOpen, setIsOpen] = useState(false)
  const [isLarge, setIsLarge] = useState(false)
  const [messages, setMessages] = useState<AskAiMessage[]>([])

  if (!can(user?.role, 'ai:ask')) return null

  const handleToggle = () => setIsOpen((open) => !open)
  const handleResize = () => setIsLarge((large) => !large)
  const handleAnswer = (result: AskAnswer) =>
    setMessages((current) => [...current, { role: 'assistant', text: result.answer }])
  const handleError = (error: ApiError) =>
    setMessages((current) => [...current, { role: 'error', text: askErrorMessage(error) }])
  const handleAsk = (question: string) => {
    setMessages((current) => [...current, { role: 'user', text: question }])
    ask.mutate(
      { question, page_context: pathname.slice(0, MAX_PAGE_CONTEXT_LENGTH) },
      { onSuccess: handleAnswer, onError: handleError },
    )
  }

  return (
    <>
      <div
        className={cn(BACKDROP_CLASS, isOpen ? BACKDROP_OPEN_CLASS : BACKDROP_CLOSED_CLASS)}
        aria-hidden
        onClick={handleToggle}
      />
      <div className="pointer-events-none fixed right-4 bottom-4 z-50 flex flex-col items-end gap-3 lg:right-6 lg:bottom-12">
        <section
          className={cn(
            PANEL_CLASS,
            PANEL_SIZE_CLASS[isLarge ? 'large' : 'normal'],
            isOpen ? PANEL_OPEN_CLASS : PANEL_CLOSED_CLASS,
          )}
          aria-label={t('Ask AI')}
        >
          <header className="flex items-center justify-between border-b py-2 pr-2 pl-4">
            <h2 className="text-sm font-medium">{t('Ask AI')}</h2>
            <Button
              variant="ghost"
              size="icon-sm"
              className="cursor-pointer"
              aria-label={isLarge ? t('Make Ask AI smaller') : t('Make Ask AI larger')}
              aria-pressed={isLarge}
              onClick={handleResize}
            >
              {isLarge ? <Minimize2 /> : <Maximize2 />}
            </Button>
          </header>
          <MessageList messages={messages} isPending={ask.isPending} />
          <QuestionForm isPending={ask.isPending} onAsk={handleAsk} />
        </section>
        <Button
          size="icon-lg"
          className="pointer-events-auto size-12 cursor-pointer rounded-full shadow-lg"
          aria-label={isOpen ? t('Close Ask AI') : t('Open Ask AI')}
          aria-expanded={isOpen}
          onClick={handleToggle}
        >
          {isOpen ? <X /> : <MessageCircle />}
        </Button>
      </div>
    </>
  )
}

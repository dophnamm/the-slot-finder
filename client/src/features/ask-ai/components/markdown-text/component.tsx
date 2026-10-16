import Markdown from 'react-markdown'
import { MARKDOWN_CLASS, REMARK_PLUGINS } from './constants'
import type { MarkdownTextProps } from './types'

/** Renders an assistant answer as markdown; raw HTML in it is not rendered. */
export function MarkdownText({ text }: MarkdownTextProps) {
  return (
    <div className={MARKDOWN_CLASS}>
      <Markdown remarkPlugins={REMARK_PLUGINS}>{text}</Markdown>
    </div>
  )
}

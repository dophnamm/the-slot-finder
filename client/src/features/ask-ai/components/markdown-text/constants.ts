import remarkGfm from 'remark-gfm'

export const REMARK_PLUGINS = [remarkGfm]

/** Spacing and list, table and code styles for the markdown an assistant answer may hold. */
export const MARKDOWN_CLASS = [
  'flex flex-col gap-2 overflow-x-auto break-words',
  '[&_ul]:list-disc [&_ul]:pl-5 [&_ol]:list-decimal [&_ol]:pl-5 [&_li]:my-0.5',
  '[&_strong]:font-semibold [&_a]:text-link [&_a]:underline [&_a]:underline-offset-2',
  '[&_h1]:text-base [&_h1]:font-semibold [&_h2]:text-sm [&_h2]:font-semibold [&_h3]:text-sm [&_h3]:font-medium',
  '[&_code]:rounded [&_code]:bg-background/60 [&_code]:px-1 [&_code]:text-xs',
  '[&_pre]:overflow-x-auto [&_pre]:rounded-md [&_pre]:bg-background/60 [&_pre]:p-2 [&_pre_code]:bg-transparent [&_pre_code]:px-0',
  // The same look as the app tables in shared/ui/table: muted small header, row borders, no border on the last row.
  '[&_table]:w-full [&_table]:border-separate [&_table]:border-spacing-0 [&_table]:text-sm',
  '[&_th]:border-b [&_th]:px-3 [&_th]:py-2 [&_th]:text-left [&_th]:text-xs [&_th]:font-medium [&_th]:whitespace-nowrap [&_th]:text-muted-foreground',
  '[&_td]:border-b [&_td]:px-3 [&_td]:py-2.5 [&_td]:whitespace-nowrap [&_tbody>tr:last-child>td]:border-b-0',
].join(' ')

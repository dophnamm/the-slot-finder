import * as React from 'react'
import { cva, type VariantProps } from 'class-variance-authority'
import { cn } from 'cn'

// A table has lines between rows and a quiet header. The lines sit on the cells (border-separate), so
// they stay with sticky headers and columns. The frame also holds the footer (e.g. the pager), outside
// the part that scrolls. On its own a table has a rounded frame; inside a card section it runs from edge
// to edge of the card instead, so there is one frame, not a frame in a frame.
const tableFrameVariants = cva('relative flex w-full flex-col overflow-hidden', {
  variants: {
    scroll: {
      x: '',
      // Fills the rest of a flex column.
      both: 'min-h-0 flex-1',
    },
    frame: {
      framed: 'rounded-lg border',
      flush: 'border-t',
    },
  },
  defaultVariants: { scroll: 'x', frame: 'framed' },
})

type TableFrame = NonNullable<VariantProps<typeof tableFrameVariants>['frame']>

/** Set by a card section, so the tables inside it drop their own frame. */
const TableFrameContext = React.createContext<TableFrame>('framed')

function TableFrameProvider({ frame, children }: { frame: TableFrame; children: React.ReactNode }) {
  return <TableFrameContext.Provider value={frame}>{children}</TableFrameContext.Provider>
}

const tableContainerVariants = cva('w-full', {
  variants: {
    scroll: {
      x: 'overflow-x-auto',
      // Scrolls both ways, so sticky headers and columns stay in view.
      both: 'min-h-0 flex-1 overflow-auto',
    },
  },
  defaultVariants: { scroll: 'x' },
})

function Table({
  className,
  scroll,
  footer,
  ...props
}: React.ComponentProps<'table'> & VariantProps<typeof tableContainerVariants> & { footer?: React.ReactNode }) {
  const frame = React.useContext(TableFrameContext)
  return (
    <div data-slot="table-frame" data-frame={frame} className={tableFrameVariants({ scroll, frame })}>
      <div data-slot="table-container" className={tableContainerVariants({ scroll })}>
        <table
          data-slot="table"
          className={cn('w-full caption-bottom border-separate border-spacing-0 text-sm', className)}
          {...props}
        />
      </div>
      {footer && (
        <div data-slot="table-footer-bar" className="flex items-center justify-end gap-2 border-t px-5 py-2.5">
          {footer}
        </div>
      )}
    </div>
  )
}

function TableHeader({ className, ...props }: React.ComponentProps<'thead'>) {
  return <thead data-slot="table-header" className={cn(className)} {...props} />
}

function TableBody({ className, ...props }: React.ComponentProps<'tbody'>) {
  return <tbody data-slot="table-body" className={cn('[&>tr:last-child>td]:border-b-0', className)} {...props} />
}

function TableFooter({ className, ...props }: React.ComponentProps<'tfoot'>) {
  return (
    <tfoot
      data-slot="table-footer"
      className={cn('border-t bg-muted/50 font-medium [&>tr]:last:border-b-0', className)}
      {...props}
    />
  )
}

function TableRow({ className, ...props }: React.ComponentProps<'tr'>) {
  return (
    <tr
      data-slot="table-row"
      className={cn(
        'transition-colors hover:bg-muted/50 has-aria-expanded:bg-muted/50 data-[state=selected]:bg-muted',
        className,
      )}
      {...props}
    />
  )
}

function TableHead({ className, ...props }: React.ComponentProps<'th'>) {
  return (
    <th
      data-slot="table-head"
      className={cn(
        // Sticky, so a table that scrolls inside its card keeps its header; opaque, so rows do not show through.
        'sticky top-0 z-10 h-10 border-b bg-background px-3 py-2 text-left align-middle text-xs font-medium whitespace-nowrap text-muted-foreground first:pl-5 last:pr-5 [&:has([role=checkbox])]:pr-0',
        className,
      )}
      {...props}
    />
  )
}

function TableCell({ className, ...props }: React.ComponentProps<'td'>) {
  return (
    <td
      data-slot="table-cell"
      className={cn(
        'border-b px-3 py-2.5 align-middle whitespace-nowrap first:pl-5 last:pr-5 [&:has([role=checkbox])]:pr-0',
        className,
      )}
      {...props}
    />
  )
}

function TableCaption({ className, ...props }: React.ComponentProps<'caption'>) {
  return (
    <caption data-slot="table-caption" className={cn('mt-4 text-sm text-muted-foreground', className)} {...props} />
  )
}

export { Table, TableFrameProvider, TableHeader, TableBody, TableFooter, TableHead, TableRow, TableCell, TableCaption }

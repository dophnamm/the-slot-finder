import * as React from 'react'
import { cva, type VariantProps } from 'class-variance-authority'
import { cn } from 'cn'

// Every table is bordered: a rounded frame, lines between cells and a shaded header. The borders sit on
// the cells (border-separate), so they stay with sticky headers and columns. The frame also holds the
// footer (e.g. the pager), outside the part that scrolls.
const tableFrameVariants = cva('relative flex w-full flex-col overflow-hidden rounded-lg border', {
  variants: {
    scroll: {
      x: '',
      // Fills the rest of a flex column.
      both: 'min-h-0 flex-1',
    },
  },
  defaultVariants: { scroll: 'x' },
})

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
  return (
    <div data-slot="table-frame" className={tableFrameVariants({ scroll })}>
      <div data-slot="table-container" className={tableContainerVariants({ scroll })}>
        <table
          data-slot="table"
          className={cn('w-full caption-bottom border-separate border-spacing-0 text-sm', className)}
          {...props}
        />
      </div>
      {footer && (
        <div
          data-slot="table-footer-bar"
          className="flex items-center justify-end gap-2 border-t bg-muted/40 px-3 py-2"
        >
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
        'h-11 border-r border-b bg-muted px-3 py-2.5 text-left align-middle font-medium whitespace-nowrap text-foreground last:border-r-0 [&:has([role=checkbox])]:pr-0',
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
        'border-r border-b px-3 py-2 align-middle whitespace-nowrap last:border-r-0 [&:has([role=checkbox])]:pr-0',
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

export { Table, TableHeader, TableBody, TableFooter, TableHead, TableRow, TableCell, TableCaption }

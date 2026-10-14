import { cn } from 'cn'

/**
 * A placeholder shape while data loads: a soft overlay that pulses, still for reduced motion. A block-level
 * span, so it may sit inside a heading or a paragraph.
 */
function Skeleton({ className, ...props }: React.ComponentProps<'span'>) {
  return (
    <span
      data-slot="skeleton"
      className={cn('block animate-pulse rounded-md bg-foreground/5 motion-reduce:animate-none', className)}
      {...props}
    />
  )
}

export { Skeleton }

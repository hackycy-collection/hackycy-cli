import type { ComponentProps } from 'react'
import * as Dialog from '@radix-ui/react-dialog'
import { X } from 'lucide-react'
import { cn } from '../../lib/utils'
import { Button } from './button'

export const Sheet = Dialog.Root
export const SheetTrigger = Dialog.Trigger

export function SheetContent({ className, children, ...props }: ComponentProps<typeof Dialog.Content>): React.JSX.Element {
  return (
    <Dialog.Portal>
      <Dialog.Overlay className="fixed inset-0 z-40 bg-black/35" />
      <Dialog.Content
        className={cn('fixed inset-y-0 left-0 z-50 w-[min(88vw,360px)] border-r border-border bg-background shadow-xl outline-none', className)}
        {...props}
      >
        <Dialog.Title className="sr-only">Files</Dialog.Title>
        <Dialog.Description className="sr-only">Comparison file tree</Dialog.Description>
        {children}
        <Dialog.Close asChild>
          <Button className="absolute right-2 top-2" size="icon" variant="ghost" aria-label="Close files">
            <X className="size-4" />
          </Button>
        </Dialog.Close>
      </Dialog.Content>
    </Dialog.Portal>
  )
}

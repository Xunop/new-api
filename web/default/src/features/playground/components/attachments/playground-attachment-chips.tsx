/*
Copyright (C) 2023-2026 QuantumNous

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as
published by the Free Software Foundation, either version 3 of the
License, or (at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program. If not, see <https://www.gnu.org/licenses/>.

For commercial licensing, please contact support@quantumnous.com
*/
import {
  AlertCircleIcon,
  FileIcon,
  ImageIcon,
  Loader2Icon,
  XIcon,
} from 'lucide-react'
import { useTranslation } from 'react-i18next'

import { Button } from '@/components/ui/button'
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from '@/components/ui/tooltip'
import { cn } from '@/lib/utils'

import {
  getAttachmentDetailLabel,
  getAttachmentStatusLabel,
  isAttachmentImage,
} from '../../lib'
import type { PlaygroundAttachment } from '../../types'

type PlaygroundAttachmentChipsProps = {
  attachments: PlaygroundAttachment[]
  className?: string
  currentModel?: string
  onRemove?: (attachment: PlaygroundAttachment) => void
}

function getStatusClass(status: PlaygroundAttachment['status']): string {
  switch (status) {
    case 'failed':
    case 'expired':
    case 'deleted':
      return 'text-destructive'
    case 'uploading':
      return 'text-muted-foreground'
    default:
      return 'text-emerald-600 dark:text-emerald-400'
  }
}

export function PlaygroundAttachmentChips(
  props: PlaygroundAttachmentChipsProps
) {
  const { t } = useTranslation()

  if (props.attachments.length === 0) {
    return null
  }

  return (
    <div className={cn('flex min-w-0 flex-wrap gap-2', props.className)}>
      {props.attachments.map((attachment) => {
        const Icon = isAttachmentImage(attachment) ? ImageIcon : FileIcon
        const statusLabel = getAttachmentStatusLabel(
          attachment.status,
          props.currentModel,
          attachment
        )
        const removable = Boolean(props.onRemove)
        const detail =
          attachment.status === 'failed' && attachment.error
            ? t(attachment.error)
            : t(getAttachmentDetailLabel(attachment, props.currentModel))

        return (
          <div
            key={attachment.id}
            className='border-border/70 bg-background text-foreground flex min-h-9 max-w-full min-w-0 items-center gap-2 rounded-md border px-2.5 py-1.5 text-xs shadow-xs'
          >
            <Icon className='text-muted-foreground size-4 shrink-0' />
            <div className='grid min-w-0 gap-0.5'>
              <span className='truncate font-medium'>
                {attachment.filename}
              </span>
              <span className='text-muted-foreground flex min-w-0 items-center gap-1.5'>
                {attachment.status === 'uploading' && (
                  <Loader2Icon className='size-3 animate-spin' />
                )}
                {attachment.status === 'failed' && (
                  <AlertCircleIcon className='size-3' />
                )}
                <span className={getStatusClass(attachment.status)}>
                  {t(statusLabel)}
                </span>
                <span aria-hidden='true'>·</span>
                <span className='truncate'>{detail}</span>
              </span>
            </div>
            {removable && (
              <Tooltip>
                <TooltipTrigger
                  render={
                    <Button
                      aria-label={t('Remove attachment')}
                      className='text-muted-foreground hover:text-destructive -mr-1 size-6 shrink-0'
                      onClick={() => props.onRemove?.(attachment)}
                      size='icon'
                      type='button'
                      variant='ghost'
                    >
                      <XIcon className='size-3.5' />
                    </Button>
                  }
                />
                <TooltipContent>
                  <p>{t('Remove attachment')}</p>
                </TooltipContent>
              </Tooltip>
            )}
          </div>
        )
      })}
    </div>
  )
}

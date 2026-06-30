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
import type { PlaygroundAttachment } from '../../types'

function isTransientAttachment(attachment: PlaygroundAttachment): boolean {
  if (attachment.status === 'uploading' || attachment.status === 'failed') {
    return true
  }

  return !attachment.id.startsWith('att_')
}

type UploadingAttachmentInput = {
  id: string
  name: string
  size: number
  type: string
}

export function createPendingAttachment(
  sessionId: string,
  file: UploadingAttachmentInput
): PlaygroundAttachment {
  return {
    id: file.id,
    sessionId,
    filename: file.name,
    mimeType: file.type || 'application/octet-stream',
    size: file.size,
    status: 'uploading',
  }
}

export function replacePendingAttachment(
  current: PlaygroundAttachment[],
  pendingId: string,
  uploaded: PlaygroundAttachment
): PlaygroundAttachment[] {
  return current.map((attachment) =>
    attachment.id === pendingId ? { ...uploaded, status: 'ready' } : attachment
  )
}

export function failPendingAttachment(
  current: PlaygroundAttachment[],
  pendingId: string,
  error: string
): PlaygroundAttachment[] {
  return current.map((attachment) =>
    attachment.id === pendingId
      ? {
          ...attachment,
          status: 'failed',
          error,
        }
      : attachment
  )
}

export function mergeListedSessionAttachments(
  current: PlaygroundAttachment[],
  listed: PlaygroundAttachment[],
  sessionId: string
): PlaygroundAttachment[] {
  const transientAttachments = current.filter(
    (attachment) =>
      attachment.sessionId === sessionId && isTransientAttachment(attachment)
  )

  if (listed.length === 0) {
    return transientAttachments
  }

  return [...listed, ...transientAttachments]
}

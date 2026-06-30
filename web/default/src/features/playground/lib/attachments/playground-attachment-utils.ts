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
import type { ContentPart, PlaygroundAttachment } from '../../types'
import { getAttachmentCompatibility } from './attachment-compatibility-utils'

const ATTACHMENT_ERROR_LABELS: Record<string, string> = {
  attachment_access_denied: 'Attachment access denied',
  attachment_expired: 'Attachment has expired',
  attachment_feature_disabled: 'Attachment uploads are disabled',
  attachment_file_too_large: 'Attachment file is too large',
  attachment_invalid_request: 'Invalid attachment request',
  attachment_invalid_signed_reference: 'Attachment link is invalid',
  attachment_mime_not_allowed: 'Attachment file type is not allowed',
  attachment_not_found: 'Attachment was not found',
  attachment_storage_delete_failed: 'Attachment storage delete failed',
  attachment_storage_driver_unavailable: 'Attachment storage is unavailable',
  attachment_storage_read_failed: 'Attachment storage read failed',
  attachment_storage_write_failed: 'Attachment storage write failed',
  attachment_too_many: 'Too many attachments',
}

export function removePendingAttachment(
  attachments: PlaygroundAttachment[],
  attachmentId: string
): PlaygroundAttachment[] {
  return attachments.filter((attachment) => attachment.id !== attachmentId)
}

export function getAttachmentErrorLabel(
  error: unknown,
  fallback = 'Attachment upload failed'
): string {
  if (typeof error === 'string') {
    return ATTACHMENT_ERROR_LABELS[error] ?? fallback
  }

  if (error instanceof Error) {
    return (
      ATTACHMENT_ERROR_LABELS[error.name] ??
      ATTACHMENT_ERROR_LABELS[error.message] ??
      fallback
    )
  }

  return fallback
}

export function isAttachmentReady(attachment: PlaygroundAttachment): boolean {
  return (
    (attachment.status === 'ready' || attachment.status === 'active') &&
    Boolean(attachment.referenceUrl || attachment.id)
  )
}

export function isAttachmentImage(attachment: PlaygroundAttachment): boolean {
  return attachment.mimeType.startsWith('image/')
}

export function formatAttachmentSize(size: number): string {
  if (!Number.isFinite(size) || size <= 0) {
    return '0 B'
  }
  if (size < 1024) {
    return `${size} B`
  }
  if (size < 1024 * 1024) {
    return `${(size / 1024).toFixed(1)} KB`
  }
  return `${(size / (1024 * 1024)).toFixed(1)} MB`
}

export function getAttachmentDetailLabel(
  attachment: PlaygroundAttachment,
  currentModel?: string
): string {
  if (currentModel) {
    const compatibility = getAttachmentCompatibility(attachment, currentModel)
    if (!compatibility.available && compatibility.reason) {
      return compatibility.reason
    }
  }

  return formatAttachmentSize(attachment.size)
}

export function buildAttachmentContentParts(
  attachments: PlaygroundAttachment[] = [],
  currentModel?: string
): ContentPart[] {
  return attachments
    .filter((attachment) => isAttachmentReady(attachment))
    .filter((attachment) =>
      currentModel
        ? getAttachmentCompatibility(attachment, currentModel).available
        : true
    )
    .filter((attachment) => Boolean(attachment.referenceUrl))
    .map((attachment) => {
      if (isAttachmentImage(attachment)) {
        return {
          type: 'image_url',
          image_url: { url: attachment.referenceUrl ?? '' },
        }
      }

      return {
        type: 'file',
        file: {
          filename: attachment.filename,
          file_data: attachment.referenceUrl ?? '',
        },
      }
    })
}

export function getAttachmentStatusLabel(
  status: PlaygroundAttachment['status'],
  currentModel?: string,
  attachment?: PlaygroundAttachment
): string {
  if (currentModel && attachment) {
    const compatibility = getAttachmentCompatibility(attachment, currentModel)
    if (!compatibility.available) {
      return 'Unavailable'
    }
  }

  switch (status) {
    case 'uploading':
      return 'Uploading'
    case 'ready':
    case 'active':
      return 'Ready'
    case 'failed':
      return 'Failed'
    case 'deleted':
      return 'Deleted'
    case 'expired':
      return 'Expired'
    default:
      return 'Unavailable'
  }
}

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

type AttachmentCompatibility = {
  available: boolean
  reason?: string
}

const IMAGE_ONLY_ATTACHMENT_MODELS = ['ollama', 'llama', 'qwen-vl-ocr']
const FILE_FRIENDLY_ATTACHMENT_MODELS = [
  'gpt',
  'o1',
  'o3',
  'o4',
  'gemini',
  'claude',
]

function normalizeModelName(model: string): string {
  return model.trim().toLowerCase()
}

function isImageAttachment(attachment: PlaygroundAttachment): boolean {
  return attachment.mimeType.startsWith('image/')
}

function isImageOnlyAttachmentModel(model: string): boolean {
  return IMAGE_ONLY_ATTACHMENT_MODELS.some((prefix) => model.includes(prefix))
}

function supportsDocumentAttachments(model: string): boolean {
  return FILE_FRIENDLY_ATTACHMENT_MODELS.some((prefix) => model.startsWith(prefix))
}

export function getAttachmentCompatibility(
  attachment: PlaygroundAttachment,
  model: string
): AttachmentCompatibility {
  const normalizedModel = normalizeModelName(model)

  if (!normalizedModel || isImageAttachment(attachment)) {
    return { available: true }
  }

  if (isImageOnlyAttachmentModel(normalizedModel)) {
    return {
      available: false,
      reason: 'Selected model only supports image attachments',
    }
  }

  if (supportsDocumentAttachments(normalizedModel)) {
    return { available: true }
  }

  return {
    available: false,
    reason: 'Selected model may not support this attachment type',
  }
}

export function hasModelIncompatibleAttachments(
  attachments: PlaygroundAttachment[],
  model: string
): boolean {
  return attachments.some((attachment) => {
    if (attachment.status !== 'ready' && attachment.status !== 'active') {
      return false
    }

    return !getAttachmentCompatibility(attachment, model).available
  })
}

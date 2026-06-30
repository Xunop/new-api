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
import { nanoid } from 'nanoid'

import { MESSAGE_ROLES, MESSAGE_STATUS } from '../../constants'
import type {
  Message,
  MessageVersion,
  ChatCompletionMessage,
  ContentPart,
  PlaygroundAttachment,
} from '../../types'
import {
  buildAttachmentContentParts,
} from '../attachments/playground-attachment-utils'
import {
  getAttachmentCompatibility,
} from '../attachments/attachment-compatibility-utils'

/**
 * Create a new message version
 */
export function createMessageVersion(content: string): MessageVersion {
  return {
    id: nanoid(),
    content,
  }
}

/**
 * Get current version from message (always returns the first version)
 */
export function getCurrentVersion(message: Message): MessageVersion {
  return message.versions[0] || { id: 'default', content: '' }
}

/**
 * Get displayable content from the current message version.
 */
export function getMessageContent(message: Message): string {
  return getCurrentVersion(message).content
}

/**
 * Check whether a message has non-empty content in its current version.
 */
export function hasMessageContent(message: Message): boolean {
  return getMessageContent(message).trim() !== ''
}

/**
 * Update current version content in message
 */
export function updateCurrentVersionContent(
  message: Message,
  content: string
): Message {
  const currentVersion = getCurrentVersion(message)
  return {
    ...message,
    versions: [{ ...currentVersion, content }],
  }
}

/**
 * Create a user message
 */
export function createUserMessage(
  content: string,
  attachments: PlaygroundAttachment[] = [],
  createdAt: number = Date.now()
): Message {
  return {
    key: nanoid(),
    from: MESSAGE_ROLES.USER,
    versions: [createMessageVersion(content)],
    attachments: attachments.length > 0 ? attachments : undefined,
    createdAt,
  }
}

/**
 * Create a loading assistant message
 */
export function createLoadingAssistantMessage(
  startedAt: number = Date.now()
): Message {
  return {
    key: nanoid(),
    from: MESSAGE_ROLES.ASSISTANT,
    versions: [createMessageVersion('')],
    createdAt: startedAt,
    startedAt,
    reasoning: undefined,
    isReasoningComplete: false,
    isContentComplete: false,
    isReasoningStreaming: false,
    status: MESSAGE_STATUS.LOADING,
  }
}

/**
 * Build message content with optional images
 */
export function buildMessageContent(
  text: string,
  attachments: PlaygroundAttachment[] = [],
  currentModel?: string
): string | ContentPart[] {
  if (currentModel) {
    const incompatibleAttachment = attachments.find(
      (attachment) =>
        isAttachmentReadyForModelValidation(attachment) &&
        !getAttachmentCompatibility(attachment, currentModel).available
    )

    if (incompatibleAttachment) {
      throw new Error(
        getAttachmentCompatibility(incompatibleAttachment, currentModel).reason ??
          'Selected model may not support this attachment type'
      )
    }
  }

  const attachmentParts = buildAttachmentContentParts(attachments, currentModel)

  if (attachmentParts.length === 0) {
    return text
  }

  const parts: ContentPart[] = [
    {
      type: 'text',
      text: text || '',
    },
    ...attachmentParts,
  ]

  return parts
}

function isAttachmentReadyForModelValidation(
  attachment: PlaygroundAttachment
): boolean {
  return attachment.status === 'ready' || attachment.status === 'active'
}

/**
 * Extract text content from message content
 */
export function getTextContent(content: string | ContentPart[]): string {
  if (typeof content === 'string') {
    return content
  }

  if (Array.isArray(content)) {
    const textPart = content.find((part) => part.type === 'text')
    return textPart?.text || ''
  }

  return ''
}

/**
 * Format message for API request
 */
export function formatMessageForAPI(
  message: Message,
  currentModel?: string
): ChatCompletionMessage {
  const currentVersion = getCurrentVersion(message)
  return {
    role: message.from,
    content: buildMessageContent(
      currentVersion.content,
      message.attachments,
      currentModel
    ),
  }
}

/**
 * Check if message is valid for API request
 * Excludes loading/streaming assistant messages and empty content
 */
export function isValidMessage(message: Message): boolean {
  if (!message || !message.from || !message.versions.length) return false

  // Exclude empty assistant messages (loading/streaming placeholders)
  if (message.from === MESSAGE_ROLES.ASSISTANT && !hasMessageContent(message)) {
    return false
  }

  return true
}

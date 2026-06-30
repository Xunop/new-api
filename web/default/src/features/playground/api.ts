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
import { api } from '@/lib/api'

import { API_ENDPOINTS } from './constants'
import type {
  ChatCompletionRequest,
  ChatCompletionResponse,
  ModelOption,
  GroupOption,
  PlaygroundAttachment,
} from './types'

type APIEnvelope<T> = {
  success: boolean
  message?: string
  code?: string
  data: T
}

type RawPlaygroundAttachment = {
  id: string
  session_id?: string
  filename: string
  mime_type: string
  size: number
  status: PlaygroundAttachment['status']
  expires_at?: number
}

type RawPlaygroundAttachmentReference = RawPlaygroundAttachment & {
  url: string
}

function mapAttachment(raw: RawPlaygroundAttachment): PlaygroundAttachment {
  return {
    id: raw.id,
    sessionId: raw.session_id ?? '',
    filename: raw.filename,
    mimeType: raw.mime_type,
    size: raw.size,
    status: raw.status === 'active' ? 'ready' : raw.status,
    expiresAt: raw.expires_at,
  }
}

function mapAttachmentReference(
  raw: RawPlaygroundAttachmentReference
): PlaygroundAttachment {
  return {
    ...mapAttachment(raw),
    referenceUrl: raw.url,
  }
}

function assertAPIEnvelope<T>(payload: APIEnvelope<T>): T {
  if (!payload.success) {
    const error = new Error(payload.message || payload.code || 'Request failed')
    error.name = payload.code || 'playground_attachment_error'
    throw error
  }
  return payload.data
}

/**
 * Send chat completion request (non-streaming)
 */
export async function sendChatCompletion(
  payload: ChatCompletionRequest,
  signal?: AbortSignal
): Promise<ChatCompletionResponse> {
  const res = await api.post(API_ENDPOINTS.CHAT_COMPLETIONS, payload, {
    signal,
    skipErrorHandler: true,
  })
  return res.data
}

/**
 * Get user available models
 */
export async function getUserModels(group: string): Promise<ModelOption[]> {
  const res = await api.get(API_ENDPOINTS.USER_MODELS, {
    params: { group },
  })
  const { data } = res

  if (!data.success || !Array.isArray(data.data)) {
    return []
  }

  return data.data.map((model: string) => ({
    label: model,
    value: model,
  }))
}

/**
 * Get user groups
 */
export async function getUserGroups(): Promise<GroupOption[]> {
  const res = await api.get(API_ENDPOINTS.USER_GROUPS)
  const { data } = res

  if (!data.success || !data.data) {
    return []
  }

  const groupData = data.data as Record<string, { desc: string; ratio: number }>

  // label is for button display (name only); desc is for dropdown content
  return Object.entries(groupData).map(([group, info]) => ({
    label: group,
    value: group,
    ratio: info.ratio,
    desc: info.desc,
  }))
}

export async function uploadPlaygroundAttachment(
  sessionId: string,
  file: File
): Promise<PlaygroundAttachment> {
  const formData = new FormData()
  formData.append('session_id', sessionId)
  formData.append('file', file)

  const res = await api.post<APIEnvelope<RawPlaygroundAttachment>>(
    API_ENDPOINTS.ATTACHMENTS,
    formData,
    {
      skipErrorHandler: true,
    }
  )
  return mapAttachment(assertAPIEnvelope(res.data))
}

export async function listPlaygroundAttachments(
  sessionId: string
): Promise<PlaygroundAttachment[]> {
  const res = await api.get<APIEnvelope<RawPlaygroundAttachment[]>>(
    API_ENDPOINTS.ATTACHMENTS,
    {
      params: { session_id: sessionId },
      skipErrorHandler: true,
    }
  )
  return assertAPIEnvelope(res.data).map(mapAttachment)
}

export async function deletePlaygroundAttachment(id: string): Promise<void> {
  const res = await api.delete<APIEnvelope<null>>(
    `${API_ENDPOINTS.ATTACHMENTS}/${id}`,
    {
      skipErrorHandler: true,
    }
  )
  assertAPIEnvelope(res.data)
}

export async function generatePlaygroundAttachmentReferences(
  attachmentIds: string[]
): Promise<PlaygroundAttachment[]> {
  const res = await api.post<APIEnvelope<RawPlaygroundAttachmentReference[]>>(
    API_ENDPOINTS.ATTACHMENT_REFERENCES,
    { attachment_ids: attachmentIds },
    {
      skipErrorHandler: true,
    }
  )
  return assertAPIEnvelope(res.data).map(mapAttachmentReference)
}

export async function deletePlaygroundSessionAttachments(
  sessionId: string
): Promise<void> {
  const attachments = await listPlaygroundAttachments(sessionId)
  await Promise.allSettled(
    attachments
      .filter((attachment) => attachment.status === 'ready')
      .map((attachment) => deletePlaygroundAttachment(attachment.id))
  )
}

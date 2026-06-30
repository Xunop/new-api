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
import { useCallback, useEffect, useState } from 'react'

import {
  deletePlaygroundAttachment,
  generatePlaygroundAttachmentReferences,
  listPlaygroundAttachments,
  uploadPlaygroundAttachment,
} from '../api'
import {
  createPendingAttachment,
  failPendingAttachment,
  getAttachmentCompatibility,
  getAttachmentErrorLabel,
  isAttachmentReady,
  mergeListedSessionAttachments,
  replacePendingAttachment,
  removePendingAttachment,
} from '../lib'
import type { PlaygroundAttachment } from '../types'

function getErrorMessage(error: unknown): string {
  return getAttachmentErrorLabel(error)
}

export function usePlaygroundAttachments(sessionId: string, currentModel: string) {
  const [attachments, setAttachments] = useState<PlaygroundAttachment[]>([])

  useEffect(() => {
    let cancelled = false

    setAttachments((current) =>
      mergeListedSessionAttachments(current, [], sessionId)
    )

    void listPlaygroundAttachments(sessionId)
      .then((listed) => {
        if (cancelled) {
          return
        }

        setAttachments((current) =>
          mergeListedSessionAttachments(current, listed, sessionId)
        )
      })
      .catch(() => undefined)

    return () => {
      cancelled = true
    }
  }, [sessionId])

  const uploadFiles = useCallback(
    (fileList: FileList | File[]) => {
      const files = Array.from(fileList)
      if (files.length === 0) return

      const uploadingAttachments = files.map((file) =>
        createPendingAttachment(sessionId, {
          id: `pending_${nanoid()}`,
          name: file.name,
          size: file.size,
          type: file.type,
        })
      )
      setAttachments((current) => [...current, ...uploadingAttachments])

      for (let index = 0; index < files.length; index++) {
        const file = files[index]
        const localID = uploadingAttachments[index].id

        void uploadPlaygroundAttachment(sessionId, file)
          .then((uploaded) => {
            setAttachments((current) =>
              replacePendingAttachment(current, localID, uploaded)
            )
          })
          .catch((error: unknown) => {
            setAttachments((current) =>
              failPendingAttachment(current, localID, getErrorMessage(error))
            )
          })
      }
    },
    [sessionId]
  )

  const removeAttachment = useCallback((attachment: PlaygroundAttachment) => {
    setAttachments((current) => removePendingAttachment(current, attachment.id))
    if (attachment.id.startsWith('att_')) {
      void deletePlaygroundAttachment(attachment.id).catch(() => undefined)
    }
  }, [])

  const clearAttachments = useCallback(() => {
    setAttachments([])
  }, [])

  const clearSubmittedAttachments = useCallback((submittedIds: string[]) => {
    const submitted = new Set(submittedIds)
    setAttachments((current) =>
      current.filter((attachment) => !submitted.has(attachment.id))
    )
  }, [])

  const resolveReadyReferences = useCallback(async () => {
    const readyAttachments = attachments.filter(isAttachmentReady)
    if (readyAttachments.length === 0) {
      return []
    }

    const incompatibleAttachment = readyAttachments.find(
      (attachment) =>
        !getAttachmentCompatibility(attachment, currentModel).available
    )
    if (incompatibleAttachment) {
      throw new Error(
        getAttachmentCompatibility(incompatibleAttachment, currentModel).reason ??
          'Selected model may not support this attachment type'
      )
    }

    const references = await generatePlaygroundAttachmentReferences(
      readyAttachments.map((attachment) => attachment.id)
    )
    const referencesByID = new Map(
      references.map((attachment) => [attachment.id, attachment])
    )

    return readyAttachments.map(
      (attachment) => referencesByID.get(attachment.id) ?? attachment
    )
  }, [attachments, currentModel])

  const hasReadyAttachments = attachments.some(
    (attachment) =>
      isAttachmentReady(attachment) &&
      getAttachmentCompatibility(attachment, currentModel).available
  )
  const hasIncompatibleReadyAttachments = attachments.some(
    (attachment) =>
      isAttachmentReady(attachment) &&
      !getAttachmentCompatibility(attachment, currentModel).available
  )
  const hasPendingUploads = attachments.some(
    (attachment) => attachment.status === 'uploading'
  )

  return {
    attachments,
    hasIncompatibleReadyAttachments,
    hasPendingUploads,
    hasReadyAttachments,
    uploadFiles,
    removeAttachment,
    clearAttachments,
    clearSubmittedAttachments,
    resolveReadyReferences,
  }
}

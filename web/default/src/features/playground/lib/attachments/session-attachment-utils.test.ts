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
import { describe, expect, test } from 'bun:test'

import type { PlaygroundAttachment } from '../../types'
import {
  createPendingAttachment,
  failPendingAttachment,
  mergeListedSessionAttachments,
  replacePendingAttachment,
} from './session-attachment-utils'

describe('session attachment state', () => {
  test('creates an uploading attachment chip from a selected file', () => {
    expect(
      createPendingAttachment('session-a', {
        id: 'pending_1',
        name: 'draft.txt',
        size: 32,
        type: 'text/plain',
      })
    ).toEqual({
      id: 'pending_1',
      sessionId: 'session-a',
      filename: 'draft.txt',
      mimeType: 'text/plain',
      size: 32,
      status: 'uploading',
    })
  })

  test('replaces a pending chip with the uploaded attachment metadata', () => {
    const current: PlaygroundAttachment[] = [
      {
        id: 'pending_1',
        sessionId: 'session-a',
        filename: 'draft.txt',
        mimeType: 'text/plain',
        size: 32,
        status: 'uploading',
      },
    ]

    const uploaded: PlaygroundAttachment = {
      id: 'att_uploaded',
      sessionId: 'session-a',
      filename: 'draft.txt',
      mimeType: 'text/plain',
      size: 32,
      status: 'active',
    }

    expect(replacePendingAttachment(current, 'pending_1', uploaded)).toEqual([
      {
        ...uploaded,
        status: 'ready',
      },
    ])
  })

  test('keeps a failed upload visible with a useful error', () => {
    const current: PlaygroundAttachment[] = [
      {
        id: 'pending_1',
        sessionId: 'session-a',
        filename: 'draft.txt',
        mimeType: 'text/plain',
        size: 32,
        status: 'uploading',
      },
    ]

    expect(
      failPendingAttachment(current, 'pending_1', 'Attachment upload failed')
    ).toEqual([
      {
        id: 'pending_1',
        sessionId: 'session-a',
        filename: 'draft.txt',
        mimeType: 'text/plain',
        size: 32,
        status: 'failed',
        error: 'Attachment upload failed',
      },
    ])
  })

  test('rehydrates listed session attachments and keeps local transient entries', () => {
    const current: PlaygroundAttachment[] = [
      {
        id: 'att_ready_local',
        sessionId: 'session-a',
        filename: 'existing.pdf',
        mimeType: 'application/pdf',
        size: 512,
        status: 'ready',
      },
      {
        id: 'pending_uploading',
        sessionId: 'session-a',
        filename: 'draft.txt',
        mimeType: 'text/plain',
        size: 12,
        status: 'uploading',
      },
      {
        id: 'pending_failed',
        sessionId: 'session-a',
        filename: 'broken.txt',
        mimeType: 'text/plain',
        size: 8,
        status: 'failed',
        error: 'Attachment upload failed',
      },
      {
        id: 'att_other_session',
        sessionId: 'session-b',
        filename: 'stale.txt',
        mimeType: 'text/plain',
        size: 4,
        status: 'ready',
      },
    ]

    const listed: PlaygroundAttachment[] = [
      {
        id: 'att_ready_server',
        sessionId: 'session-a',
        filename: 'server.pdf',
        mimeType: 'application/pdf',
        size: 1024,
        status: 'ready',
      },
    ]

    expect(
      mergeListedSessionAttachments(current, listed, 'session-a')
    ).toEqual([
      listed[0],
      current[1],
      current[2],
    ])
  })

  test('clears attachments from other sessions when no server attachments are listed', () => {
    const current: PlaygroundAttachment[] = [
      {
        id: 'att_previous',
        sessionId: 'session-a',
        filename: 'old.txt',
        mimeType: 'text/plain',
        size: 16,
        status: 'ready',
      },
      {
        id: 'pending_next',
        sessionId: 'session-b',
        filename: 'next.txt',
        mimeType: 'text/plain',
        size: 16,
        status: 'uploading',
      },
    ]

    expect(
      mergeListedSessionAttachments(current, [], 'session-b')
    ).toEqual([current[1]])
  })
})

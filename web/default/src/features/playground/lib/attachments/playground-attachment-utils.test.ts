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

import { removePendingAttachment } from './playground-attachment-utils'
import type { PlaygroundAttachment } from '../../types'

const readyImageAttachment: PlaygroundAttachment = {
  id: 'att_image',
  sessionId: 'session-a',
  filename: 'diagram.png',
  mimeType: 'image/png',
  size: 128,
  status: 'ready',
  referenceUrl: 'https://gateway.example/diagram.png',
}

const readyFileAttachment: PlaygroundAttachment = {
  id: 'att_file',
  sessionId: 'session-a',
  filename: 'notes.txt',
  mimeType: 'text/plain',
  size: 64,
  status: 'ready',
  referenceUrl: 'https://gateway.example/notes.txt',
}

describe('playground attachment payloads', () => {
  test('removing a selected attachment excludes it from submission', () => {
    const attachments = removePendingAttachment(
      [readyImageAttachment, readyFileAttachment],
      'att_file'
    )

    expect(attachments.map((attachment) => attachment.id)).toEqual(['att_image'])
  })
})

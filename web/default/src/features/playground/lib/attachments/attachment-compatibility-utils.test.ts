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

import {
  getAttachmentCompatibility,
  hasModelIncompatibleAttachments,
} from './attachment-compatibility-utils'

const readyImageAttachment = {
  id: 'att_image',
  sessionId: 'session-a',
  filename: 'diagram.png',
  mimeType: 'image/png',
  size: 128,
  status: 'ready' as const,
  referenceUrl: 'https://gateway.example/diagram.png',
}

const readyFileAttachment = {
  id: 'att_file',
  sessionId: 'session-a',
  filename: 'notes.txt',
  mimeType: 'text/plain',
  size: 64,
  status: 'ready' as const,
  referenceUrl: 'https://gateway.example/notes.txt',
}

describe('attachment compatibility utils', () => {
  test('marks text attachments unavailable for ollama-style models', () => {
    expect(
      getAttachmentCompatibility(readyFileAttachment, 'llama3.2-vision')
    ).toEqual({
      available: false,
      reason: 'Selected model only supports image attachments',
    })

    expect(getAttachmentCompatibility(readyImageAttachment, 'llama3.2-vision'))
      .toEqual({
        available: true,
      })
  })

  test('marks text attachments available for gpt and gemini models', () => {
    expect(getAttachmentCompatibility(readyFileAttachment, 'gpt-4o')).toEqual({
      available: true,
    })
    expect(
      getAttachmentCompatibility(readyFileAttachment, 'gemini-2.5-flash')
    ).toEqual({
      available: true,
    })
  })

  test('detects when a ready attachment is incompatible with the selected model', () => {
    expect(
      hasModelIncompatibleAttachments(
        [readyImageAttachment, readyFileAttachment],
        'llama3.2-vision'
      )
    ).toBe(true)

    expect(
      hasModelIncompatibleAttachments(
        [readyImageAttachment, readyFileAttachment],
        'gpt-4o'
      )
    ).toBe(false)
  })
})

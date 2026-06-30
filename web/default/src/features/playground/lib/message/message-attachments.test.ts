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
import { beforeAll, describe, expect, test, mock } from 'bun:test'

import { MESSAGE_ROLES } from '../../constants'
import type { PlaygroundAttachment } from '../../types'

const readyAttachment: PlaygroundAttachment = {
  id: 'att_file',
  sessionId: 'session-a',
  filename: 'notes.txt',
  mimeType: 'text/plain',
  size: 64,
  status: 'ready',
  referenceUrl: 'https://gateway.example/notes.txt',
}

let appendUserMessagePair: typeof import('./conversation-message-utils').appendUserMessagePair
let getMessageContentState: typeof import('./message-content-utils').getMessageContentState

beforeAll(async () => {
  mock.module('nanoid', () => ({
    nanoid: () => 'fixed-nanoid',
  }))

  ;({ appendUserMessagePair } = await import('./conversation-message-utils'))
  ;({ getMessageContentState } = await import('./message-content-utils'))
})

describe('playground message attachments', () => {
  test('sent user messages preserve attachment metadata', () => {
    const messages = appendUserMessagePair([], 'review this', [readyAttachment])

    expect(messages).toHaveLength(2)
    expect(messages[0].from).toBe(MESSAGE_ROLES.USER)
    expect(messages[0].attachments).toEqual([readyAttachment])
  })

  test('attachment-only user messages still render as message content', () => {
    const messages = appendUserMessagePair([], '', [readyAttachment])
    const userMessage = messages[0]

    expect(
      getMessageContentState(userMessage, userMessage.versions[0]?.content ?? '')
        .showMessageContent
    ).toBe(true)
  })
})

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
import { beforeEach, describe, expect, test } from 'bun:test'

import { MESSAGE_ROLES } from '../../constants'
import type { Message } from '../../types'
import { loadMessages, saveMessages } from './storage'

function installLocalStorage() {
  const values = new Map<string, string>()

  Object.defineProperty(globalThis, 'localStorage', {
    configurable: true,
    value: {
      get length() {
        return values.size
      },
      clear() {
        values.clear()
      },
      getItem(key: string) {
        return values.get(key) ?? null
      },
      key(index: number) {
        return Array.from(values.keys())[index] ?? null
      },
      removeItem(key: string) {
        values.delete(key)
      },
      setItem(key: string, value: string) {
        values.set(key, value)
      },
    } satisfies Storage,
  })
}

describe('playground attachment storage', () => {
  beforeEach(() => {
    installLocalStorage()
  })

  test('loads expired attachment metadata as expired', () => {
    const messages: Message[] = [
      {
        key: 'message-a',
        from: MESSAGE_ROLES.USER,
        versions: [{ id: 'version-a', content: 'review this' }],
        attachments: [
          {
            id: 'att_old',
            sessionId: 'session-a',
            filename: 'old.txt',
            mimeType: 'text/plain',
            size: 12,
            status: 'ready',
            expiresAt: 1,
          },
        ],
      },
    ]

    saveMessages(messages)

    expect(loadMessages()?.[0].attachments?.[0].status).toBe('expired')
  })
})

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
import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { toast } from 'sonner'

import { deletePlaygroundSessionAttachments } from './api'
import { PlaygroundChat } from './components/chat/playground-chat'
import { PlaygroundInput } from './components/input/playground-input'
import {
  useChatHandler,
  usePlaygroundAttachments,
  usePlaygroundConversation,
  usePlaygroundOptions,
  usePlaygroundState,
} from './hooks'
import { getAttachmentErrorLabel } from './lib'

export function Playground() {
  const {
    config,
    parameterEnabled,
    messages,
    sessionId,
    isLoadingMessages,
    models,
    groups,
    updateMessages,
    setModels,
    setGroups,
    updateConfig,
    clearMessages,
    resetSession,
  } = usePlaygroundState()
  const { t } = useTranslation()
  const [isPreparingAttachments, setIsPreparingAttachments] = useState(false)

  const { sendChat, stopGeneration, isGenerating } = useChatHandler({
    config,
    parameterEnabled,
    onMessageUpdate: updateMessages,
  })

  const {
    editingMessageKey,
    handleSendMessage,
    handleRegenerateMessage,
    handleEditMessage,
    handleEditOpenChange,
    applyEdit,
    handleDeleteMessage,
  } = usePlaygroundConversation({
    messages,
    updateMessages,
    sendChat,
  })

  const {
    attachments,
    hasPendingUploads,
    hasReadyAttachments,
    uploadFiles,
    removeAttachment,
    clearAttachments,
    clearSubmittedAttachments,
    resolveReadyReferences,
  } = usePlaygroundAttachments(sessionId, config.model)

  const handleSendMessageWithAttachments = async (text: string) => {
    try {
      setIsPreparingAttachments(true)
      const referencedAttachments = await resolveReadyReferences()
      handleSendMessage(text, referencedAttachments)
      clearSubmittedAttachments(
        referencedAttachments.map((attachment) => attachment.id)
      )
    } catch (error: unknown) {
      const description = t(
        getAttachmentErrorLabel(error, 'Attachment reference failed')
      )
      toast.error(t('Attachment reference failed'), { description })
    } finally {
      setIsPreparingAttachments(false)
    }
  }

  const handleClearMessages = () => {
    handleEditOpenChange(false)
    void deletePlaygroundSessionAttachments(sessionId).catch(() => undefined)
    clearAttachments()
    resetSession()
    clearMessages()
  }

  const { isLoadingModels } = usePlaygroundOptions({
    currentGroup: config.group,
    currentModel: config.model,
    setGroups,
    setModels,
    updateConfig,
  })

  return (
    <div className='relative flex size-full min-h-0 flex-col overflow-hidden'>
      {/* Full-width scroll container: scrolling works even over side whitespace */}
      <div className='flex min-h-0 flex-1 flex-col overflow-hidden'>
        <PlaygroundChat
          messages={messages}
          isLoadingMessages={isLoadingMessages}
          onRegenerateMessage={handleRegenerateMessage}
          onEditMessage={handleEditMessage}
          onDeleteMessage={handleDeleteMessage}
          onSelectPrompt={handleSendMessage}
          isGenerating={isGenerating}
          editingKey={editingMessageKey}
          onCancelEdit={handleEditOpenChange}
          onSaveEdit={(newContent) => applyEdit(newContent, false)}
          onSaveEditAndSubmit={(newContent) => applyEdit(newContent, true)}
        />
      </div>

      {/* Input area: center content and constrain to the same container width */}
      <div className='mx-auto w-full max-w-4xl'>
        <PlaygroundInput
          attachments={attachments}
          disabled={isGenerating || isPreparingAttachments || hasPendingUploads}
          currentModel={config.model}
          groups={groups}
          groupValue={config.group}
          isGenerating={isGenerating}
          isModelLoading={isLoadingModels}
          modelValue={config.model}
          models={models}
          onGroupChange={(value) => updateConfig('group', value)}
          onClearMessages={handleClearMessages}
          onFilesSelected={uploadFiles}
          onModelChange={(value) => updateConfig('model', value)}
          onRemoveAttachment={removeAttachment}
          onStop={stopGeneration}
          onSubmit={handleSendMessageWithAttachments}
          hasReadyAttachments={hasReadyAttachments}
          hasMessages={messages.length > 0}
        />
      </div>
    </div>
  )
}

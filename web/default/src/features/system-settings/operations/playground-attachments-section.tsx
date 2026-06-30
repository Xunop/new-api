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
import { useEffect, useMemo, useRef } from 'react'
import * as z from 'zod'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { useTranslation } from 'react-i18next'
import { toast } from 'sonner'

import {
  Form,
  FormControl,
  FormDescription,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from '@/components/ui/form'
import { Input } from '@/components/ui/input'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Switch } from '@/components/ui/switch'
import { Textarea } from '@/components/ui/textarea'

import {
  SettingsForm,
  SettingsSwitchContent,
  SettingsSwitchItem,
} from '../components/settings-form-layout'
import { SettingsPageFormActions } from '../components/settings-page-context'
import { SettingsSection } from '../components/settings-section'
import { useUpdateOption } from '../hooks/use-update-option'

const sectionSchema = z.object({
  playground_attachment: z.object({
    enabled: z.boolean(),
    storage_driver: z.enum(['local', 'oss']),
    ttl_hours: z.coerce.number().int().min(1),
    max_file_size_bytes: z.coerce.number().int().min(1),
    max_files_per_message: z.coerce.number().int().min(1),
    max_files_per_session: z.coerce.number().int().min(1),
    allowed_mime_types: z.string(),
    reference_ttl_seconds: z.coerce.number().int().min(1),
    local_base_path: z.string(),
    cleanup_interval_minutes: z.coerce.number().int().min(1),
    cleanup_batch_size: z.coerce.number().int().min(1),
    oss_endpoint: z.string(),
    oss_bucket: z.string(),
    oss_region: z.string(),
    oss_api_key: z.string(),
    oss_secret: z.string(),
    oss_object_prefix: z.string(),
  }),
})

type PlaygroundAttachmentsFormValues = z.infer<typeof sectionSchema>
type PlaygroundAttachmentsFormInput = z.input<typeof sectionSchema>

type PlaygroundAttachmentsSectionProps = {
  defaultValues: {
    'playground_attachment.enabled': boolean
    'playground_attachment.storage_driver': string
    'playground_attachment.ttl_hours': number
    'playground_attachment.max_file_size_bytes': number
    'playground_attachment.max_files_per_message': number
    'playground_attachment.max_files_per_session': number
    'playground_attachment.allowed_mime_types': string
    'playground_attachment.reference_ttl_seconds': number
    'playground_attachment.local_base_path': string
    'playground_attachment.cleanup_interval_minutes': number
    'playground_attachment.cleanup_batch_size': number
    'playground_attachment.oss_endpoint': string
    'playground_attachment.oss_bucket': string
    'playground_attachment.oss_region': string
    'playground_attachment.oss_api_key': string
    'playground_attachment.oss_secret': string
    'playground_attachment.oss_object_prefix': string
  }
}

type NormalizedValues = PlaygroundAttachmentsSectionProps['defaultValues']

function normalizeMIMEValue(value: string): string {
  try {
    const parsed = JSON.parse(value)
    if (Array.isArray(parsed)) {
      return parsed.join('\n')
    }
  } catch {
    // keep raw text fallback
  }

  return value
}

function parseAllowedMIMETypes(value: string): string {
  return JSON.stringify(
    value
      .split('\n')
      .map((entry) => entry.trim())
      .filter(Boolean)
  )
}

function buildFormDefaults(
  defaults: PlaygroundAttachmentsSectionProps['defaultValues']
): PlaygroundAttachmentsFormInput {
  return {
    playground_attachment: {
      enabled: defaults['playground_attachment.enabled'],
      storage_driver:
        defaults['playground_attachment.storage_driver'] === 'oss'
          ? 'oss'
          : 'local',
      ttl_hours: defaults['playground_attachment.ttl_hours'],
      max_file_size_bytes: defaults['playground_attachment.max_file_size_bytes'],
      max_files_per_message:
        defaults['playground_attachment.max_files_per_message'],
      max_files_per_session:
        defaults['playground_attachment.max_files_per_session'],
      allowed_mime_types: normalizeMIMEValue(
        defaults['playground_attachment.allowed_mime_types']
      ),
      reference_ttl_seconds:
        defaults['playground_attachment.reference_ttl_seconds'],
      local_base_path: defaults['playground_attachment.local_base_path'],
      cleanup_interval_minutes:
        defaults['playground_attachment.cleanup_interval_minutes'],
      cleanup_batch_size:
        defaults['playground_attachment.cleanup_batch_size'],
      oss_endpoint: defaults['playground_attachment.oss_endpoint'],
      oss_bucket: defaults['playground_attachment.oss_bucket'],
      oss_region: defaults['playground_attachment.oss_region'],
      oss_api_key: '',
      oss_secret: '',
      oss_object_prefix: defaults['playground_attachment.oss_object_prefix'],
    },
  }
}

function normalizeDefaults(
  defaults: PlaygroundAttachmentsSectionProps['defaultValues']
): NormalizedValues {
  return {
    ...defaults,
    'playground_attachment.allowed_mime_types': parseAllowedMIMETypes(
      normalizeMIMEValue(defaults['playground_attachment.allowed_mime_types'])
    ),
  }
}

function normalizeFormValues(
  values: PlaygroundAttachmentsFormValues,
  defaults: PlaygroundAttachmentsSectionProps['defaultValues']
): NormalizedValues {
  return {
    'playground_attachment.enabled': values.playground_attachment.enabled,
    'playground_attachment.storage_driver':
      values.playground_attachment.storage_driver,
    'playground_attachment.ttl_hours': values.playground_attachment.ttl_hours,
    'playground_attachment.max_file_size_bytes':
      values.playground_attachment.max_file_size_bytes,
    'playground_attachment.max_files_per_message':
      values.playground_attachment.max_files_per_message,
    'playground_attachment.max_files_per_session':
      values.playground_attachment.max_files_per_session,
    'playground_attachment.allowed_mime_types': parseAllowedMIMETypes(
      values.playground_attachment.allowed_mime_types
    ),
    'playground_attachment.reference_ttl_seconds':
      values.playground_attachment.reference_ttl_seconds,
    'playground_attachment.local_base_path':
      values.playground_attachment.local_base_path.trim(),
    'playground_attachment.cleanup_interval_minutes':
      values.playground_attachment.cleanup_interval_minutes,
    'playground_attachment.cleanup_batch_size':
      values.playground_attachment.cleanup_batch_size,
    'playground_attachment.oss_endpoint':
      values.playground_attachment.oss_endpoint.trim(),
    'playground_attachment.oss_bucket':
      values.playground_attachment.oss_bucket.trim(),
    'playground_attachment.oss_region':
      values.playground_attachment.oss_region.trim(),
    'playground_attachment.oss_api_key':
      values.playground_attachment.oss_api_key.trim() ||
      defaults['playground_attachment.oss_api_key'],
    'playground_attachment.oss_secret':
      values.playground_attachment.oss_secret.trim() ||
      defaults['playground_attachment.oss_secret'],
    'playground_attachment.oss_object_prefix':
      values.playground_attachment.oss_object_prefix.trim(),
  }
}

function isEqual(a: unknown, b: unknown): boolean {
  if (Array.isArray(a) && Array.isArray(b)) {
    return JSON.stringify(a) === JSON.stringify(b)
  }
  return a === b
}

export function PlaygroundAttachmentsSection({
  defaultValues,
}: PlaygroundAttachmentsSectionProps) {
  const { t } = useTranslation()
  const updateOption = useUpdateOption()
  const baselineRef = useRef<NormalizedValues>(normalizeDefaults(defaultValues))

  const formDefaults = useMemo(
    () => buildFormDefaults(defaultValues),
    [defaultValues]
  )

  const form = useForm<
    PlaygroundAttachmentsFormInput,
    unknown,
    PlaygroundAttachmentsFormValues
  >({
    resolver: zodResolver(sectionSchema),
    defaultValues: formDefaults,
  })

  useEffect(() => {
    baselineRef.current = normalizeDefaults(defaultValues)
    form.reset(buildFormDefaults(defaultValues))
  }, [defaultValues, form])

  const onSubmit = async (data: PlaygroundAttachmentsFormValues) => {
    const normalized = normalizeFormValues(data, defaultValues)
    const updates = (
      Object.keys(normalized) as Array<keyof NormalizedValues>
    ).filter((key) => !isEqual(normalized[key], baselineRef.current[key]))

    if (updates.length === 0) {
      toast.info(t('No changes to save'))
      return
    }

    for (const key of updates) {
      await updateOption.mutateAsync({
        key,
        value: normalized[key] as string | number | boolean,
      })
    }

    baselineRef.current = normalized
    form.setValue('playground_attachment.oss_api_key', '')
    form.setValue('playground_attachment.oss_secret', '')
  }

  const storageDriver = form.watch('playground_attachment.storage_driver')
  const useOSS = storageDriver === 'oss'

  return (
    <SettingsSection title={t('Playground Attachments')}>
      <Form {...form}>
        <SettingsForm onSubmit={form.handleSubmit(onSubmit)} autoComplete='off'>
          <SettingsPageFormActions
            onSave={form.handleSubmit(onSubmit)}
            isSaving={updateOption.isPending}
            saveLabel='Save Playground attachment settings'
          />

          <FormField
            control={form.control}
            name='playground_attachment.enabled'
            render={({ field }) => (
              <SettingsSwitchItem>
                <SettingsSwitchContent>
                  <FormLabel>{t('Enable Playground attachments')}</FormLabel>
                  <FormDescription>
                    {t(
                      'Allow dashboard users to upload temporary session attachments in Playground.'
                    )}
                  </FormDescription>
                </SettingsSwitchContent>
                <FormControl>
                  <Switch
                    checked={field.value}
                    onCheckedChange={field.onChange}
                  />
                </FormControl>
              </SettingsSwitchItem>
            )}
          />

          <FormField
            control={form.control}
            name='playground_attachment.storage_driver'
            render={({ field }) => (
              <FormItem>
                <FormLabel>{t('Storage driver')}</FormLabel>
                <Select onValueChange={field.onChange} value={field.value}>
                  <FormControl>
                    <SelectTrigger>
                      <SelectValue placeholder={t('Select storage driver')} />
                    </SelectTrigger>
                  </FormControl>
                  <SelectContent>
                    <SelectItem value='local'>{t('Local filesystem')}</SelectItem>
                    <SelectItem value='oss'>{t('Alibaba OSS')}</SelectItem>
                  </SelectContent>
                </Select>
                <FormDescription>
                  {t(
                    'Choose where temporary Playground attachment files are stored.'
                  )}
                </FormDescription>
                <FormMessage />
              </FormItem>
            )}
          />

          <div className='grid gap-4 md:grid-cols-2'>
            <FormField
              control={form.control}
              name='playground_attachment.ttl_hours'
              render={({ field }) => (
                <FormItem>
                  <FormLabel>{t('Attachment TTL (hours)')}</FormLabel>
                  <FormControl>
                    <Input type='number' min={1} {...field} />
                  </FormControl>
                  <FormDescription>
                    {t('Attachments expire after this many hours.')}
                  </FormDescription>
                  <FormMessage />
                </FormItem>
              )}
            />
            <FormField
              control={form.control}
              name='playground_attachment.reference_ttl_seconds'
              render={({ field }) => (
                <FormItem>
                  <FormLabel>{t('Reference TTL (seconds)')}</FormLabel>
                  <FormControl>
                    <Input type='number' min={1} {...field} />
                  </FormControl>
                  <FormDescription>
                    {t('Signed read links remain valid for this many seconds.')}
                  </FormDescription>
                  <FormMessage />
                </FormItem>
              )}
            />
            <FormField
              control={form.control}
              name='playground_attachment.max_file_size_bytes'
              render={({ field }) => (
                <FormItem>
                  <FormLabel>{t('Max file size (bytes)')}</FormLabel>
                  <FormControl>
                    <Input type='number' min={1} {...field} />
                  </FormControl>
                  <FormDescription>
                    {t('Maximum allowed size for a single attachment upload.')}
                  </FormDescription>
                  <FormMessage />
                </FormItem>
              )}
            />
            <FormField
              control={form.control}
              name='playground_attachment.max_files_per_message'
              render={({ field }) => (
                <FormItem>
                  <FormLabel>{t('Max files per message')}</FormLabel>
                  <FormControl>
                    <Input type='number' min={1} {...field} />
                  </FormControl>
                  <FormDescription>
                    {t('Maximum number of attachments allowed per send.')}
                  </FormDescription>
                  <FormMessage />
                </FormItem>
              )}
            />
            <FormField
              control={form.control}
              name='playground_attachment.max_files_per_session'
              render={({ field }) => (
                <FormItem>
                  <FormLabel>{t('Max files per session')}</FormLabel>
                  <FormControl>
                    <Input type='number' min={1} {...field} />
                  </FormControl>
                  <FormDescription>
                    {t(
                      'Maximum number of active attachments kept for one Playground session.'
                    )}
                  </FormDescription>
                  <FormMessage />
                </FormItem>
              )}
            />
            <FormField
              control={form.control}
              name='playground_attachment.cleanup_interval_minutes'
              render={({ field }) => (
                <FormItem>
                  <FormLabel>{t('Cleanup interval (minutes)')}</FormLabel>
                  <FormControl>
                    <Input type='number' min={1} {...field} />
                  </FormControl>
                  <FormDescription>
                    {t('Expired attachments are scanned on this schedule.')}
                  </FormDescription>
                  <FormMessage />
                </FormItem>
              )}
            />
            <FormField
              control={form.control}
              name='playground_attachment.cleanup_batch_size'
              render={({ field }) => (
                <FormItem>
                  <FormLabel>{t('Cleanup batch size')}</FormLabel>
                  <FormControl>
                    <Input type='number' min={1} {...field} />
                  </FormControl>
                  <FormDescription>
                    {t('Number of expired attachments processed per cleanup run.')}
                  </FormDescription>
                  <FormMessage />
                </FormItem>
              )}
            />
          </div>

          <FormField
            control={form.control}
            name='playground_attachment.allowed_mime_types'
            render={({ field }) => (
              <FormItem>
                <FormLabel>{t('Allowed MIME types')}</FormLabel>
                <FormControl>
                  <Textarea
                    className='min-h-32 font-mono text-sm'
                    placeholder={t('One MIME type per line')}
                    {...field}
                  />
                </FormControl>
                <FormDescription>
                  {t(
                    'Each line becomes one allowed MIME type for Playground attachment uploads.'
                  )}
                </FormDescription>
                <FormMessage />
              </FormItem>
            )}
          />

          {!useOSS && (
            <FormField
              control={form.control}
              name='playground_attachment.local_base_path'
              render={({ field }) => (
                <FormItem>
                  <FormLabel>{t('Local storage path')}</FormLabel>
                  <FormControl>
                    <Input
                      placeholder='./data/playground-attachments'
                      {...field}
                    />
                  </FormControl>
                  <FormDescription>
                    {t(
                      'Filesystem directory used when the local storage driver is selected.'
                    )}
                  </FormDescription>
                  <FormMessage />
                </FormItem>
              )}
            />
          )}

          {useOSS && (
            <>
              <div className='grid gap-4 md:grid-cols-2'>
                <FormField
                  control={form.control}
                  name='playground_attachment.oss_endpoint'
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>{t('OSS endpoint')}</FormLabel>
                      <FormControl>
                        <Input placeholder='oss-cn-hangzhou.aliyuncs.com' {...field} />
                      </FormControl>
                      <FormMessage />
                    </FormItem>
                  )}
                />
                <FormField
                  control={form.control}
                  name='playground_attachment.oss_bucket'
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>{t('OSS bucket')}</FormLabel>
                      <FormControl>
                        <Input placeholder='example-playground-bucket' {...field} />
                      </FormControl>
                      <FormMessage />
                    </FormItem>
                  )}
                />
                <FormField
                  control={form.control}
                  name='playground_attachment.oss_region'
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>{t('OSS region')}</FormLabel>
                      <FormControl>
                        <Input placeholder='cn-hangzhou' {...field} />
                      </FormControl>
                      <FormMessage />
                    </FormItem>
                  )}
                />
                <FormField
                  control={form.control}
                  name='playground_attachment.oss_object_prefix'
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>{t('OSS object prefix')}</FormLabel>
                      <FormControl>
                        <Input placeholder='playground' {...field} />
                      </FormControl>
                      <FormMessage />
                    </FormItem>
                  )}
                />
              </div>

              <div className='grid gap-4 md:grid-cols-2'>
                <FormField
                  control={form.control}
                  name='playground_attachment.oss_api_key'
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>{t('OSS access key ID')}</FormLabel>
                      <FormControl>
                        <Input
                          type='password'
                          autoComplete='new-password'
                          placeholder={t('Leave blank to keep the existing key')}
                          {...field}
                        />
                      </FormControl>
                      <FormDescription>
                        {t(
                          'Sensitive values are hidden when loaded. Enter a new value only when you want to replace the current credential.'
                        )}
                      </FormDescription>
                      <FormMessage />
                    </FormItem>
                  )}
                />
                <FormField
                  control={form.control}
                  name='playground_attachment.oss_secret'
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>{t('OSS access key secret')}</FormLabel>
                      <FormControl>
                        <Input
                          type='password'
                          autoComplete='new-password'
                          placeholder={t('Leave blank to keep the existing secret')}
                          {...field}
                        />
                      </FormControl>
                      <FormDescription>
                        {t(
                          'Sensitive values are hidden when loaded. Enter a new value only when you want to replace the current credential.'
                        )}
                      </FormDescription>
                      <FormMessage />
                    </FormItem>
                  )}
                />
              </div>
            </>
          )}
        </SettingsForm>
      </Form>
    </SettingsSection>
  )
}

import { useCallback, useState } from 'react'
import { createTask } from '@/services/monitoringApi'
import { parseMonitoringFetchError } from '@/services/monitoringSnapshot'

/**
 * Отправка POST /tasks и колбэк после успеха (например refetch снимка).
 */
export function useCreateTask(onSuccess) {
  const [submitting, setSubmitting] = useState(false)
  const [error, setError] = useState(null)

  const submit = useCallback(
    async (body) => {
      setError(null)
      setSubmitting(true)
      try {
        await createTask(body)
        onSuccess?.()
      } catch (e) {
        setError(parseMonitoringFetchError(e, 'Не удалось создать задачу'))
      } finally {
        setSubmitting(false)
      }
    },
    [onSuccess],
  )

  return { submit, submitting, error, clearError: () => setError(null) }
}

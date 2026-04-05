import { apiClient } from '@/services/apiClient'

/**
 * Список агентов (GET /agents).
 * Ответ: массив объектов с полями id, name, last_heartbeat, created_at.
 */
export async function fetchAgents() {
  const { data } = await apiClient.get('/agents')
  return data
}

/**
 * Список задач (GET /tasks).
 */
export async function fetchTasks() {
  const { data } = await apiClient.get('/tasks')
  return data
}

/**
 * Список результатов выполнения задач (GET /results).
 */
export async function fetchResults() {
  const { data } = await apiClient.get('/results')
  return data
}

/**
 * Создание задачи (POST /tasks).
 * Тело: agent_id, check_type, payload (строка), max_retries.
 */
export async function createTask(body) {
  const { data } = await apiClient.post('/tasks', body)
  return data
}

/**
 * Повторная постановка задачи в очередь (POST /tasks/{id}/retry).
 * Допустимо для задач в статусах done или failed (проверка на сервере).
 */
export async function retryTask(taskId) {
  const { data } = await apiClient.post(`/tasks/${taskId}/retry`)
  return data
}

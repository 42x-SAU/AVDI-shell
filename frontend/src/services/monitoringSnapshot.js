import { fetchAgents, fetchResults, fetchTasks } from '@/services/monitoringApi'

/**
 * Единая загрузка снимка для панели: агенты, задачи, результаты (параллельно).
 * Используется хуком снимка и при необходимости — другими слоями (polling, будущий WS-merge).
 */
export async function fetchMonitoringSnapshot() {
  const [agents, tasks, results] = await Promise.all([
    fetchAgents(),
    fetchTasks(),
    fetchResults(),
  ])
  return {
    agents: Array.isArray(agents) ? agents : [],
    tasks: Array.isArray(tasks) ? tasks : [],
    results: Array.isArray(results) ? results : [],
  }
}

/**
 * Разбор сообщения об ошибке HTTP (axios / Go text) для мониторинга и связанных действий.
 * @param {unknown} e
 * @param {string} [fallbackMessage] — если тело ответа не дало текста
 */
export function parseMonitoringFetchError(e, fallbackMessage = 'Не удалось загрузить данные') {
  const body = e?.response?.data
  const fromServer =
    (typeof body === 'string' && body.trim() ? body : null) ||
    body?.message ||
    (typeof e?.message === 'string' ? e.message : null)
  return fromServer ? String(fromServer) : fallbackMessage
}

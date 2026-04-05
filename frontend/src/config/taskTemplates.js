/**
 * Локальные шаблоны проверок для UI (без отдельного API).
 * Позже источник можно заменить на fetch, сохранив те же поля для адаптера.
 */

/** Базовые поля шаблона для формы создания задачи */
export const TASK_TEMPLATES = [
  {
    id: 'node-basic',
    name: 'Базовая диагностика узла',
    description: 'Имя хоста через встроенную проверку агента (hostname).',
    check_type: 'hostname',
    examplePayload: '',
    maxRetries: 0,
  },
  {
    id: 'network-context',
    name: 'Сбор сетевого контекста',
    description: 'Диагностическая команда из YAML: список интерфейсов и адресов (ip addr).',
    check_type: 'diagnostic',
    examplePayload: JSON.stringify({ command: 'ip-addresses', vars: {} }),
    maxRetries: 0,
  },
  {
    id: 'internal-services',
    name: 'Проверка доступности внутренних сервисов',
    description: 'ICMP-пинг цели; укажите в payload хост или IP нужного сервиса.',
    check_type: 'ping',
    examplePayload: '10.0.0.1',
    maxRetries: 1,
  },
]

/**
 * Значения формы «с нуля» (ручной сценарий).
 */
export function getManualTaskDefaults() {
  return {
    check_type: 'hostname',
    payload: '',
    max_retries: 0,
  }
}

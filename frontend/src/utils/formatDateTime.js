/**
 * Форматирование даты/времени с сервера для таблиц мониторинга.
 * @param {string | null | undefined} iso
 * @returns {string}
 */
export function formatDateTime(iso) {
  if (iso == null || iso === '') return '—'
  const d = new Date(iso)
  if (Number.isNaN(d.getTime())) return '—'
  return new Intl.DateTimeFormat('ru-RU', {
    dateStyle: 'short',
    timeStyle: 'medium',
  }).format(d)
}

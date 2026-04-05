/**
 * Пытается отформатировать строку как JSON для отображения в UI;
 * если не JSON — возвращает исходную строку.
 */
export function formatJsonForDisplay(raw) {
  if (raw == null || raw === '') return '—'
  const s = String(raw)
  try {
    const parsed = JSON.parse(s)
    return JSON.stringify(parsed, null, 2)
  } catch {
    return s
  }
}

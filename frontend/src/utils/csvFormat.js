/**
 * Простое экранирование полей для CSV (RFC-стиль).
 */
export function escapeCsvField(value) {
  if (value == null) return ''
  const s = String(value)
  if (/[",\n\r]/.test(s)) {
    return `"${s.replace(/"/g, '""')}"`
  }
  return s
}

/** Строка из массива ячеек */
export function csvLine(cells) {
  return cells.map(escapeCsvField).join(',')
}

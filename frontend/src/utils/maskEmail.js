/**
 * Маскирует начало локальной части адреса: первая буква, затем звёздочки, затем @ и домен.
 */
export function maskEmailForDisplay(email) {
  const trimmed = (email || '').trim()
  const at = trimmed.indexOf('@')
  if (at <= 0) return '***'
  const local = trimmed.slice(0, at)
  const domain = trimmed.slice(at + 1)
  if (!domain) return '***'
  if (local.length <= 1) {
    return `*@${domain}`
  }
  const first = local[0]
  const restCount = Math.min(local.length - 1, 4)
  const stars = '*'.repeat(restCount)
  return `${first}${stars}@${domain}`
}

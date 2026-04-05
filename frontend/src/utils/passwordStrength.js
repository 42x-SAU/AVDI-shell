/**
 * Оценка надёжности пароля: длина и классы символов (строчные/прописные латиница, цифры, не-буквы).
 * Красный: до 8 символов. Жёлтый: 8–13 или 14+ при одном классе. Зелёный: 14+ и не менее двух классов.
 * @returns {{ level: 'empty' | 'weak' | 'medium' | 'strong', label: string }}
 */
export function getPasswordStrength(password) {
  if (!password || password.length === 0) {
    return { level: 'empty', label: '' }
  }

  const len = password.length
  const hasLower = /[a-z]/.test(password)
  const hasUpper = /[A-Z]/.test(password)
  const hasDigit = /\d/.test(password)
  const hasSpecial = /[^a-zA-Z0-9]/.test(password)
  const variety = [hasLower, hasUpper, hasDigit, hasSpecial].filter(Boolean).length

  if (len < 8) {
    return { level: 'weak', label: 'Ненадежный' }
  }
  if (len < 14) {
    return { level: 'medium', label: 'Нежелательный' }
  }
  if (variety >= 2) {
    return { level: 'strong', label: 'Отличный' }
  }
  return { level: 'medium', label: 'Нежелательный' }
}

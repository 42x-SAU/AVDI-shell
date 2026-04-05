/**
 * Локальная часть (до @): латиница, цифры и типичные символы; длина как в обычных почтовых клиентах.
 */
function isLocalPartValid(local) {
  if (!local || local.length > 64) return false
  return /^(?:[a-zA-Z0-9]|[a-zA-Z0-9][a-zA-Z0-9._+-]*[a-zA-Z0-9])$/.test(local)
}

/**
 * Домен: только латиница, цифры и дефис в метках; зона — только буквы латиницы (.com, .ru, .info).
 */
function isDomainPartValid(domain) {
  if (!domain || domain.length > 253) return false
  if (domain.startsWith('.') || domain.endsWith('.') || domain.includes('..')) return false
  const labels = domain.split('.')
  if (labels.length < 2) return false
  for (const label of labels) {
    if (!label || label.length > 63) return false
    if (!/^[a-zA-Z0-9](?:[a-zA-Z0-9-]*[a-zA-Z0-9])?$/.test(label)) return false
  }
  const tld = labels[labels.length - 1]
  return tld.length >= 2 && /^[a-zA-Z]+$/.test(tld)
}

/**
 * Формат эл. почты: только латиница/ASCII в духе RFC (без кириллицы и пробелов).
 */
function isEmailFormatValid(email) {
  const trimmed = email.trim()
  if (!trimmed.includes('@')) return false
  // Только печатные ASCII-символы (латиница, цифры, @ . _ + - и т.д. — без кириллицы)
  if (/[^\u0020-\u007E]/.test(trimmed)) return false
  if (/\s/.test(trimmed)) return false

  const at = trimmed.indexOf('@')
  if (trimmed.indexOf('@', at + 1) !== -1) return false

  const local = trimmed.slice(0, at)
  const domain = trimmed.slice(at + 1)

  return isLocalPartValid(local) && isDomainPartValid(domain)
}

/** Логин без «@»: только латинские буквы (без цифр и символов). */
function isLoginLettersOnly(id) {
  if (!id || id.length > 64) return false
  return /^[a-zA-Z]+$/.test(id)
}

/**
 * Клиентская проверка полей входа.
 * @returns {{ identifier: string | null, password: string | null }} тексты ошибок или null
 */
export function validateLoginForm(identifier, password) {
  const id = identifier.trim()
  const pwd = password

  let identifierErr = null
  let passwordErr = null

  if (!id) {
    identifierErr = 'Введите логин или эл. почту.'
  } else if (id.includes('@')) {
    if (!isEmailFormatValid(id)) {
      identifierErr =
        'Укажите адрес латиницей в формате эл. почты: имя@домен.зона (например, user@example.com или contact@company.ru).'
    }
  } else if (!isLoginLettersOnly(id)) {
    identifierErr = 'Логин может состоять только из латинских букв без пробелов и цифр.'
  }

  if (!pwd) {
    passwordErr = 'Введите пароль.'
  } else if (pwd.length < 8) {
    passwordErr = 'Пароль должен содержать не менее 8 символов.'
  }

  return { identifier: identifierErr, password: passwordErr }
}

export function hasLoginErrors(errors) {
  return Boolean(errors.identifier || errors.password)
}

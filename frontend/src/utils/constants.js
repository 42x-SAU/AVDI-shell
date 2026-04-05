/**
 * Ключи localStorage и прочие константы приложения.
 */
export const STORAGE_KEYS = {
  THEME: 'avdi-shell-theme',
}

/** Значения темы: пользовательский выбор или «системная». */
export const THEME = {
  LIGHT: 'light',
  DARK: 'dark',
  SYSTEM: 'system',
}

/** Публичные маршруты (база для будущего расширения). */
export const ROUTES = {
  HOME: '/',
  LOGIN: '/login',
  REGISTER: '/register',
  /** Подтверждение почты после регистрации */
  REGISTER_VERIFY_EMAIL: '/register/verify-email',
  /** Правила обработки персональных данных */
  LEGAL_PERSONAL_DATA: '/legal/personal-data',
  /** Политика конфиденциальности */
  LEGAL_PRIVACY: '/legal/privacy',
  /** Рабочая область после входа (заготовка) */
  APP_SHELL: '/app',
}

/**
 * Исходные пути к брендовым файлам (для документации и замены копий в /src/assets).
 * Рабочие импорты в коде — из src/assets (см. README).
 */
export const ASSET_SOURCE_HINTS = {
  LOGO_PNG: 'C:\\Users\\angel\\AVDI-Contents\\avdi-full.png',
  FAVICON_ICO: 'C:\\Users\\angel\\AVDI-Contents\\avdi-mini.ico',
  MADRIGAL_FOOTER_PNG: 'C:\\Users\\angel\\AVDI-Contents\\madrigal.png',
  GITHUB_PNG: 'C:\\Users\\angel\\AVDI-Contents\\github.png',
  TG_SVG: 'C:\\Users\\angel\\AVDI-Contents\\tg.svg',
}

/** Репозиторий проекта на GitHub (нижняя полоса подвала). */
export const FOOTER_GITHUB_URL = 'https://github.com/42x-SAU/AVDI-shell'

/** Сайт IT-консорциума Madrigal (логотип в подвале). */
export const FOOTER_MADRIGAL_URL = 'https://madrigal.ru/'

/** Ссылки на Telegram в нижней полосе подвала */
export const FOOTER_TELEGRAM_LINKS = [
  { label: '@Stasechka1408', href: 'https://t.me/Stasechka1408' },
  { label: '@PotJoke', href: 'https://t.me/PotJoke' },
  { label: '@propastin', href: 'https://t.me/propastin' },
  { label: '@ebssy', href: 'https://t.me/ebssy' },
  { label: '@XCLLNT', href: 'https://t.me/XCLLNT' },
]

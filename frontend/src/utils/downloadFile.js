/**
 * Скачивание текстового файла в браузере (без backend).
 */
export function downloadTextFile(filename, content, mimeType = 'text/plain;charset=utf-8') {
  const blob = new Blob([content], { type: mimeType })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = filename
  a.rel = 'noopener'
  document.body.appendChild(a)
  a.click()
  document.body.removeChild(a)
  URL.revokeObjectURL(url)
}

/** Имя файла с меткой времени для экспорта */
export function exportFilename(prefix, ext) {
  const ts = new Date().toISOString().replace(/[:.]/g, '-').slice(0, 19)
  return `${prefix}-${ts}.${ext}`
}

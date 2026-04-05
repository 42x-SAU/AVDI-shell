import '@/styles/legal-document.css'

/**
 * Страница правил обработки персональных данных (заглушка «В разработке»).
 */
export function LegalPersonalDataPage() {
  return (
    <div className="legal-doc-page">
      <section className="legal-doc-page__hero" aria-labelledby="legal-pd-title">
        <div className="container legal-doc-page__inner">
          <h1 id="legal-pd-title" className="legal-doc-page__title legal-doc-page__title--center">
            В разработке
          </h1>
          <p className="legal-doc-page__lead legal-doc-page__lead--center">
            Проект находится в разработке или доступен в узких кругах общественности
          </p>
        </div>
      </section>
    </div>
  )
}

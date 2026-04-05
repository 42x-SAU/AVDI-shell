import { Link } from 'react-router-dom'
import { ROUTES } from '@/utils/constants'
import '@/pages/HomePage/HomePage.css'

/**
 * Главная страница: landing-like каркас под тематику AVDI Shell (тексты — рабочие заглушки).
 */
export function HomePage() {
  return (
    <div className="home-page">
      <section className="home-hero">
        <div className="container home-hero__inner">
          <div className="home-hero__copy">
            <p className="home-hero__eyebrow">Внутренняя диагностика инфраструктуры</p>
            <h1 className="home-hero__title">AVDI Shell</h1>
            <p className="home-hero__subtitle">
              Соберите целостную картину доступности сервисов, сетевой видимости и потенциально рискованных
              настроек — в одном спокойном интерфейсе для команд эксплуатации и безопасности.
            </p>
            <div className="home-hero__actions">
              <Link to={ROUTES.LOGIN} className="btn btn--primary">
                Перейти в панель мониторинга
              </Link>
              <a href="#home-why-teams" className="btn btn--ghost">
                Наши возможности
              </a>
            </div>
            <ul className="home-hero__facts" aria-label="Ключевые акценты">
              <li>Агенты уже работают на стороне инфраструктуры</li>
              <li>Интерфейс готов к подключению вашего API</li>
            </ul>
          </div>
          <div className="home-hero__visual" aria-hidden="true">
            <div className="home-hero__panel">
              <div className="home-hero__panel-row">
                <span className="home-hero__dot home-hero__dot--ok" />
                <span>Сводка узлов</span>
              </div>
              <div className="home-hero__panel-row">
                <span className="home-hero__dot home-hero__dot--warn" />
                <span>Проверка сервисов</span>
              </div>
              <div className="home-hero__panel-row">
                <span className="home-hero__dot home-hero__dot--idle" />
                <span>Сетевые маршруты</span>
              </div>
              <div className="home-hero__chart">
                <div className="home-hero__bar" style={{ height: '40%' }} />
                <div className="home-hero__bar" style={{ height: '65%' }} />
                <div className="home-hero__bar" style={{ height: '55%' }} />
                <div className="home-hero__bar" style={{ height: '80%' }} />
                <div className="home-hero__bar" style={{ height: '50%' }} />
              </div>
            </div>
          </div>
        </div>
      </section>

      <section id="home-why-teams" className="home-section home-section--anchor-target">
        <div className="container">
          <header className="home-section__head">
            <h2 className="home-section__title">Почему это удобно командам</h2>
            <p className="home-section__desc">
              Фокус на ясности: меньше ручного сбора данных, больше времени на решения. Ниже — типовые
              сценарии, которые поддерживает концепция AVDI Shell.
            </p>
          </header>
          <div className="home-cards">
            <article className="home-card">
              <h3 className="home-card__title">Диагностика доступности</h3>
              <p className="home-card__text">
                Быстро понять, какие сервисы отвечают ожидаемо, а где есть таймауты и деградации — без
                переключения между десятком инструментов.
              </p>
            </article>
            <article className="home-card">
              <h3 className="home-card__title">Обнаружение и учёт сервисов</h3>
              <p className="home-card__text">
                Упорядочить представление о том, что реально работает в среде: имена, порты, зависимости —
                как основа для аудита и изменений.
              </p>
            </article>
            <article className="home-card">
              <h3 className="home-card__title">Рискованные конфигурации</h3>
              <p className="home-card__text">
                Выделить настройки, которые чаще приводят к инцидентам: избыточные разрешения, слабые точки
                входа, несогласованные политики.
              </p>
            </article>
            <article className="home-card">
              <h3 className="home-card__title">Сетевая видимость</h3>
              <p className="home-card__text">
                Свести картину маршрутов и «дыр» видимости между сегментами — чтобы проще планировать
                изоляцию и исключения.
              </p>
            </article>
          </div>
        </div>
      </section>

      <section className="home-section home-section--alt">
        <div className="container">
          <header className="home-section__head">
            <h2 className="home-section__title">Сценарии использования</h2>
            <p className="home-section__desc">
              Платформа рассчитана на ежедневную эксплуатацию: от короткого «health-check» до подготовки
              отчётности и согласований.
            </p>
          </header>
          <div className="home-split">
            <div className="home-split__item">
              <h3 className="home-split__title">Единая сводка перед релизом</h3>
              <p className="home-split__text">
                Перед выкладкой проверить критичные зависимости и сетевые предпосылки — в виде понятной
                сводки для дежурного инженера.
              </p>
            </div>
            <div className="home-split__item">
              <h3 className="home-split__title">Разбор инцидента</h3>
              <p className="home-split__text">
                Сузить круг гипотез: доступность, маршрутизация, конфигурация — с опорой на собранные
                агентом данные, а не на хаотичные заметки.
              </p>
            </div>
            <div className="home-split__item">
              <h3 className="home-split__title">Подготовка к аудиту</h3>
              <p className="home-split__text">
                Зафиксировать «как есть» по ключевым областям, чтобы проще сравнивать изменения и
                договорённости между командами.
              </p>
            </div>
          </div>
        </div>
      </section>

      <section className="home-section">
        <div className="container">
          <div className="home-scale">
            <div className="home-scale__content">
              <h2 className="home-scale__title">Архитектурная гибкость</h2>
              <p className="home-scale__text">
                Интерфейс спроектирован как расширяемая оболочка: отдельные страницы, общие компоненты и
                слой доступа к API — чтобы подключать новые источники данных и визуализации без полной
                пересборки frontend.
              </p>
              <ul className="home-scale__list">
                <li>Маршрутизация и layout готовы к новым разделам</li>
                <li>Стили на CSS-переменных — темы и брендирование без «ломания» вёрстки</li>
                <li>Клиент API — точка расширения под авторизацию и контракты backend</li>
              </ul>
            </div>
            <div className="home-scale__aside" aria-hidden="true">
              <div className="home-scale__diagram">
                <div className="home-scale__node">UI</div>
                <div className="home-scale__arrow">→</div>
                <div className="home-scale__node">API слой</div>
                <div className="home-scale__arrow">→</div>
                <div className="home-scale__node">Агенты</div>
              </div>
              <p className="home-scale__hint">Схематично: место под будущие диаграммы и виджеты</p>
            </div>
          </div>
        </div>
      </section>
    </div>
  )
}

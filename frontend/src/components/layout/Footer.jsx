import { Link } from 'react-router-dom'
import tgIcon from '@/assets/icons/tg.svg'
import githubMark from '@/assets/images/github.png'
import madrigalLogo from '@/assets/images/madrigal.png'
import {
  FOOTER_GITHUB_URL,
  FOOTER_MADRIGAL_URL,
  FOOTER_TELEGRAM_LINKS,
  ROUTES,
} from '@/utils/constants'
import '@/components/layout/Footer.css'

/**
 * Подвал-заготовка: колонки ссылок и юридический блок — данные-заглушки под замену.
 */
export function Footer() {
  const year = new Date().getFullYear()

  return (
    <footer className="site-footer">
      <div className="site-footer__top container">
        <div className="site-footer__grid">
          <div className="site-footer__col">
            <h2 className="site-footer__heading">Продукт</h2>
            <ul className="site-footer__list">
              <li>
                <Link to={ROUTES.HOME}>Главная</Link>
              </li>
              <li>
                <span className="site-footer__muted">Сводка состояния (скоро)</span>
              </li>
              <li>
                <span className="site-footer__muted">Интеграции (скоро)</span>
              </li>
            </ul>
          </div>
          <div className="site-footer__col">
            <h2 className="site-footer__heading">Ресурсы</h2>
            <ul className="site-footer__list">
              <li>
                <span className="site-footer__muted">Документация (скоро)</span>
              </li>
              <li>
                <span className="site-footer__muted">Статус сервиса (скоро)</span>
              </li>
              <li>
                <span className="site-footer__muted">Поддержка (скоро)</span>
              </li>
            </ul>
          </div>
          <div className="site-footer__col site-footer__col--wide">
            <h2 className="site-footer__heading">AVDI Shell</h2>
            <p className="site-footer__lead">
              Единая точка входа для внутренней диагностики инфраструктуры: обзор доступности сервисов,
              сетевой видимости и рисков конфигурации — в удобном веб-интерфейсе.
            </p>
          </div>
        </div>
      </div>
      <div className="site-footer__bottom">
        <div className="container site-footer__bottom-inner">
          <div className="site-footer__copy-block">
            <p className="site-footer__copy">
              © {year} AVDI Shell. Разработано в рамках IT-консорциума «UMIRHack II».
            </p>
            <a
              className="site-footer__github"
              href={FOOTER_GITHUB_URL}
              target="_blank"
              rel="noopener noreferrer"
              aria-label="Репозиторий AVDI Shell на GitHub"
            >
              <img
                src={githubMark}
                alt=""
                width={240}
                height={240}
                className="site-footer__github-img"
                decoding="async"
              />
            </a>
            <div className="site-footer__tg-wrap">
              <ul className="site-footer__tg-list" aria-label="Telegram">
                {FOOTER_TELEGRAM_LINKS.map(({ label, href }) => (
                  <li key={href}>
                    <a
                      className="site-footer__tg-link"
                      href={href}
                      target="_blank"
                      rel="noopener noreferrer"
                    >
                      <img className="site-footer__tg-icon" src={tgIcon} alt="" width={14} height={14} />
                      <span>{label}</span>
                    </a>
                  </li>
                ))}
              </ul>
            </div>
          </div>
          <div className="site-footer__aside">
            <a
              className="site-footer__madrigal-link"
              href={FOOTER_MADRIGAL_URL}
              target="_blank"
              rel="noopener noreferrer"
              aria-label="Madrigal — открыть сайт в новой вкладке"
            >
              <img
                className="site-footer__madrigal-logo"
                src={madrigalLogo}
                alt=""
                width={280}
                height={120}
                decoding="async"
              />
            </a>
          </div>
        </div>
      </div>
    </footer>
  )
}

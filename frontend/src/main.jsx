import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import { BrowserRouter } from 'react-router-dom'
import NProgress from 'nprogress'
import 'nprogress/nprogress.css'
import '@/styles/nprogress-overrides.css'
import '@/styles/global.css'
import App from '@/App.jsx'

NProgress.configure({ showSpinner: false, trickleSpeed: 175, minimum: 0.12 })
NProgress.start()

createRoot(document.getElementById('root')).render(
  <StrictMode>
    <BrowserRouter>
      <App />
    </BrowserRouter>
  </StrictMode>,
)

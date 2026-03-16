import { createContext, useContext, useState } from 'react'
import tr from './tr.json'
import en from './en.json'

const locales = { tr, en }
const I18nContext = createContext()

export function I18nProvider({ children }) {
  const [lang, setLang] = useState(() => {
    try { return localStorage.getItem('lang') || 'tr' } catch { return 'tr' }
  })

  const switchLang = (l) => {
    setLang(l)
    try { localStorage.setItem('lang', l) } catch {}
  }

  const t = (key, params) => {
    const keys = key.split('.')
    let val = locales[lang]
    for (const k of keys) {
      val = val?.[k]
    }
    if (typeof val !== 'string') return key
    if (params) {
      return val.replace(/\{\{(\w+)\}\}/g, (_, k) => params[k] ?? '')
    }
    return val
  }

  return (
    <I18nContext.Provider value={{ lang, setLang: switchLang, t }}>
      {children}
    </I18nContext.Provider>
  )
}

export const useI18n = () => useContext(I18nContext)

import { useI18n } from '../i18n/I18nContext'

export default function LangSwitch() {
  const { lang, setLang } = useI18n()

  const active = 'bg-wg-red text-white shadow-sm'
  const inactive = 'text-wg-body dark:text-wg-gray-light/60 hover:text-wg-headline dark:hover:text-white'

  return (
    <div className="flex items-center bg-wg-gray-light dark:bg-wg-headline rounded-md p-0.5 text-sm" id="lang-switch">
      <button
        onClick={() => setLang('tr')}
        className={`px-3 py-1 rounded transition-all duration-200 font-medium ${
          lang === 'tr' ? active : inactive
        }`}
      >
        TR
      </button>
      <button
        onClick={() => setLang('en')}
        className={`px-3 py-1 rounded transition-all duration-200 font-medium ${
          lang === 'en' ? active : inactive
        }`}
      >
        EN
      </button>
    </div>
  )
}

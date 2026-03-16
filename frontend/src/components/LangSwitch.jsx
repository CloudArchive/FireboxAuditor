import { useI18n } from '../i18n/I18nContext'

export default function LangSwitch() {
  const { lang, setLang } = useI18n()

  return (
    <div className="flex items-center bg-gray-200 dark:bg-gray-800 rounded-lg p-0.5 text-sm">
      <button
        onClick={() => setLang('tr')}
        className={`px-3 py-1 rounded-md transition font-medium ${
          lang === 'tr' ? 'bg-white dark:bg-gray-700 text-gray-900 dark:text-white shadow-sm' : 'text-gray-500 dark:text-gray-400 hover:text-gray-900 dark:hover:text-white'
        }`}
      >
        TR
      </button>
      <button
        onClick={() => setLang('en')}
        className={`px-3 py-1 rounded-md transition font-medium ${
          lang === 'en' ? 'bg-white dark:bg-gray-700 text-gray-900 dark:text-white shadow-sm' : 'text-gray-500 dark:text-gray-400 hover:text-gray-900 dark:hover:text-white'
        }`}
      >
        EN
      </button>
    </div>
  )
}

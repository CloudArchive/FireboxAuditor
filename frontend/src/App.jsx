import { useState } from 'react'
import { useI18n } from './i18n/I18nContext'
import ScoreGauge from './components/ScoreGauge'
import AuditCard from './components/AuditCard'
import ConnectionForm from './components/ConnectionForm'
import UploadForm from './components/UploadForm'
import LangSwitch from './components/LangSwitch'
import ThemeSwitch from './components/ThemeSwitch'

export default function App() {
  const { t } = useI18n()
  const [report, setReport] = useState(null)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState(null)
  const [mode, setMode] = useState(null)

  const handleResult = async (fetchFn) => {
    setLoading(true)
    setError(null)
    setReport(null)
    try {
      const res = await fetchFn()
      if (!res.ok) {
        const body = await res.json().catch(() => ({}))
        throw new Error(body.error || `Server error (${res.status})`)
      }
      setReport(await res.json())
    } catch (e) {
      setError(e.message)
    } finally {
      setLoading(false)
    }
  }

  const sortedResults = report?.results
    ? [...report.results].sort((a, b) => {
        if (a.passed !== b.passed) return a.passed ? 1 : -1
        const sev = { critical: 0, high: 1, medium: 2 }
        return (sev[a.severity] ?? 3) - (sev[b.severity] ?? 3)
      })
    : []

  return (
    <div className="min-h-screen bg-gray-50 dark:bg-gray-950 transition-colors">
      <header className="border-b border-gray-200 dark:border-gray-800 bg-white/80 dark:bg-gray-900/50 backdrop-blur sticky top-0 z-10">
        <div className="max-w-6xl mx-auto px-6 py-4 flex items-center justify-between">
          <div className="flex items-center gap-3">
            <div className="w-10 h-10 rounded-lg bg-red-600 flex items-center justify-center font-bold text-lg text-white">
              FA
            </div>
            <div>
              <h1 className="text-xl font-bold text-gray-900 dark:text-white">{t('app.title')}</h1>
              <p className="text-xs text-gray-500 dark:text-gray-400">{t('app.subtitle')}</p>
            </div>
          </div>
          <div className="flex items-center gap-3">
            <ThemeSwitch />
            <LangSwitch />
            {report && (
              <button
                onClick={() => { setReport(null); setMode(null); setError(null) }}
                className="text-sm text-gray-500 dark:text-gray-400 hover:text-gray-900 dark:hover:text-white transition"
              >
                {t('app.newAudit')}
              </button>
            )}
          </div>
        </div>
      </header>

      <main className="max-w-6xl mx-auto px-6 py-10">
        {!report && !mode && (
          <div className="flex flex-col items-center gap-8 mt-16">
            <h2 className="text-2xl font-semibold text-gray-900 dark:text-white">{t('app.selectMethod')}</h2>
            <div className="grid grid-cols-1 sm:grid-cols-2 gap-6 w-full max-w-2xl">
              <button
                onClick={() => setMode('upload')}
                className="p-8 rounded-2xl border border-gray-200 dark:border-gray-800 bg-white dark:bg-gray-900 hover:border-blue-500 dark:hover:bg-gray-800 transition group"
              >
                <div className="text-4xl mb-4">&#128194;</div>
                <h3 className="text-lg font-semibold text-gray-900 dark:text-white group-hover:text-blue-500 dark:group-hover:text-blue-400 transition">{t('upload.cardTitle')}</h3>
                <p className="text-sm text-gray-500 dark:text-gray-400 mt-2">{t('upload.cardDesc')}</p>
              </button>
              <button
                onClick={() => setMode('ssh')}
                className="p-8 rounded-2xl border border-gray-200 dark:border-gray-800 bg-white dark:bg-gray-900 hover:border-green-500 dark:hover:bg-gray-800 transition group"
              >
                <div className="text-4xl mb-4">&#128274;</div>
                <h3 className="text-lg font-semibold text-gray-900 dark:text-white group-hover:text-green-500 dark:group-hover:text-green-400 transition">{t('ssh.cardTitle')}</h3>
                <p className="text-sm text-gray-500 dark:text-gray-400 mt-2">{t('ssh.cardDesc')}</p>
              </button>
            </div>
          </div>
        )}

        {!report && mode === 'upload' && (
          <div className="max-w-xl mx-auto mt-10">
            <button onClick={() => setMode(null)} className="text-sm text-gray-500 dark:text-gray-400 hover:text-gray-900 dark:hover:text-white mb-6 block">&larr; {t('app.back')}</button>
            <UploadForm onSubmit={handleResult} loading={loading} />
          </div>
        )}

        {!report && mode === 'ssh' && (
          <div className="max-w-xl mx-auto mt-10">
            <button onClick={() => setMode(null)} className="text-sm text-gray-500 dark:text-gray-400 hover:text-gray-900 dark:hover:text-white mb-6 block">&larr; {t('app.back')}</button>
            <ConnectionForm onSubmit={handleResult} loading={loading} />
          </div>
        )}

        {error && (
          <div className="max-w-xl mx-auto mt-6 p-4 rounded-xl bg-red-50 dark:bg-red-900/30 border border-red-200 dark:border-red-700 text-red-700 dark:text-red-300 text-sm">
            {error}
          </div>
        )}

        {loading && (
          <div className="flex justify-center mt-20">
            <div className="w-12 h-12 border-4 border-gray-300 dark:border-gray-700 border-t-blue-500 rounded-full animate-spin"></div>
          </div>
        )}

        {report && (
          <div className="space-y-10">
            <ScoreGauge score={report.score} />
            <div>
              <h2 className="text-lg font-semibold text-gray-900 dark:text-white mb-4">{t('app.auditResults')}</h2>
              <div className="space-y-4">
                {sortedResults.map((r) => (
                  <AuditCard key={r.rule_id} result={r} />
                ))}
              </div>
            </div>
          </div>
        )}
      </main>
    </div>
  )
}

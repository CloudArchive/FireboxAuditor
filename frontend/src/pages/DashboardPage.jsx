import { useState, useRef } from 'react'
import { useI18n } from '../i18n/I18nContext'
import { useAuth } from '../contexts/AuthContext'
import AuditHistoryCard from '../components/AuditHistoryCard'
import LangSwitch from '../components/LangSwitch'
import ThemeSwitch from '../components/ThemeSwitch'

export default function DashboardPage({ history, onView, onDelete, onNewAudit, loading }) {
  const { t } = useI18n()
  const { user, logout } = useAuth()

  return (
    <div className="min-h-screen hexagon-bg wg-watermark transition-colors relative">
      {/* Header */}
      <header className="border-b border-wg-gray-light dark:border-wg-headline/30 bg-white/95 dark:bg-wg-headline/90 backdrop-blur-md sticky top-0 z-20">
        <div className="max-w-4xl mx-auto px-6 py-2.5 flex items-center justify-between">
          <div className="flex items-center gap-3">
            <img src="/Icon_Magnifying_Glass.png" alt="WatchGuard" className="w-10 h-10 object-contain" />
            <div>
              <h1 className="text-lg font-semibold text-wg-headline dark:text-white tracking-tight">
                <span className="wg-accent mr-1">&gt;</span>
                {t('app.title')}
              </h1>
              <p className="text-xs text-wg-body dark:text-wg-gray-light/50">{t('app.subtitle')}</p>
            </div>
          </div>
          <div className="flex items-center gap-3">
            <ThemeSwitch />
            <LangSwitch />
            <span className="text-sm text-wg-body dark:text-wg-gray-light/60 hidden sm:inline">
              {user}
            </span>
            <button
              onClick={logout}
              className="text-sm px-3 py-1.5 rounded-md border border-wg-gray-light dark:border-wg-headline/50 text-wg-body dark:text-wg-gray-light/60 hover:border-wg-red hover:text-wg-red transition-colors"
            >
              {t('auth.logout')}
            </button>
          </div>
        </div>
      </header>

      <main className="max-w-4xl mx-auto px-6 py-10 relative z-10">
        {/* Page title + new audit button */}
        <div className="flex items-end justify-between mb-8">
          <div>
            <h2 className="text-2xl font-semibold text-wg-headline dark:text-white">
              {t('dashboard.title')}
            </h2>
            <p className="text-sm text-wg-body dark:text-wg-gray-light/50 mt-1">
              {t('dashboard.subtitle')}
            </p>
          </div>
          <UploadButton onAudit={onNewAudit} loading={loading} t={t} />
        </div>

        {/* History list */}
        <section>
          <h3 className="text-xs font-bold text-wg-body dark:text-wg-gray-light/50 uppercase tracking-wider mb-4">
            {t('dashboard.recentAudits')}
          </h3>

          {history.length === 0 ? (
            <div className="text-center py-20 rounded-2xl border-2 border-dashed border-wg-gray-light dark:border-wg-headline/30">
              <div className="text-5xl mb-4 opacity-40">📋</div>
              <p className="text-wg-headline dark:text-white font-medium">{t('dashboard.noAudits')}</p>
              <p className="text-sm text-wg-body dark:text-wg-gray-light/50 mt-1">{t('dashboard.noAuditsHint')}</p>
            </div>
          ) : (
            <div className="space-y-3">
              {history.map(record => (
                <AuditHistoryCard
                  key={record.id}
                  record={record}
                  onView={onView}
                  onDelete={onDelete}
                />
              ))}
            </div>
          )}
        </section>
      </main>
    </div>
  )
}

/* ── Inline Upload Button ─────────────────────────────────────────────────── */

function UploadButton({ onAudit, loading, t }) {
  const { apiFetch } = useAuth()
  const inputRef = useRef()
  const [fileName, setFileName] = useState('')

  const handleFile = async (e) => {
    const file = e.target.files?.[0]
    if (!file) return
    setFileName(file.name)

    const formData = new FormData()
    formData.append('config', file)

    await onAudit(async () => {
      const resp = await apiFetch('/api/audit/upload', { method: 'POST', body: formData })
      const json = await resp.json()
      if (!resp.ok) throw new Error(json.error || 'Upload failed')
      return json // { id, report }
    })

    // Reset input
    e.target.value = ''
    setFileName('')
  }

  return (
    <div>
      <input
        ref={inputRef}
        type="file"
        accept=".xml"
        className="hidden"
        onChange={handleFile}
      />
      <button
        onClick={() => inputRef.current?.click()}
        disabled={loading}
        className="flex items-center gap-2 px-5 py-2.5 rounded-lg bg-wg-red hover:bg-wg-red-hover text-white font-semibold text-sm transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
      >
        <span className="text-base">+</span>
        {loading ? t('upload.loading') : t('dashboard.newAuditBtn')}
      </button>
    </div>
  )
}

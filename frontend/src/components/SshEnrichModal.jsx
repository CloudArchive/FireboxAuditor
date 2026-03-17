import { useState } from 'react'
import { useI18n } from '../i18n/I18nContext'
import { useAuth } from '../contexts/AuthContext'

export default function SshEnrichModal({ auditId, onEnriched, onSkip, initialValues }) {
  const { t } = useI18n()
  const { apiFetch } = useAuth()
  const [form, setForm] = useState({
    host: initialValues?.host || '',
    port: initialValues?.port ? String(initialValues.port) : '4118',
    username: initialValues?.username || 'admin',
    password: '',
  })
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState(null)
  const [logs, setLogs] = useState([])
  const [success, setSuccess] = useState(false)

  const handleSubmit = async (e) => {
    e.preventDefault()
    setLoading(true)
    setError(null)
    setLogs([])

    try {
      const resp = await apiFetch('/api/ssh/enrich', {
        method: 'POST',
        body: JSON.stringify({
          audit_id: auditId,
          host: form.host,
          port: parseInt(form.port, 10),
          username: form.username,
          password: form.password,
        }),
      })
      const data = await resp.json()
      if (data.logs) setLogs(data.logs)

      if (!resp.ok) {
        setError(t('enrich.errorPrefix') + ': ' + (data.error || 'Bağlantı başarısız'))
        return
      }

      setSuccess(true)
      setTimeout(() => onEnriched(data.enrichment, data.logs || []), 1200)
    } catch (err) {
      setError(t('enrich.errorPrefix') + ': ' + err.message)
    } finally {
      setLoading(false)
    }
  }

  const field = (labelKey, key, type = 'text', placeholder = '') => (
    <div>
      <label className="block text-xs font-medium text-wg-body dark:text-wg-gray-light/70 mb-1">
        {t(labelKey)}
      </label>
      <input
        type={type}
        value={form[key]}
        onChange={e => setForm({ ...form, [key]: e.target.value })}
        placeholder={placeholder}
        required
        disabled={loading || success}
        className="w-full px-3 py-2 rounded-md bg-white dark:bg-wg-black border border-wg-gray-light dark:border-wg-headline text-wg-headline dark:text-wg-gray-light text-sm placeholder-wg-body/50 focus:outline-none focus:border-wg-red focus:ring-1 focus:ring-wg-red/30 transition disabled:opacity-60"
      />
    </div>
  )

  return (
    /* Backdrop */
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/50 backdrop-blur-sm animate-fade-in">
      <div className="w-full max-w-md rounded-2xl border border-wg-gray-light dark:border-wg-headline/50 bg-white dark:bg-[#1a1f26] shadow-2xl animate-slide-up">
        {/* Header */}
        <div className="px-6 pt-6 pb-4 border-b border-wg-gray-light dark:border-wg-headline/30">
          <div className="flex items-center gap-3">
            <div className="w-10 h-10 rounded-lg bg-wg-blue/10 dark:bg-wg-blue/20 flex items-center justify-center text-xl">
              🔑
            </div>
            <div>
              <h2 className="font-semibold text-wg-headline dark:text-white">
                {t('enrich.modalTitle')}
              </h2>
              <p className="text-xs text-wg-body dark:text-wg-gray-light/50">
                {t('enrich.modalSubtitle')}
              </p>
            </div>
          </div>
        </div>

        {/* Success state */}
        {success ? (
          <div className="px-6 py-10 text-center space-y-3">
            <div className="text-4xl">✅</div>
            <p className="font-semibold text-wg-headline dark:text-white">{t('enrich.successTitle')}</p>
            <p className="text-sm text-wg-body dark:text-wg-gray-light/60">{t('enrich.successDesc')}</p>
          </div>
        ) : (
          <form onSubmit={handleSubmit} className="px-6 py-5 space-y-4">
            <div className="grid grid-cols-3 gap-3">
              <div className="col-span-2">{field('enrich.host', 'host', 'text', '192.168.1.1')}</div>
              {field('enrich.port', 'port', 'number')}
            </div>
            {field('enrich.username', 'username')}
            {field('enrich.password', 'password', 'password')}

            {error && (
              <p className="text-xs text-wg-red bg-wg-red/5 border border-wg-red/20 rounded-md px-3 py-2">
                ⚠ {error}
              </p>
            )}

            {/* SSH Console */}
            {logs.length > 0 && (
              <div className="rounded-lg bg-black/80 p-3 font-mono text-[11px] text-emerald-400 max-h-28 overflow-y-auto space-y-0.5">
                {logs.map((l, i) => <div key={i}>{l}</div>)}
              </div>
            )}

            {/* Actions */}
            <div className="flex gap-3 pt-1">
              <button
                type="button"
                onClick={onSkip}
                disabled={loading}
                className="flex-1 py-2.5 rounded-md border border-wg-gray-light dark:border-wg-headline/50 text-wg-body dark:text-wg-gray-light/60 text-sm font-medium hover:border-wg-red hover:text-wg-red transition disabled:opacity-50"
              >
                {t('enrich.cancel')}
              </button>
              <button
                type="submit"
                disabled={loading}
                className="flex-[2] py-2.5 rounded-md bg-wg-red hover:bg-wg-red-hover text-white text-sm font-semibold transition disabled:opacity-50"
              >
                {loading ? t('enrich.loading') : t('enrich.submit')}
              </button>
            </div>
          </form>
        )}
      </div>
    </div>
  )
}

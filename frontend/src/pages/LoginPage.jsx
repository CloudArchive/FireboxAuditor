import { useState } from 'react'
import { useI18n } from '../i18n/I18nContext'
import { useAuth } from '../contexts/AuthContext'
import LangSwitch from '../components/LangSwitch'
import ThemeSwitch from '../components/ThemeSwitch'

export default function LoginPage() {
  const { t } = useI18n()
  const { login } = useAuth()
  const [form, setForm]     = useState({ username: '', password: '' })
  const [loading, setLoading] = useState(false)
  const [error, setError]   = useState(null)

  const handleSubmit = async (e) => {
    e.preventDefault()
    setLoading(true)
    setError(null)
    try {
      const resp = await fetch('/api/auth/login', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(form),
      })
      const data = await resp.json()
      if (!resp.ok) {
        setError(data.error || t('auth.errorInvalid'))
        return
      }
      login(data.token, data.username)
    } catch {
      setError(t('auth.errorServer'))
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="min-h-screen hexagon-bg wg-watermark flex flex-col">
      {/* Header */}
      <header className="border-b border-wg-gray-light dark:border-wg-headline/30 bg-white/95 dark:bg-wg-headline/90 backdrop-blur-md">
        <div className="max-w-6xl mx-auto px-6 py-2.5 flex items-center justify-between">
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
          </div>
        </div>
      </header>

      {/* Login Card */}
      <main className="flex-1 flex items-center justify-center px-6 py-16">
        <div className="w-full max-w-sm animate-slide-up">
          <div className="wg-card p-8 rounded-2xl border border-wg-gray-light dark:border-wg-headline/50 bg-white dark:bg-wg-headline/20 wg-concrete space-y-6">
            {/* Icon + Title */}
            <div className="text-center space-y-2">
              <div className="w-14 h-14 mx-auto rounded-xl bg-wg-red/10 dark:bg-wg-red/20 flex items-center justify-center text-3xl">
                🔐
              </div>
              <h2 className="text-xl font-semibold text-wg-headline dark:text-white">
                {t('auth.loginTitle')}
              </h2>
              <p className="text-sm text-wg-body dark:text-wg-gray-light/60">
                {t('auth.loginSubtitle')}
              </p>
            </div>

            {/* Form */}
            <form onSubmit={handleSubmit} className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-wg-body dark:text-wg-gray-light/70 mb-1.5">
                  {t('auth.username')}
                </label>
                <input
                  type="text"
                  autoComplete="username"
                  value={form.username}
                  onChange={e => setForm({ ...form, username: e.target.value })}
                  required
                  className="w-full px-4 py-2.5 rounded-md bg-white dark:bg-wg-black border border-wg-gray-light dark:border-wg-headline text-wg-headline dark:text-wg-gray-light placeholder-wg-body/50 focus:outline-none focus:border-wg-red focus:ring-1 focus:ring-wg-red/30 transition"
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-wg-body dark:text-wg-gray-light/70 mb-1.5">
                  {t('auth.password')}
                </label>
                <input
                  type="password"
                  autoComplete="current-password"
                  value={form.password}
                  onChange={e => setForm({ ...form, password: e.target.value })}
                  required
                  className="w-full px-4 py-2.5 rounded-md bg-white dark:bg-wg-black border border-wg-gray-light dark:border-wg-headline text-wg-headline dark:text-wg-gray-light placeholder-wg-body/50 focus:outline-none focus:border-wg-red focus:ring-1 focus:ring-wg-red/30 transition"
                />
              </div>

              {error && (
                <p className="text-sm text-wg-red bg-wg-red/5 dark:bg-wg-red/10 border border-wg-red/20 rounded-md px-3 py-2">
                  ⚠ {error}
                </p>
              )}

              <button
                type="submit"
                disabled={loading}
                className="w-full py-3 rounded-md bg-wg-red hover:bg-wg-red-hover text-white font-semibold transition-colors duration-200 disabled:opacity-50 disabled:cursor-not-allowed"
              >
                {loading ? t('auth.loginLoading') : t('auth.loginBtn')}
              </button>
            </form>
          </div>
        </div>
      </main>
    </div>
  )
}

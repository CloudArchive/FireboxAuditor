import { useState } from 'react'
import { useI18n } from '../i18n/I18nContext'

export default function ConnectionForm({ onSubmit, loading }) {
  const { t } = useI18n()
  const [form, setForm] = useState({ host: '', port: '4118', username: 'admin', password: '' })

  const handleSubmit = (e) => {
    e.preventDefault()
    onSubmit(() =>
      fetch('/api/audit/ssh', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ ...form, port: parseInt(form.port, 10) }),
      })
    )
  }

  const field = (labelKey, key, type = 'text', placeholder = '') => (
    <div>
      <label className="block text-sm font-medium text-wg-body dark:text-wg-gray-light/70 mb-1.5">{t(labelKey)}</label>
      <input
        type={type}
        value={form[key]}
        onChange={(e) => setForm({ ...form, [key]: e.target.value })}
        placeholder={placeholder}
        required
        className="w-full px-4 py-2.5 rounded-md bg-white dark:bg-wg-black border border-wg-gray-light dark:border-wg-headline text-wg-headline dark:text-wg-gray-light placeholder-wg-body/50 focus:outline-none focus:border-wg-red focus:ring-1 focus:ring-wg-red/30 transition"
      />
    </div>
  )

  return (
    <form onSubmit={handleSubmit} className="wg-card p-8 rounded-xl border border-wg-gray-light dark:border-wg-headline/50 bg-white dark:bg-wg-headline/20 wg-concrete space-y-5" id="ssh-form">
      <h3 className="text-lg font-medium text-wg-headline dark:text-white">
        <span className="wg-accent mr-1">&gt;</span>
        {t('ssh.title')}
      </h3>
      <div className="grid grid-cols-3 gap-4">
        <div className="col-span-2">{field('ssh.host', 'host', 'text', '192.168.1.1')}</div>
        {field('ssh.port', 'port', 'number')}
      </div>
      {field('ssh.username', 'username')}
      {field('ssh.password', 'password', 'password')}
      <button
        type="submit"
        disabled={loading}
        className="w-full py-3 rounded-md bg-wg-red hover:bg-wg-red-hover text-white font-semibold transition-colors duration-200 disabled:opacity-50 disabled:cursor-not-allowed"
        id="ssh-submit"
      >
        {loading ? t('ssh.loading') : t('ssh.submit')}
      </button>
    </form>
  )
}

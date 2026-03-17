import { useState } from 'react'
import { useI18n } from '../i18n/I18nContext'
import SshConsole from './SshConsole'

export default function ConnectionForm({ onSubmit, loading, onCancel }) {
  const { t } = useI18n()
  const [form, setForm] = useState({ host: '', port: '4118', username: 'admin', password: '' })
  const [sshLogs, setSshLogs] = useState([])
  const [lastAction, setLastAction] = useState(null)

  const handleAction = async (action) => {
    setLastAction(action)
    setSshLogs([])
    onSubmit(async () => {
      const resp = await fetch('/api/audit/ssh', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ ...form, port: parseInt(form.port, 10), action }),
      })
      
      const result = await resp.json()
      if (result.logs) {
        setSshLogs(result.logs)
      }
      
      return { ok: resp.ok, action, ...result }
    })
  }

  const handleSubmit = (e) => {
    e.preventDefault()
    handleAction('audit')
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
    <div className="space-y-4">
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
        <div className="flex gap-3 pt-2">
          <button
            type="button"
            disabled={loading}
            onClick={() => handleAction('sysinfo')}
            className="flex-1 py-3 rounded-md border border-wg-body/20 dark:border-wg-headline/50 text-wg-headline dark:text-white font-medium hover:bg-wg-gray-light dark:hover:bg-wg-headline/40 transition active:scale-95 disabled:opacity-50"
          >
            {loading && lastAction === 'sysinfo' ? '...' : 'SysInfo'}
          </button>
          <button
            type="button"
            disabled={loading}
            onClick={() => handleAction('feature-key')}
            className="flex-1 py-3 rounded-md border border-wg-body/20 dark:border-wg-headline/50 text-wg-headline dark:text-white font-medium hover:bg-wg-gray-light dark:hover:bg-wg-headline/40 transition active:scale-95 disabled:opacity-50"
          >
            {loading && lastAction === 'feature-key' ? '...' : 'Feature Key'}
          </button>
          <button
            type="submit"
            disabled={loading}
            className="flex-[2] py-3 rounded-md bg-wg-red hover:bg-wg-red-hover text-white font-semibold transition active:scale-95 disabled:opacity-50"
            id="ssh-submit"
          >
            {loading && lastAction === 'audit' ? t('ssh.loading') : t('ssh.submit')}
          </button>
        </div>
        {onCancel && (
          <button
            type="button"
            onClick={onCancel}
            className="w-full py-2 text-sm text-wg-body dark:text-wg-gray-light/40 hover:text-wg-red transition-colors"
          >
            {t('ssh.cancel')}
          </button>
        )}
      </form>

      <SshConsole logs={sshLogs} visible={sshLogs.length > 0 || loading} />
    </div>
  )
}

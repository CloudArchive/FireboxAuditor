import { useState } from 'react'

export default function ConnectionForm({ onSubmit, loading }) {
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

  const field = (label, key, type = 'text', placeholder = '') => (
    <div>
      <label className="block text-sm text-gray-400 mb-1">{label}</label>
      <input
        type={type}
        value={form[key]}
        onChange={(e) => setForm({ ...form, [key]: e.target.value })}
        placeholder={placeholder}
        required
        className="w-full px-4 py-2.5 rounded-lg bg-gray-800 border border-gray-700 text-white placeholder-gray-500 focus:outline-none focus:border-blue-500 transition"
      />
    </div>
  )

  return (
    <form onSubmit={handleSubmit} className="p-8 rounded-2xl border border-gray-800 bg-gray-900 space-y-5">
      <h3 className="text-lg font-semibold text-white">SSH Baglantisi</h3>
      <div className="grid grid-cols-3 gap-4">
        <div className="col-span-2">{field('IP Adresi', 'host', 'text', '192.168.1.1')}</div>
        {field('Port', 'port', 'number')}
      </div>
      {field('Kullanici Adi', 'username')}
      {field('Sifre', 'password', 'password')}
      <button
        type="submit"
        disabled={loading}
        className="w-full py-3 rounded-lg bg-green-600 hover:bg-green-500 text-white font-semibold transition disabled:opacity-50"
      >
        {loading ? 'Baglaniyor...' : 'Baglan ve Denetle'}
      </button>
    </form>
  )
}

import { useState, useEffect } from 'react'

export default function ThemeSwitch() {
  const [dark, setDark] = useState(() => {
    try {
      const saved = localStorage.getItem('theme')
      if (saved) return saved === 'dark'
    } catch {}
    return true
  })

  useEffect(() => {
    document.documentElement.classList.toggle('dark', dark)
    try { localStorage.setItem('theme', dark ? 'dark' : 'light') } catch {}
  }, [dark])

  return (
    <button
      onClick={() => setDark(!dark)}
      className="w-9 h-9 rounded-lg bg-gray-200 dark:bg-gray-800 flex items-center justify-center text-lg transition hover:opacity-80"
      title={dark ? 'Light Mode' : 'Dark Mode'}
    >
      {dark ? '\u2600\uFE0F' : '\uD83C\uDF19'}
    </button>
  )
}

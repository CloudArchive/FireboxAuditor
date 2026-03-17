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
      className="relative w-14 h-8 rounded-full transition-colors duration-300 focus:outline-none focus:ring-2 focus:ring-wg-red/40"
      style={{ backgroundColor: dark ? '#2D3237' : '#EAEAEA' }}
      title={dark ? 'Light Mode' : 'Dark Mode'}
      id="theme-toggle"
    >
      <span
        className="absolute top-1 left-1 w-6 h-6 rounded-full bg-white dark:bg-wg-red flex items-center justify-center text-sm shadow-md transition-all duration-300"
        style={{ transform: dark ? 'translateX(24px)' : 'translateX(0)' }}
      >
        {dark ? '🌙' : '☀️'}
      </span>
    </button>
  )
}

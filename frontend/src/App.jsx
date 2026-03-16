import { useState } from 'react'
import ScoreGauge from './components/ScoreGauge'
import AuditCard from './components/AuditCard'
import ConnectionForm from './components/ConnectionForm'
import UploadForm from './components/UploadForm'

export default function App() {
  const [report, setReport] = useState(null)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState(null)
  const [mode, setMode] = useState(null) // 'upload' | 'ssh'

  const handleResult = async (fetchFn) => {
    setLoading(true)
    setError(null)
    setReport(null)
    try {
      const res = await fetchFn()
      if (!res.ok) {
        const body = await res.json().catch(() => ({}))
        throw new Error(body.error || `Sunucu hatası (${res.status})`)
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
    <div className="min-h-screen bg-gray-950">
      {/* Header */}
      <header className="border-b border-gray-800 bg-gray-900/50 backdrop-blur sticky top-0 z-10">
        <div className="max-w-6xl mx-auto px-6 py-4 flex items-center justify-between">
          <div className="flex items-center gap-3">
            <div className="w-10 h-10 rounded-lg bg-red-600 flex items-center justify-center font-bold text-lg">
              FA
            </div>
            <div>
              <h1 className="text-xl font-bold text-white">Firebox Auditor</h1>
              <p className="text-xs text-gray-400">WatchGuard Configuration Security Audit</p>
            </div>
          </div>
          {report && (
            <button
              onClick={() => { setReport(null); setMode(null); setError(null) }}
              className="text-sm text-gray-400 hover:text-white transition"
            >
              Yeni Denetim
            </button>
          )}
        </div>
      </header>

      <main className="max-w-6xl mx-auto px-6 py-10">
        {/* Mode selection */}
        {!report && !mode && (
          <div className="flex flex-col items-center gap-8 mt-16">
            <h2 className="text-2xl font-semibold text-white">Denetim Yöntemini Seçin</h2>
            <div className="grid grid-cols-1 sm:grid-cols-2 gap-6 w-full max-w-2xl">
              <button
                onClick={() => setMode('upload')}
                className="p-8 rounded-2xl border border-gray-800 bg-gray-900 hover:border-blue-500 hover:bg-gray-800 transition group"
              >
                <div className="text-4xl mb-4">&#128194;</div>
                <h3 className="text-lg font-semibold text-white group-hover:text-blue-400 transition">XML Dosya Yukle</h3>
                <p className="text-sm text-gray-400 mt-2">Daha once disa aktardiginiz .xml konfigurasyon dosyasini yukleyin.</p>
              </button>
              <button
                onClick={() => setMode('ssh')}
                className="p-8 rounded-2xl border border-gray-800 bg-gray-900 hover:border-green-500 hover:bg-gray-800 transition group"
              >
                <div className="text-4xl mb-4">&#128274;</div>
                <h3 className="text-lg font-semibold text-white group-hover:text-green-400 transition">CLI ile Baglan</h3>
                <p className="text-sm text-gray-400 mt-2">SSH uzerinden Firebox cihazina baglanip konfigurasyonu otomatik alin.</p>
              </button>
            </div>
          </div>
        )}

        {/* Forms */}
        {!report && mode === 'upload' && (
          <div className="max-w-xl mx-auto mt-10">
            <button onClick={() => setMode(null)} className="text-sm text-gray-400 hover:text-white mb-6 block">&larr; Geri</button>
            <UploadForm onSubmit={handleResult} loading={loading} />
          </div>
        )}

        {!report && mode === 'ssh' && (
          <div className="max-w-xl mx-auto mt-10">
            <button onClick={() => setMode(null)} className="text-sm text-gray-400 hover:text-white mb-6 block">&larr; Geri</button>
            <ConnectionForm onSubmit={handleResult} loading={loading} />
          </div>
        )}

        {error && (
          <div className="max-w-xl mx-auto mt-6 p-4 rounded-xl bg-red-900/30 border border-red-700 text-red-300 text-sm">
            {error}
          </div>
        )}

        {loading && (
          <div className="flex justify-center mt-20">
            <div className="w-12 h-12 border-4 border-gray-700 border-t-blue-500 rounded-full animate-spin"></div>
          </div>
        )}

        {/* Report */}
        {report && (
          <div className="space-y-10">
            <ScoreGauge score={report.score} />
            <div>
              <h2 className="text-lg font-semibold text-white mb-4">Denetim Sonuclari</h2>
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

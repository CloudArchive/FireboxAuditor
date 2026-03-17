import { useState, useEffect } from 'react'
import { useI18n } from './i18n/I18nContext'
import ScoreGauge from './components/ScoreGauge'
import AuditCard from './components/AuditCard'
import DeviceInfoCard from './components/DeviceInfoCard'
import ConnectionForm from './components/ConnectionForm'
import UploadForm from './components/UploadForm'
import LangSwitch from './components/LangSwitch'
import ThemeSwitch from './components/ThemeSwitch'
import PolicyTable from './components/PolicyTable'

/* ── WatchGuard Logo SVG ──────────────────────────── */
function WGLogo({ className = '' }) {
  return (
    <div className={`${className} flex items-center justify-center p-1.5 dark:bg-wg-gray-light/10 dark:backdrop-blur-sm rounded-xl transition-all duration-300`}>
      <img 
        src="/Icon_Magnifying_Glass.png" 
        alt="WatchGuard" 
        className="w-full h-full object-contain filter drop-shadow-sm"
      />
    </div>
  )
}

export default function App() {
  const { t } = useI18n()
  const [report, setReport] = useState(null)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState(null)
  const [mode, setMode] = useState(null)
  const [highlightedIndices, setHighlightedIndices] = useState([])
  const [sysInfo, setSysInfo] = useState(null)
  const [featureKey, setFeatureKey] = useState(null)
  const [isEnriching, setIsEnriching] = useState(false)

  const handleResult = async (fetchFn) => {
    setLoading(true)
    setError(null)
    try {
      const result = await fetchFn()
      if (!result.ok && result.error) {
        throw new Error(result.error)
      }
      
      // result from ConnectionForm is { ok, data, logs, action }
      const { data, action } = result

      if (action === 'sysinfo' && data) {
        setSysInfo(data)
        setIsEnriching(false)
      } else if (action === 'feature-key' && data) {
        setFeatureKey(typeof data === 'string' ? data : JSON.stringify(data, null, 2))
      } else if (result.report) {
        setReport(result.report)
      } else if (data && !action) {
        // Fallback or upload case
        setReport(data)
      }
    } catch (e) {
      setError(e.message)
    } finally {
      setLoading(false)
    }
  }

  // Handle enrichment event
  useEffect(() => {
    const handler = () => setIsEnriching(true)
    window.addEventListener('open-ssh-enrich', handler)
    return () => window.removeEventListener('open-ssh-enrich', handler)
  }, [])

  const sortedResults = report?.results
    ? [...report.results].sort((a, b) => {
        if (a.passed !== b.passed) return a.passed ? 1 : -1
        const sev = { critical: 0, high: 1, medium: 2 }
        return (sev[a.severity] ?? 3) - (sev[b.severity] ?? 3)
      })
    : []

  return (
    <div className="min-h-screen hexagon-bg wg-watermark transition-colors relative">
      {/* ── Header ──────────────────────────────────── */}
      <header className="border-b border-wg-gray-light dark:border-wg-headline/30 bg-white/95 dark:bg-wg-headline/90 backdrop-blur-md sticky top-0 z-20">
        <div className="max-w-6xl mx-auto px-6 py-2.5 flex items-center justify-between">
          <div className="flex items-center gap-4">
            <WGLogo className="w-14 h-14" />
            <div>
              <h1 className="text-xl font-semibold text-wg-headline dark:text-white tracking-tight">
                <span className="wg-accent mr-1">&gt;</span>
                {t('app.title')}
              </h1>
              <p className="text-xs text-wg-body dark:text-wg-gray-light/50">{t('app.subtitle')}</p>
            </div>
          </div>
          <div className="flex items-center gap-3">
            <ThemeSwitch />
            <LangSwitch />
            {report && (
              <button
                onClick={() => { 
                  setReport(null); 
                  setMode(null); 
                  setError(null);
                  setSysInfo(null);
                  setFeatureKey(null);
                }}
                className="ml-2 text-sm px-4 py-2 rounded-md border border-wg-red text-wg-red hover:bg-wg-red hover:text-white transition-colors duration-200 font-medium"
                id="new-audit-btn"
              >
                {t('app.newAudit')}
              </button>
            )}
          </div>
        </div>
      </header>

      {/* ── Main Content ────────────────────────────── */}
      <main className="max-w-6xl mx-auto px-6 py-10 relative z-10">

        {/* Method Selection */}
        {!report && !mode && (
          <div className="flex flex-col items-center gap-8 mt-16 animate-fade-in">
            <h2 className="text-2xl font-medium text-wg-headline dark:text-white">
              <span className="wg-accent mr-2">&gt;</span>
              {t('app.selectMethod')}
            </h2>
            <div className="grid grid-cols-1 sm:grid-cols-2 gap-6 w-full max-w-2xl">
              <button
                onClick={() => setMode('upload')}
                className="p-8 rounded-xl border border-wg-gray-light dark:border-wg-headline/40 bg-white dark:bg-wg-headline/15 hover:border-wg-red dark:hover:border-wg-red transition-all duration-300 group text-left"
                id="select-upload"
              >
                <div className="w-12 h-12 rounded-lg bg-wg-red/10 dark:bg-wg-red/20 flex items-center justify-center text-2xl mb-4 group-hover:bg-wg-red group-hover:text-white transition-colors duration-300">
                  📄
                </div>
                <h3 className="text-lg font-medium text-wg-headline dark:text-white group-hover:text-wg-red dark:group-hover:text-wg-red transition-colors duration-200">
                  {t('upload.cardTitle')}
                </h3>
                <p className="text-sm text-wg-body dark:text-wg-gray-light/60 mt-2">{t('upload.cardDesc')}</p>
              </button>
              <button
                onClick={() => setMode('ssh')}
                className="p-8 rounded-xl border border-wg-gray-light dark:border-wg-headline/40 bg-white dark:bg-wg-headline/15 hover:border-wg-red dark:hover:border-wg-red transition-all duration-300 group text-left"
                id="select-ssh"
              >
                <div className="w-12 h-12 rounded-lg bg-wg-red/10 dark:bg-wg-red/20 flex items-center justify-center text-2xl mb-4 group-hover:bg-wg-red group-hover:text-white transition-colors duration-300">
                  🔐
                </div>
                <h3 className="text-lg font-medium text-wg-headline dark:text-white group-hover:text-wg-red dark:group-hover:text-wg-red transition-colors duration-200">
                  {t('ssh.cardTitle')}
                </h3>
                <p className="text-sm text-wg-body dark:text-wg-gray-light/60 mt-2">{t('ssh.cardDesc')}</p>
              </button>
            </div>
          </div>
        )}

        {/* Upload Form */}
        {!report && mode === 'upload' && (
          <div className="max-w-xl mx-auto mt-10 animate-slide-up">
            <button
              onClick={() => setMode(null)}
              className="text-sm text-wg-body dark:text-wg-gray-light/50 hover:text-wg-red dark:hover:text-wg-red mb-6 block transition-colors font-medium"
            >
              &larr; {t('app.back')}
            </button>
            <UploadForm onSubmit={handleResult} loading={loading} />
          </div>
        )}

        {/* SSH Form */}
        {!report && mode === 'ssh' && (
          <div className="max-w-xl mx-auto mt-10 animate-slide-up">
            <button
              onClick={() => setMode(null)}
              className="text-sm text-wg-body dark:text-wg-gray-light/50 hover:text-wg-red dark:hover:text-wg-red mb-6 block transition-colors font-medium"
            >
              &larr; {t('app.back')}
            </button>
            <ConnectionForm onSubmit={handleResult} loading={loading} />
          </div>
        )}

        {/* Error */}
        {error && (
          <div className="max-w-xl mx-auto mt-6 p-4 rounded-xl bg-wg-red/5 dark:bg-wg-red/10 border border-wg-red/20 text-wg-red text-sm animate-fade-in" id="error-banner">
            <span className="font-semibold mr-1">⚠</span> {error}
          </div>
        )}

        {/* Loading */}
        {loading && (
          <div className="flex flex-col items-center justify-center mt-20 gap-4 animate-fade-in">
            <div className="w-12 h-12 border-4 border-wg-gray-light dark:border-wg-headline border-t-wg-red rounded-full animate-spin"></div>
            <p className="text-sm text-wg-body dark:text-wg-gray-light/50">{t('app.subtitle')}</p>
          </div>
        )}

        {/* Results */}
        {(report || sysInfo) && (
          <div className="space-y-10 animate-fade-in">
            <DeviceInfoCard info={{
              ...(report?.device_info || {}),
              model: sysInfo?.model || report?.device_info?.model,
              serial_number: sysInfo?.serial_number || report?.device_info?.serial_number,
              firmware_version: sysInfo?.version || report?.device_info?.firmware_version,
              system_name: sysInfo?.system_name || report?.device_info?.system_name,
              contact: sysInfo?.contact || report?.device_info?.contact,
              location: sysInfo?.location || report?.device_info?.location,
              up_time: sysInfo?.up_time,
              cpu_usage: sysInfo?.cpu_usage,
              memory_usage: sysInfo?.memory_usage
            }} />
            {report && <ScoreGauge score={report.score} />}
            
            {featureKey && (
              <div className="wg-card p-6 border-wg-gray-light dark:border-wg-headline/50 bg-white dark:bg-wg-headline/10">
                <h3 className="text-sm font-semibold text-wg-headline dark:text-white mb-3 uppercase tracking-wider">Feature Key</h3>
                <pre className="text-xs font-mono text-wg-body dark:text-wg-gray-light/70 bg-black/5 dark:bg-black/20 p-4 rounded overflow-auto max-h-60 leading-relaxed capitalize whitespace-pre-wrap">
                  {featureKey}
                </pre>
              </div>
            )}
            <div>
              <h2 className="text-lg font-medium text-wg-headline dark:text-white mb-4">
                <span className="wg-accent mr-2">&gt;</span>
                {t('app.auditResults')}
              </h2>
              <div className="space-y-4">
                {sortedResults.map((r, i) => (
                  <div 
                    key={r.rule_id} 
                    style={{ animationDelay: `${i * 60}ms` }}
                    onClick={() => {
                      const matches = r.details?.join(' ').match(/\[(\d+)\]/g) || []
                      const indices = matches.map(m => parseInt(m.replace(/[\[\]]/g, '')))
                      setHighlightedIndices(indices)
                      if (indices.length > 0) {
                        document.getElementById(`policy-row-${indices[0]}`)?.scrollIntoView({ behavior: 'smooth', block: 'center' })
                      }
                    }}
                    className="cursor-pointer"
                  >
                    <AuditCard result={r} isHighlighted={r.details?.some(d => highlightedIndices.some(idx => d.includes(`[${idx}]`)))} />
                  </div>
                ))}
              </div>
            </div>

            {report.policies && (
              <PolicyTable 
                policies={report.policies} 
                highlightedIndices={highlightedIndices} 
              />
            )}
          </div>
        )}
      </main>

      {/* SSH Enrichment Modal */}
      {isEnriching && (
        <div className="fixed inset-0 z-50 flex items-center justify-center p-6 bg-wg-headline/80 backdrop-blur-sm animate-fade-in">
          <div className="w-full max-w-xl animate-scale-in">
            <div className="flex justify-between items-center mb-4 px-2">
              <h3 className="text-xl font-semibold text-white">SSH Bilgi Güncelleme</h3>
              <button 
                onClick={() => setIsEnriching(false)}
                className="text-white/60 hover:text-white transition-colors"
                id="close-enrich-btn"
              >
                ✕
              </button>
            </div>
            <ConnectionForm 
              onSubmit={handleResult} 
              loading={loading} 
              onCancel={() => setIsEnriching(false)} 
            />
          </div>
        </div>
      )}

      {/* ── Footer ──────────────────────────────────── */}
      <footer className="relative z-10 border-t border-wg-gray-light dark:border-wg-headline/20 py-6 mt-10">
        <div className="max-w-6xl mx-auto px-6 flex items-center justify-between">
          <p className="text-xs text-wg-body dark:text-wg-gray-light/40">
            © {new Date().getFullYear()} CloudArchive. All rights reserved.
          </p>
          <p className="text-xs text-wg-body/60 dark:text-wg-gray-light/30">
            Firebox Auditor v1.0
          </p>
        </div>
      </footer>
    </div>
  )
}

import { useState } from 'react'
import { useI18n } from '../i18n/I18nContext'
import { useAuth } from '../contexts/AuthContext'
import DeviceInfoCard from '../components/DeviceInfoCard'
import ScoreGauge from '../components/ScoreGauge'
import AuditCard from '../components/AuditCard'
import PolicyTable from '../components/PolicyTable'
import PolicyDetailPanel from '../components/PolicyDetailPanel'
import SshEnrichModal from '../components/SshEnrichModal'
import SshLogsModal from '../components/SshLogsModal'
import LangSwitch from '../components/LangSwitch'
import ThemeSwitch from '../components/ThemeSwitch'

export default function AuditPage({ auditRecord, onBack, onRecordUpdate }) {
  const { t } = useI18n()
  const { user, logout } = useAuth()

  const [showEnrich, setShowEnrich]         = useState(false)
  const [enrichment, setEnrichment]         = useState(auditRecord?.enrichment || null)
  const [sshLogs, setSshLogs]               = useState([])
  const [showLogs, setShowLogs]             = useState(false)
  const [highlightedIndices, setHighlighted] = useState([])
  const [selectedPolicy, setSelectedPolicy]   = useState(null)

  const { report } = auditRecord
  const deviceInfo  = report?.device_info || {}

  const sortedResults = report?.results
    ? [...report.results].sort((a, b) => {
        if (a.passed !== b.passed) return a.passed ? 1 : -1
        const sev = { critical: 0, high: 1, medium: 2 }
        return (sev[a.severity] ?? 3) - (sev[b.severity] ?? 3)
      })
    : []

  const handleEnriched = (newEnrichment, logs = []) => {
    setEnrichment(newEnrichment)
    setSshLogs(logs)
    setShowEnrich(false)
    onRecordUpdate?.({ ...auditRecord, enrichment: newEnrichment })
  }

  const handleDisconnect = () => {
    setEnrichment(null)
    onRecordUpdate?.({ ...auditRecord, enrichment: null })
  }

  return (
    <div className="min-h-screen hexagon-bg wg-watermark transition-colors relative">

      {/* SSH Logs Modal */}
      {showLogs && (
        <SshLogsModal
          logs={sshLogs}
          host={enrichment?.ssh_host || ''}
          onClose={() => setShowLogs(false)}
        />
      )}

      {/* SSH Enrich Modal */}
      {showEnrich && (
        <SshEnrichModal
          auditId={auditRecord.id}
          onEnriched={handleEnriched}
          onSkip={() => setShowEnrich(false)}
          initialValues={enrichment ? {
            host: enrichment.ssh_host,
            port: enrichment.ssh_port,
            username: enrichment.ssh_username,
          } : null}
        />
      )}

      {/* Header */}
      <header className="border-b border-wg-gray-light dark:border-wg-headline/30 bg-white/95 dark:bg-wg-headline/90 backdrop-blur-md sticky top-0 z-20">
        <div className="max-w-6xl mx-auto px-6 py-2.5 flex items-center justify-between">
          <div className="flex items-center gap-3">
            <img src="/Icon_Magnifying_Glass.png" alt="WatchGuard" className="w-10 h-10 object-contain" />
            <div>
              <h1 className="text-lg font-semibold text-wg-headline dark:text-white tracking-tight">
                <span className="wg-accent mr-1">&gt;</span>
                {t('app.title')}
              </h1>
              <p className="text-xs text-wg-body dark:text-wg-gray-light/50">
                {auditRecord.device_name} · {auditRecord.file_name}
              </p>
            </div>
          </div>
          <div className="flex items-center gap-3">
            <ThemeSwitch />
            <LangSwitch />
            <button
              onClick={onBack}
              className="text-sm px-4 py-1.5 rounded-md border border-wg-gray-light dark:border-wg-headline/50 text-wg-body dark:text-wg-gray-light/60 hover:border-wg-red hover:text-wg-red transition-colors"
            >
              ← {t('app.back')}
            </button>
          </div>
        </div>
      </header>

      <main className="max-w-6xl mx-auto px-6 py-10 relative z-10 space-y-10">

        {/* Enrich banner — show only if not yet enriched */}
        {!enrichment && (
          <div className="rounded-xl border border-wg-blue/30 dark:border-wg-blue/20 bg-wg-blue/5 dark:bg-wg-blue/10 px-5 py-4 flex items-center justify-between gap-4 animate-slide-up">
            <div>
              <p className="font-semibold text-wg-headline dark:text-white text-sm">
                🔑 {t('enrich.bannerTitle')}
              </p>
              <p className="text-xs text-wg-body dark:text-wg-gray-light/60 mt-0.5">
                {t('enrich.bannerDesc')}
              </p>
            </div>
            <div className="flex items-center gap-3 shrink-0">
              <button
                onClick={() => setShowEnrich(true)}
                className="px-4 py-2 rounded-lg bg-wg-red hover:bg-wg-red-hover text-white text-sm font-semibold transition active:scale-95"
              >
                {t('enrich.connectBtn')}
              </button>
              <button
                onClick={() => setEnrichment(false)} // mark as intentionally skipped
                className="text-xs text-wg-body dark:text-wg-gray-light/50 hover:text-wg-red transition underline-offset-2 hover:underline"
              >
                {t('enrich.skipBtn')}
              </button>
            </div>
          </div>
        )}

        {/* Two-column layout: device card left, score right */}
        <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
          <div className="lg:col-span-2">
            <DeviceInfoCard
              info={deviceInfo}
              enrichment={enrichment || null}
              onEnrichRequest={() => setShowEnrich(true)}
              onReconnect={() => setShowEnrich(true)}
              onDisconnect={handleDisconnect}
              onShowLogs={sshLogs.length > 0 ? () => setShowLogs(true) : undefined}
            />
          </div>
          <div className="flex flex-col gap-6">
            <ScoreGauge score={report?.score ?? 0} />
          </div>
        </div>

        {/* Audit Results */}
        <div>
          <h2 className="text-lg font-semibold text-wg-headline dark:text-white mb-4 flex items-center gap-2">
            <span className="bg-wg-red w-1 h-5 rounded-full" />
            {t('app.auditResults')}
          </h2>
          <div className="space-y-4">
            {sortedResults.map((r, i) => (
              <div
                key={r.rule_id}
                style={{ animationDelay: `${i * 60}ms` }}
                onClick={() => {
                  const matches = r.details?.join(' ').match(/\[(\d+)\]/g) || []
                  setHighlighted(matches.map(m => parseInt(m.replace(/[\[\]]/g, ''))))
                }}
                className="cursor-pointer"
              >
                <AuditCard result={r} />
              </div>
            ))}
          </div>
        </div>

        {/* Policy Table */}
        {report?.policies?.length > 0 && (
          <PolicyTable
            policies={report.policies}
            highlightedIndices={highlightedIndices}
            onSelectPolicy={setSelectedPolicy}
          />
        )}
      </main>

      {selectedPolicy && (
        <PolicyDetailPanel
          policy={selectedPolicy}
          aliases={report?.aliases || []}
          onClose={() => setSelectedPolicy(null)}
        />
      )}
    </div>
  )
}

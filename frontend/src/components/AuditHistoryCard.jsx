import { useI18n } from '../i18n/I18nContext'

function ScoreDot({ score }) {
  if (score >= 80) return <span className="w-2.5 h-2.5 rounded-full bg-emerald-500 inline-block" />
  if (score >= 50) return <span className="w-2.5 h-2.5 rounded-full bg-amber-400 inline-block" />
  return <span className="w-2.5 h-2.5 rounded-full bg-wg-red inline-block" />
}

function formatDate(dateStr, lang) {
  const d = new Date(dateStr)
  const now = new Date()
  const diffMs = now - d
  const diffH = Math.floor(diffMs / 3600000)
  const diffD = Math.floor(diffMs / 86400000)

  if (diffH < 1) return lang === 'tr' ? 'Az önce' : 'Just now'
  if (diffH < 24) return lang === 'tr' ? `${diffH} saat önce` : `${diffH}h ago`
  if (diffD === 1) return lang === 'tr' ? 'Dün' : 'Yesterday'
  if (diffD < 7) return lang === 'tr' ? `${diffD} gün önce` : `${diffD}d ago`

  return d.toLocaleDateString(lang === 'tr' ? 'tr-TR' : 'en-GB', {
    day: 'numeric', month: 'short', year: 'numeric',
  })
}

export default function AuditHistoryCard({ record, onView, onDelete }) {
  const { t, lang } = useI18n()

  const criticalCount = 0  // We don't have full results in summary; backend could add this
  const scoreColor =
    record.score >= 80 ? 'text-emerald-600 dark:text-emerald-400' :
    record.score >= 50 ? 'text-amber-500 dark:text-amber-400' :
    'text-wg-red'

  return (
    <div className="wg-card rounded-xl border border-wg-gray-light dark:border-wg-headline/40 bg-white dark:bg-wg-headline/15 p-5 transition-all hover:border-wg-red/30 hover:shadow-md dark:hover:shadow-black/20">
      <div className="flex items-start justify-between gap-4">
        {/* Left: info */}
        <div className="min-w-0 flex-1">
          <div className="flex items-center gap-2 mb-1">
            <ScoreDot score={record.score} />
            <span className="font-semibold text-wg-headline dark:text-white truncate">
              {record.device_name || 'Firebox'}
            </span>
            {record.enriched && (
              <span className="text-[10px] px-1.5 py-0.5 rounded bg-wg-blue/10 dark:bg-wg-blue/20 text-wg-blue dark:text-blue-400 font-medium shrink-0">
                SSH ✓
              </span>
            )}
          </div>
          <p className="text-xs text-wg-body dark:text-wg-gray-light/50 truncate mb-3">
            📄 {record.file_name} &nbsp;·&nbsp; {formatDate(record.created_at, lang)}
          </p>
          <div className="flex items-center gap-3">
            <span className={`text-2xl font-bold ${scoreColor}`}>
              {record.score}
            </span>
            <span className="text-xs text-wg-body dark:text-wg-gray-light/50">
              {t('dashboard.score')}
            </span>
          </div>
        </div>

        {/* Right: actions */}
        <div className="flex flex-col gap-2 shrink-0">
          <button
            onClick={() => onView(record.id)}
            className="px-4 py-1.5 rounded-md bg-wg-red hover:bg-wg-red-hover text-white text-sm font-medium transition-colors"
          >
            {t('dashboard.viewReport')}
          </button>
          <button
            onClick={() => {
              if (window.confirm(t('dashboard.deleteConfirm'))) onDelete(record.id)
            }}
            className="px-4 py-1.5 rounded-md border border-wg-gray-light dark:border-wg-headline/50 text-wg-body dark:text-wg-gray-light/60 text-sm hover:border-wg-red hover:text-wg-red transition-colors"
          >
            {t('dashboard.deleteAudit')}
          </button>
        </div>
      </div>
    </div>
  )
}

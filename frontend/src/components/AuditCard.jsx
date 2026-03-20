import { useI18n } from '../i18n/I18nContext'

const severityStyles = {
  critical: {
    border: 'border-wg-red/30 dark:border-wg-red/40',
    bg: 'bg-wg-red/5 dark:bg-wg-red/10',
    badge: 'bg-wg-red/10 text-wg-red dark:bg-wg-red/20 dark:text-red-300',
    icon: '🔴',
  },
  high: {
    border: 'border-orange-300 dark:border-orange-700',
    bg: 'bg-orange-50 dark:bg-orange-900/15',
    badge: 'bg-orange-100 text-orange-700 dark:bg-orange-900/30 dark:text-orange-300',
    icon: '🟠',
  },
  medium: {
    border: 'border-yellow-300 dark:border-yellow-700',
    bg: 'bg-yellow-50 dark:bg-yellow-900/15',
    badge: 'bg-yellow-100 text-yellow-700 dark:bg-yellow-900/30 dark:text-yellow-300',
    icon: '🟡',
  },
}

export default function AuditCard({ result }) {
  const { t } = useI18n()
  const { passed, severity, rule_id, details, points_deducted } = result
  const sev = severityStyles[severity] || severityStyles.medium

  const ruleName = t(`rules.${rule_id}.name`)
  const description = passed
    ? t(`rules.${rule_id}.descPass`)
    : t(`rules.${rule_id}.descFail`, { details: details?.join(', ') || '' })
  const remediation = t(`rules.${rule_id}.remediation`)

  if (passed) {
    return (
      <div className="wg-card p-5 rounded-xl border border-wg-gray-light dark:border-wg-headline/40 bg-white dark:bg-wg-headline/15 flex items-start gap-4">
        <div className="text-green-500 text-2xl mt-0.5">✔</div>
        <div className="flex-1">
          <div className="flex items-center gap-3 flex-wrap">
            <span className="font-medium text-wg-headline dark:text-white">{ruleName}</span>
            <span className="text-xs px-2 py-0.5 rounded bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-300">{t('status.passed')}</span>
            <span className="wg-accent text-sm">|</span>
            <span className="text-xs text-wg-body dark:text-wg-gray-light/50 font-mono">{rule_id}</span>
          </div>
          <p className="text-sm text-wg-body dark:text-wg-gray-light/70 mt-1">{description}</p>
        </div>
      </div>
    )
  }

  return (
    <div className={`wg-card p-5 rounded-xl border ${sev.border} ${sev.bg}`}>
      <div className="flex items-start gap-4">
        <div className="text-2xl mt-0.5">{sev.icon}</div>
        <div className="flex-1">
          <div className="flex items-center gap-3 flex-wrap">
            <span className="font-medium text-wg-headline dark:text-white">{ruleName}</span>
            <span className={`text-xs px-2 py-0.5 rounded ${sev.badge}`}>{t(`severity.${severity}`)}</span>
            <span className="text-xs px-2 py-0.5 rounded bg-wg-red/10 text-wg-red dark:bg-wg-red/20 dark:text-red-300">{t('status.failed')}</span>
            {points_deducted > 0 && (
              <span className="text-xs px-2 py-0.5 rounded bg-red-100 text-red-700 dark:bg-red-900/40 dark:text-red-300 font-bold border border-red-200 dark:border-red-800">
                {t('status.ptsDeducted', { points: points_deducted })}
              </span>
            )}
            <span className="wg-accent text-sm">|</span>
            <span className="text-xs text-wg-body dark:text-wg-gray-light/50 font-mono">{rule_id}</span>
          </div>
          <p className="text-sm text-wg-body dark:text-wg-gray-light/80 mt-2">{description}</p>
          <div className="mt-3 p-3 rounded-lg bg-white/60 dark:bg-wg-black/30 border-l-4 border-l-wg-blue dark:border-l-wg-blue">
            <p className="text-xs text-wg-blue dark:text-blue-300 font-semibold uppercase mb-1">{t('card.howToFix')}</p>
            <p className="text-sm text-wg-body dark:text-wg-gray-light/80">{remediation}</p>
          </div>
        </div>
      </div>
    </div>
  )
}

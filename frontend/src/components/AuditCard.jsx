import { useI18n } from '../i18n/I18nContext'

const severityStyles = {
  critical: {
    border: 'border-red-200 dark:border-red-800',
    bg: 'bg-red-50 dark:bg-red-900/20',
    badge: 'bg-red-100 text-red-700 dark:bg-red-900 dark:text-red-300',
  },
  high: {
    border: 'border-orange-200 dark:border-orange-800',
    bg: 'bg-orange-50 dark:bg-orange-900/20',
    badge: 'bg-orange-100 text-orange-700 dark:bg-orange-900 dark:text-orange-300',
  },
  medium: {
    border: 'border-yellow-200 dark:border-yellow-800',
    bg: 'bg-yellow-50 dark:bg-yellow-900/20',
    badge: 'bg-yellow-100 text-yellow-700 dark:bg-yellow-900 dark:text-yellow-300',
  },
}

export default function AuditCard({ result }) {
  const { t } = useI18n()
  const { passed, severity, rule_id, details } = result
  const sev = severityStyles[severity] || severityStyles.medium

  const ruleName = t(`rules.${rule_id}.name`)
  const description = passed
    ? t(`rules.${rule_id}.descPass`)
    : t(`rules.${rule_id}.descFail`, { details: details?.join(', ') || '' })
  const remediation = t(`rules.${rule_id}.remediation`)

  if (passed) {
    return (
      <div className="p-5 rounded-xl border border-gray-200 dark:border-gray-800 bg-white dark:bg-gray-900/50 flex items-start gap-4">
        <div className="text-green-500 text-2xl mt-0.5">&#10004;</div>
        <div>
          <div className="flex items-center gap-3">
            <span className="font-semibold text-gray-900 dark:text-white">{ruleName}</span>
            <span className="text-xs px-2 py-0.5 rounded bg-green-100 text-green-700 dark:bg-green-900 dark:text-green-300">{t('status.passed')}</span>
            <span className="text-xs text-gray-500">{rule_id}</span>
          </div>
          <p className="text-sm text-gray-500 dark:text-gray-400 mt-1">{description}</p>
        </div>
      </div>
    )
  }

  return (
    <div className={`p-5 rounded-xl border ${sev.border} ${sev.bg}`}>
      <div className="flex items-start gap-4">
        <div className="text-red-500 text-2xl mt-0.5">&#9888;</div>
        <div className="flex-1">
          <div className="flex items-center gap-3 flex-wrap">
            <span className="font-semibold text-gray-900 dark:text-white">{ruleName}</span>
            <span className={`text-xs px-2 py-0.5 rounded ${sev.badge}`}>{t(`severity.${severity}`)}</span>
            <span className="text-xs px-2 py-0.5 rounded bg-red-100 text-red-700 dark:bg-red-900 dark:text-red-300">{t('status.failed')}</span>
            <span className="text-xs text-gray-500">{rule_id}</span>
          </div>
          <p className="text-sm text-gray-700 dark:text-gray-300 mt-2">{description}</p>
          <div className="mt-3 p-3 rounded-lg bg-white/50 dark:bg-gray-900/50 border border-gray-200 dark:border-gray-700">
            <p className="text-xs text-gray-500 dark:text-gray-400 font-semibold uppercase mb-1">{t('card.howToFix')}</p>
            <p className="text-sm text-gray-700 dark:text-gray-300">{remediation}</p>
          </div>
        </div>
      </div>
    </div>
  )
}

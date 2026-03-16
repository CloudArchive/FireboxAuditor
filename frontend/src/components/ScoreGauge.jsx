import { useI18n } from '../i18n/I18nContext'

export default function ScoreGauge({ score }) {
  const { t } = useI18n()
  const radius = 90
  const stroke = 12
  const circumference = 2 * Math.PI * radius
  const progress = (score / 100) * circumference
  const size = (radius + stroke) * 2

  const color =
    score >= 71 ? '#22c55e' : score >= 41 ? '#eab308' : '#ef4444'
  const label =
    score >= 71 ? t('app.scoreGood') : score >= 41 ? t('app.scoreMedium') : t('app.scoreCritical')

  const bgLight =
    score >= 71
      ? 'from-green-50 to-green-50/30'
      : score >= 41
        ? 'from-yellow-50 to-yellow-50/30'
        : 'from-red-50 to-red-50/30'
  const bgDark =
    score >= 71
      ? 'dark:from-green-900/20 dark:to-green-900/5'
      : score >= 41
        ? 'dark:from-yellow-900/20 dark:to-yellow-900/5'
        : 'dark:from-red-900/20 dark:to-red-900/5'

  return (
    <div className={`flex flex-col items-center p-10 rounded-2xl bg-gradient-to-b ${bgLight} ${bgDark} border border-gray-200 dark:border-gray-800`}>
      <svg width={size} height={size} className="transform -rotate-90">
        <circle
          cx={radius + stroke}
          cy={radius + stroke}
          r={radius}
          fill="none"
          className="stroke-gray-200 dark:stroke-gray-800"
          strokeWidth={stroke}
        />
        <circle
          cx={radius + stroke}
          cy={radius + stroke}
          r={radius}
          fill="none"
          stroke={color}
          strokeWidth={stroke}
          strokeLinecap="round"
          strokeDasharray={circumference}
          strokeDashoffset={circumference - progress}
          style={{ transition: 'stroke-dashoffset 1s ease-out' }}
        />
      </svg>
      <div className="flex flex-col items-center -mt-[130px] mb-8">
        <span className="text-6xl font-bold text-gray-900 dark:text-white">{score}</span>
        <span className="text-sm font-medium mt-1" style={{ color }}>
          {label}
        </span>
        <span className="text-xs text-gray-500 mt-1">/ 100</span>
      </div>
      <p className="text-gray-500 dark:text-gray-400 text-sm text-center mt-2">{t('app.securityScore')}</p>
    </div>
  )
}

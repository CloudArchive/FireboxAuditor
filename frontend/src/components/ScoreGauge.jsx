import { useI18n } from '../i18n/I18nContext'

export default function ScoreGauge({ score }) {
  const { t } = useI18n()
  const radius = 90
  const stroke = 12
  const circumference = 2 * Math.PI * radius
  const progress = (score / 100) * circumference
  const size = (radius + stroke) * 2

  const color =
    score >= 71 ? '#22c55e' : score >= 41 ? '#eab308' : '#E81410'
  const label =
    score >= 71 ? t('app.scoreGood') : score >= 41 ? t('app.scoreMedium') : t('app.scoreCritical')

  return (
    <div
      className="wg-card flex flex-col items-center p-10 rounded-xl border border-wg-gray-light dark:border-wg-headline/40 bg-white dark:bg-wg-headline/15"
      id="score-gauge"
    >
      <svg width={size} height={size} className="transform -rotate-90">
        <circle
          cx={radius + stroke}
          cy={radius + stroke}
          r={radius}
          fill="none"
          className="stroke-wg-gray-light dark:stroke-wg-headline"
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
        <span className="text-6xl font-bold text-wg-headline dark:text-white">{score}</span>
        <span className="text-sm font-medium mt-1" style={{ color }}>
          {label}
        </span>
        <span className="text-xs text-wg-body mt-1">/ 100</span>
      </div>
      <p className="text-wg-body dark:text-wg-gray-light/60 text-sm text-center mt-2">{t('app.securityScore')}</p>
    </div>
  )
}

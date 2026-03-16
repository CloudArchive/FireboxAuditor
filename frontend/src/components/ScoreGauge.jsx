export default function ScoreGauge({ score }) {
  const radius = 90
  const stroke = 12
  const circumference = 2 * Math.PI * radius
  const progress = (score / 100) * circumference
  const size = (radius + stroke) * 2

  const color =
    score >= 71 ? '#22c55e' : score >= 41 ? '#eab308' : '#ef4444'
  const label =
    score >= 71 ? 'Iyi' : score >= 41 ? 'Orta' : 'Kritik'
  const bg =
    score >= 71
      ? 'from-green-900/20 to-green-900/5'
      : score >= 41
        ? 'from-yellow-900/20 to-yellow-900/5'
        : 'from-red-900/20 to-red-900/5'

  return (
    <div className={`flex flex-col items-center p-10 rounded-2xl bg-gradient-to-b ${bg} border border-gray-800`}>
      <svg width={size} height={size} className="transform -rotate-90">
        <circle
          cx={radius + stroke}
          cy={radius + stroke}
          r={radius}
          fill="none"
          stroke="#1f2937"
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
        <span className="text-6xl font-bold text-white">{score}</span>
        <span className="text-sm font-medium mt-1" style={{ color }}>
          {label}
        </span>
        <span className="text-xs text-gray-500 mt-1">/ 100</span>
      </div>
      <p className="text-gray-400 text-sm text-center mt-2">Guvenlik Skoru</p>
    </div>
  )
}

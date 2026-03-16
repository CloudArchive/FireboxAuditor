const severityConfig = {
  critical: { label: 'Kritik', color: 'text-red-400', border: 'border-red-800', bg: 'bg-red-900/20', badge: 'bg-red-900 text-red-300' },
  high:     { label: 'Yuksek', color: 'text-orange-400', border: 'border-orange-800', bg: 'bg-orange-900/20', badge: 'bg-orange-900 text-orange-300' },
  medium:   { label: 'Orta', color: 'text-yellow-400', border: 'border-yellow-800', bg: 'bg-yellow-900/20', badge: 'bg-yellow-900 text-yellow-300' },
}

export default function AuditCard({ result }) {
  const { passed, severity, name, rule_id, description, remediation } = result
  const sev = severityConfig[severity] || severityConfig.medium

  if (passed) {
    return (
      <div className="p-5 rounded-xl border border-gray-800 bg-gray-900/50 flex items-start gap-4">
        <div className="text-green-500 text-2xl mt-0.5">&#10004;</div>
        <div>
          <div className="flex items-center gap-3">
            <span className="font-semibold text-white">{name}</span>
            <span className="text-xs px-2 py-0.5 rounded bg-green-900 text-green-300">Gecti</span>
            <span className="text-xs text-gray-500">{rule_id}</span>
          </div>
          <p className="text-sm text-gray-400 mt-1">{description}</p>
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
            <span className="font-semibold text-white">{name}</span>
            <span className={`text-xs px-2 py-0.5 rounded ${sev.badge}`}>{sev.label}</span>
            <span className="text-xs px-2 py-0.5 rounded bg-red-900 text-red-300">Basarisiz</span>
            <span className="text-xs text-gray-500">{rule_id}</span>
          </div>
          <p className="text-sm text-gray-300 mt-2">{description}</p>
          <div className="mt-3 p-3 rounded-lg bg-gray-900/50 border border-gray-700">
            <p className="text-xs text-gray-400 font-semibold uppercase mb-1">Nasil Duzeltilir?</p>
            <p className="text-sm text-gray-300">{remediation}</p>
          </div>
        </div>
      </div>
    </div>
  )
}

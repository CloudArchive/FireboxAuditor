import { useI18n } from '../i18n/I18nContext'

export default function DeviceInfoCard({ info }) {
  const { t } = useI18n()

  if (!info || !info.model) return null

  const fields = [
    { label: t('device.model'), value: info.model },
    { label: t('device.serial'), value: info.serial_number || '-' },
    { label: t('device.firmware'), value: info.firmware_version },
    { label: t('device.systemName'), value: info.system_name },
    { label: t('device.domain'), value: info.domain_name || '-' },
    { label: t('device.contact'), value: info.contact || '-' },
    { label: t('device.location'), value: info.location || '-' },
    { label: t('device.dns'), value: (info.dns_servers || []).join(', ') || '-' },
  ]

  const typeColors = {
    External: 'bg-red-100 text-red-700 dark:bg-red-900/40 dark:text-red-300',
    Trusted: 'bg-green-100 text-green-700 dark:bg-green-900/40 dark:text-green-300',
    Optional: 'bg-blue-100 text-blue-700 dark:bg-blue-900/40 dark:text-blue-300',
    Mixed: 'bg-yellow-100 text-yellow-700 dark:bg-yellow-900/40 dark:text-yellow-300',
    Other: 'bg-gray-100 text-gray-700 dark:bg-gray-800 dark:text-gray-300',
  }

  return (
    <div className="rounded-2xl border border-gray-200 dark:border-gray-800 bg-white dark:bg-gray-900 overflow-hidden">
      <div className="px-6 py-4 border-b border-gray-100 dark:border-gray-800 flex items-center gap-3">
        <div className="w-8 h-8 rounded-lg bg-blue-600 flex items-center justify-center text-white text-sm font-bold">
          {info.model?.[0] || 'F'}
        </div>
        <div>
          <h3 className="font-semibold text-gray-900 dark:text-white">{t('device.title')}</h3>
          <p className="text-xs text-gray-500 dark:text-gray-400">WatchGuard {info.model} &mdash; {info.system_name}</p>
        </div>
      </div>

      <div className="px-6 py-4 grid grid-cols-2 sm:grid-cols-4 gap-4">
        {fields.map((f) => (
          <div key={f.label}>
            <p className="text-xs text-gray-500 dark:text-gray-400">{f.label}</p>
            <p className="text-sm font-medium text-gray-900 dark:text-white truncate" title={f.value}>{f.value}</p>
          </div>
        ))}
      </div>

      {info.interfaces?.length > 0 && (
        <div className="px-6 pb-4">
          <p className="text-xs text-gray-500 dark:text-gray-400 mb-2">{t('device.interfaces')}</p>
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead>
                <tr className="text-xs text-gray-500 dark:text-gray-400 border-b border-gray-100 dark:border-gray-800">
                  <th className="text-left py-1 pr-4 font-medium">{t('device.ifName')}</th>
                  <th className="text-left py-1 pr-4 font-medium">{t('device.ifType')}</th>
                  <th className="text-left py-1 pr-4 font-medium">{t('device.ifDevice')}</th>
                  <th className="text-left py-1 pr-4 font-medium">{t('device.ifIP')}</th>
                  <th className="text-left py-1 font-medium">{t('device.ifStatus')}</th>
                </tr>
              </thead>
              <tbody>
                {info.interfaces.map((iface, i) => (
                  <tr key={i} className="border-b border-gray-50 dark:border-gray-800/50 last:border-0">
                    <td className="py-1.5 pr-4 font-medium text-gray-900 dark:text-white">{iface.name}</td>
                    <td className="py-1.5 pr-4">
                      <span className={`inline-block px-2 py-0.5 rounded text-xs font-medium ${typeColors[iface.type] || typeColors.Other}`}>
                        {iface.type}
                      </span>
                    </td>
                    <td className="py-1.5 pr-4 text-gray-600 dark:text-gray-400">{iface.device}</td>
                    <td className="py-1.5 pr-4 text-gray-600 dark:text-gray-400 font-mono text-xs">
                      {iface.ip}{iface.ip !== 'DHCP' && iface.netmask ? ` / ${iface.netmask}` : ''}
                    </td>
                    <td className="py-1.5">
                      <span className={`inline-block w-2 h-2 rounded-full ${iface.enabled ? 'bg-green-500' : 'bg-gray-300 dark:bg-gray-600'}`} />
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      )}
    </div>
  )
}

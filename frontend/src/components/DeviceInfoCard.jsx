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
    External: 'bg-wg-red/10 text-wg-red dark:bg-wg-red/20 dark:text-red-300',
    Trusted: 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-300',
    Optional: 'bg-wg-blue/10 text-wg-blue dark:bg-wg-blue/20 dark:text-blue-300',
    Mixed: 'bg-yellow-100 text-yellow-700 dark:bg-yellow-900/30 dark:text-yellow-300',
    Other: 'bg-wg-gray-light text-wg-body dark:bg-wg-headline dark:text-wg-gray-light/70',
  }

  return (
    <div className="wg-card rounded-xl border border-wg-gray-light dark:border-wg-headline/40 bg-white dark:bg-wg-headline/15 overflow-hidden" id="device-info-card">
      <div className="px-6 py-4 border-b border-wg-gray-light dark:border-wg-headline/30 flex items-center gap-3">
        <div className="w-9 h-9 rounded-lg bg-wg-red flex items-center justify-center text-white text-sm font-bold shadow-sm">
          {info.model?.[0] || 'F'}
        </div>
        <div>
          <h3 className="font-medium text-wg-headline dark:text-white">
            <span className="wg-accent mr-1">&gt;</span>
            {t('device.title')}
          </h3>
          <p className="text-xs text-wg-body dark:text-wg-gray-light/50">WatchGuard {info.model} &mdash; {info.system_name}</p>
        </div>
      </div>

      <div className="px-6 py-4 grid grid-cols-2 sm:grid-cols-4 gap-4">
        {fields.map((f) => (
          <div key={f.label}>
            <p className="text-xs text-wg-body dark:text-wg-gray-light/50 mb-0.5">{f.label}</p>
            <p className="text-sm font-medium text-wg-headline dark:text-white truncate" title={f.value}>{f.value}</p>
          </div>
        ))}
      </div>

      {info.interfaces?.length > 0 && (
        <div className="px-6 pb-4">
          <p className="text-xs text-wg-body dark:text-wg-gray-light/50 mb-2 font-medium uppercase tracking-wide">{t('device.interfaces')}</p>
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead>
                <tr className="text-xs text-wg-headline dark:text-wg-gray-light/60 border-b border-wg-gray-light dark:border-wg-headline/30">
                  <th className="text-left py-2 pr-4 font-medium">{t('device.ifName')}</th>
                  <th className="text-left py-2 pr-4 font-medium">{t('device.ifType')}</th>
                  <th className="text-left py-2 pr-4 font-medium">{t('device.ifDevice')}</th>
                  <th className="text-left py-2 pr-4 font-medium">{t('device.ifIP')}</th>
                  <th className="text-left py-2 font-medium">{t('device.ifStatus')}</th>
                </tr>
              </thead>
              <tbody>
                {info.interfaces.map((iface, i) => (
                  <tr key={i} className="border-b border-wg-gray-light/50 dark:border-wg-headline/20 last:border-0">
                    <td className="py-2 pr-4 font-medium text-wg-headline dark:text-white">{iface.name}</td>
                    <td className="py-2 pr-4">
                      <span className={`inline-block px-2 py-0.5 rounded text-xs font-medium ${typeColors[iface.type] || typeColors.Other}`}>
                        {iface.type}
                      </span>
                    </td>
                    <td className="py-2 pr-4 text-wg-body dark:text-wg-gray-light/60">{iface.device}</td>
                    <td className="py-2 pr-4 text-wg-body dark:text-wg-gray-light/60 font-mono text-xs">
                      {iface.ip}{iface.ip !== 'DHCP' && iface.netmask ? ` / ${iface.netmask}` : ''}
                    </td>
                    <td className="py-2">
                      <span className={`inline-flex items-center gap-1.5 text-xs ${iface.enabled ? 'text-green-600 dark:text-green-400' : 'text-wg-body dark:text-wg-gray-light/40'}`}>
                        <span className={`inline-block w-2 h-2 rounded-full ${iface.enabled ? 'bg-green-500' : 'bg-wg-body/30 dark:bg-wg-gray-light/20'}`} />
                        {iface.enabled ? 'Active' : 'Off'}
                      </span>
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

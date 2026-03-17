import { useI18n } from '../i18n/I18nContext'

export default function PolicyTable({ policies, highlightedIndices = [], onSelectPolicy }) {
  const { t } = useI18n()

  const getActionIcon = (policy) => {
    if (policy.enabled === 'false' || policy.enabled === '0') return '⚪'
    return '✅'
  }

  return (
    <div className="mt-12 animate-fade-in" id="policy-visualization">
      <h2 className="text-xl font-semibold text-wg-headline dark:text-white mb-6 flex items-center gap-3">
        <span className="bg-wg-red w-1.5 h-6 rounded-full"></span>
        {t('policyTable.title') || 'Firewall Policy Visualization'}
      </h2>

      <div className="wg-card overflow-hidden border border-wg-gray-light dark:border-wg-headline/30 bg-white/80 dark:bg-wg-headline/10 backdrop-blur-md rounded-2xl shadow-xl shadow-wg-black/5">
        <div className="overflow-x-auto">
          <table className="w-full text-left border-collapse">
            <thead>
              <tr className="bg-wg-gray-light/30 dark:bg-wg-headline/40 border-b border-wg-gray-light dark:border-wg-headline/20">
                <th className="px-4 py-3 text-[10px] font-bold text-wg-body dark:text-wg-gray-light uppercase tracking-wider">{t('policyTable.order') || 'Order'}</th>
                <th className="px-4 py-3 text-[10px] font-bold text-wg-body dark:text-wg-gray-light uppercase tracking-wider">{t('policyTable.action') || 'Action'}</th>
                <th className="px-4 py-3 text-[10px] font-bold text-wg-body dark:text-wg-gray-light uppercase tracking-wider">{t('policyTable.name') || 'Policy Name'}</th>
                <th className="px-4 py-3 text-[10px] font-bold text-wg-body dark:text-wg-gray-light uppercase tracking-wider">{t('policyTable.type') || 'Type'}</th>
                <th className="px-4 py-3 text-[10px] font-bold text-wg-body dark:text-wg-gray-light uppercase tracking-wider">{t('policyTable.from') || 'From'}</th>
                <th className="px-4 py-3 text-[10px] font-bold text-wg-body dark:text-wg-gray-light uppercase tracking-wider">{t('policyTable.to') || 'To'}</th>
                <th className="px-4 py-3 text-[10px] font-bold text-wg-body dark:text-wg-gray-light uppercase tracking-wider">{t('policyTable.port') || 'Port / Service'}</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-wg-gray-light/10 dark:divide-wg-headline/20">
              {policies.map((policy) => {
                const isHighlighted = highlightedIndices.includes(policy.order)
                return (
                  <tr
                    key={policy.order}
                    id={`policy-row-${policy.order}`}
                    onClick={() => onSelectPolicy?.(policy)}
                    className={`transition-all duration-300 cursor-pointer ${isHighlighted ? 'bg-wg-red/10 dark:bg-wg-red/20 scale-[1.002] shadow-sm z-10' : 'hover:bg-wg-gray-light/5 dark:hover:bg-white/5'}`}
                  >
                    <td className="px-4 py-3.5 text-xs font-mono text-wg-body dark:text-wg-gray-light/60">{policy.order}</td>
                    <td className="px-4 py-3.5 text-lg">{getActionIcon(policy)}</td>
                    <td className="px-4 py-3.5">
                      <div className="flex flex-col">
                        <span className={`text-sm font-medium ${isHighlighted ? 'text-wg-red' : 'text-wg-headline dark:text-wg-gray-light'}`}>
                          {policy.name}
                        </span>
                        {policy.proxy && (
                          <span className="text-[10px] text-wg-blue dark:text-blue-400 font-semibold uppercase">{policy.proxy}</span>
                        )}
                      </div>
                    </td>
                    <td className="px-4 py-3.5 text-xs text-wg-body dark:text-wg-gray-light/70">{policy.type}</td>
                    <td className="px-4 py-3.5 text-xs text-wg-body dark:text-wg-gray-light/70">{policy.from?.aliases?.join(', ')}</td>
                    <td className="px-4 py-3.5 text-xs text-wg-body dark:text-wg-gray-light/70">{policy.to?.aliases?.join(', ')}</td>
                    <td className="px-4 py-3.5 text-xs font-mono text-wg-body dark:text-wg-gray-light/70">{policy.service}</td>
                  </tr>
                )
              })}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  )
}

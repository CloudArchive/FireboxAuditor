import { useState, useEffect } from 'react'
import { useI18n } from '../i18n/I18nContext'

const TABS = ['general', 'connections', 'logging', 'services']

export default function PolicyDetailPanel({ policy, aliases = [], onClose }) {
  const { t } = useI18n()
  const [activeTab, setActiveTab] = useState('general')

  const hasProxyServices = !!policy.proxy_services
  const visibleTabs = hasProxyServices ? TABS : TABS.filter(tab => tab !== 'services')

  // ESC to close + body scroll lock
  useEffect(() => {
    const onKey = (e) => { if (e.key === 'Escape') onClose() }
    document.addEventListener('keydown', onKey)
    document.body.classList.add('overflow-hidden')
    return () => {
      document.removeEventListener('keydown', onKey)
      document.body.classList.remove('overflow-hidden')
    }
  }, [onClose])

  const resolveAlias = (name) => {
    const alias = aliases.find(a => a.name === name)
    return alias?.members?.length ? alias.members : null
  }

  const tabLabel = (tab) => t(`policyDetail.tab.${tab}`)

  const isEnabled = policy.enabled !== 'false' && policy.enabled !== '0'

  return (
    <>
      {/* Backdrop */}
      <div
        className="fixed inset-0 bg-black/40 backdrop-blur-sm z-40"
        onClick={onClose}
      />

      {/* Panel */}
      <div className="fixed right-0 top-0 h-full w-[420px] max-w-full bg-white dark:bg-[#1a1a2e] border-l border-wg-gray-light dark:border-wg-headline/30 shadow-2xl z-50 flex flex-col animate-slide-in-right">
        {/* Header */}
        <div className="px-6 py-4 border-b border-wg-gray-light dark:border-wg-headline/20 flex items-start justify-between gap-3">
          <div className="min-w-0">
            <h3 className="text-lg font-semibold text-wg-headline dark:text-white truncate">
              {policy.name}
            </h3>
            <div className="flex items-center gap-2 mt-1">
              <span className={`inline-block w-2 h-2 rounded-full ${isEnabled ? 'bg-green-500' : 'bg-gray-400'}`} />
              <span className="text-xs text-wg-body dark:text-wg-gray-light/60">
                {isEnabled ? t('policyDetail.enabled') : t('policyDetail.disabled')}
              </span>
              <span className="text-xs text-wg-body dark:text-wg-gray-light/40">·</span>
              <span className="text-xs font-mono text-wg-body dark:text-wg-gray-light/60">#{policy.order}</span>
            </div>
          </div>
          <button
            onClick={onClose}
            className="p-1.5 rounded-lg hover:bg-wg-gray-light/20 dark:hover:bg-white/10 text-wg-body dark:text-wg-gray-light/60 transition shrink-0"
          >
            <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>

        {/* Tabs */}
        <div className="flex border-b border-wg-gray-light dark:border-wg-headline/20 px-6 gap-1">
          {visibleTabs.map(tab => (
            <button
              key={tab}
              onClick={() => setActiveTab(tab)}
              className={`px-3 py-2.5 text-xs font-semibold transition-colors relative ${
                activeTab === tab
                  ? 'text-wg-red'
                  : 'text-wg-body dark:text-wg-gray-light/50 hover:text-wg-headline dark:hover:text-white'
              }`}
            >
              {tabLabel(tab)}
              {activeTab === tab && (
                <span className="absolute bottom-0 left-0 right-0 h-0.5 bg-wg-red rounded-full" />
              )}
            </button>
          ))}
        </div>

        {/* Content */}
        <div className="flex-1 overflow-y-auto px-6 py-5 space-y-5">
          {activeTab === 'general' && <GeneralTab policy={policy} t={t} />}
          {activeTab === 'connections' && <ConnectionsTab policy={policy} resolveAlias={resolveAlias} t={t} />}
          {activeTab === 'logging' && <LoggingTab policy={policy} t={t} />}
          {activeTab === 'services' && hasProxyServices && <ServicesTab policy={policy} t={t} />}
        </div>
      </div>
    </>
  )
}

function Field({ label, value }) {
  return (
    <div>
      <dt className="text-[10px] font-bold text-wg-body dark:text-wg-gray-light/50 uppercase tracking-wider mb-1">{label}</dt>
      <dd className="text-sm text-wg-headline dark:text-wg-gray-light">{value || '—'}</dd>
    </div>
  )
}

function StatusBadge({ active, label }) {
  return (
    <span className={`inline-flex items-center gap-1.5 px-2.5 py-1 rounded-full text-xs font-medium ${
      active
        ? 'bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-400'
        : 'bg-red-100 text-red-800 dark:bg-red-900/30 dark:text-red-400'
    }`}>
      <span className={`w-1.5 h-1.5 rounded-full ${active ? 'bg-green-500' : 'bg-red-500'}`} />
      {label}
    </span>
  )
}

function GeneralTab({ policy, t }) {
  return (
    <dl className="grid grid-cols-2 gap-4">
      <Field label={t('policyDetail.name')} value={policy.name} />
      <Field label={t('policyDetail.type')} value={policy.type} />
      <Field label={t('policyDetail.service')} value={policy.service} />
      <Field label={t('policyDetail.proxy')} value={policy.proxy} />
      <div className="col-span-2">
        <Field label={t('policyDetail.description')} value={policy.description} />
      </div>
      {policy.schedule && (
        <Field label={t('policyDetail.schedule')} value={policy.schedule} />
      )}
      {policy.nat && (
        <div className="col-span-2">
          <Field
            label="NAT"
            value={[policy.nat.dynamic && `Dynamic: ${policy.nat.dynamic}`, policy.nat.static && `Static: ${policy.nat.static}`].filter(Boolean).join(' · ') || '—'}
          />
        </div>
      )}
    </dl>
  )
}

function AliasMembers({ aliases, resolveAlias, t }) {
  if (!aliases?.length) return <span className="text-sm text-wg-body dark:text-wg-gray-light/50">—</span>

  return (
    <div className="space-y-2">
      {aliases.map(name => {
        const members = resolveAlias(name)
        return (
          <div key={name}>
            <span className="text-sm font-medium text-wg-headline dark:text-wg-gray-light">{name}</span>
            {members ? (
              <ul className="ml-3 mt-1 space-y-0.5">
                {members.map((m, i) => (
                  <li key={i} className="text-xs text-wg-body dark:text-wg-gray-light/60 flex items-center gap-1.5">
                    <span className="w-1 h-1 rounded-full bg-wg-blue shrink-0" />
                    {m}
                  </li>
                ))}
              </ul>
            ) : (
              <span className="ml-2 text-xs text-wg-body dark:text-wg-gray-light/40">—</span>
            )}
          </div>
        )
      })}
    </div>
  )
}

function ConnectionsTab({ policy, resolveAlias, t }) {
  return (
    <div className="space-y-5">
      <div>
        <h4 className="text-[10px] font-bold text-wg-body dark:text-wg-gray-light/50 uppercase tracking-wider mb-2">
          {t('policyDetail.from')}
        </h4>
        <AliasMembers aliases={policy.from?.aliases} resolveAlias={resolveAlias} t={t} />
      </div>
      <div>
        <h4 className="text-[10px] font-bold text-wg-body dark:text-wg-gray-light/50 uppercase tracking-wider mb-2">
          {t('policyDetail.to')}
        </h4>
        <AliasMembers aliases={policy.to?.aliases} resolveAlias={resolveAlias} t={t} />
      </div>
    </div>
  )
}

function LoggingTab({ policy, t }) {
  const log = policy.logging || {}
  const boolVal = (v) => v === 'true' || v === '1'

  return (
    <div className="space-y-3">
      <div className="flex items-center justify-between">
        <span className="text-sm text-wg-headline dark:text-wg-gray-light">{t('policyDetail.logEnabled')}</span>
        <StatusBadge active={boolVal(log.enabled)} label={boolVal(log.enabled) ? t('policyDetail.on') : t('policyDetail.off')} />
      </div>
      <div className="flex items-center justify-between">
        <span className="text-sm text-wg-headline dark:text-wg-gray-light">{t('policyDetail.logMessage')}</span>
        <StatusBadge active={boolVal(log.log_message)} label={boolVal(log.log_message) ? t('policyDetail.on') : t('policyDetail.off')} />
      </div>
      <div className="flex items-center justify-between">
        <span className="text-sm text-wg-headline dark:text-wg-gray-light">{t('policyDetail.logReport')}</span>
        <StatusBadge active={boolVal(log.for_report)} label={boolVal(log.for_report) ? t('policyDetail.on') : t('policyDetail.off')} />
      </div>
    </div>
  )
}

function ServicesTab({ policy, t }) {
  const ps = policy.proxy_services
  const items = [
    { label: 'Gateway AntiVirus', active: ps.gateway_av },
    { label: 'IPS', active: ps.ips },
    { label: 'WebBlocker', active: ps.web_blocker },
    { label: 'APT Blocker', active: ps.apt_blocker },
  ]

  return (
    <div className="space-y-3">
      <p className="text-xs text-wg-body dark:text-wg-gray-light/50 mb-2">
        {t('policyDetail.servicesDesc')}
      </p>
      {items.map(item => (
        <div key={item.label} className="flex items-center justify-between">
          <span className="text-sm text-wg-headline dark:text-wg-gray-light">{item.label}</span>
          <StatusBadge active={item.active} label={item.active ? t('policyDetail.on') : t('policyDetail.off')} />
        </div>
      ))}
    </div>
  )
}

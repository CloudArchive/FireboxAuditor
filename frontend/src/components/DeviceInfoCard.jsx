import { useState } from 'react'
import { useI18n } from '../i18n/I18nContext'

/* ── Small helpers ──────────────────────────────────────────────────────────── */

function Row({ label, value, mono = false }) {
  return (
    <div className="flex items-start justify-between gap-4 py-2 border-b border-wg-gray-light/50 dark:border-wg-headline/20 last:border-0">
      <span className="text-xs text-wg-body dark:text-wg-gray-light/50 shrink-0 pt-0.5">{label}</span>
      <span className={`text-xs text-right text-wg-headline dark:text-white break-all ${mono ? 'font-mono' : 'font-medium'}`}>
        {value || '—'}
      </span>
    </div>
  )
}

function SectionHeader({ icon, title, badge, rightElement }) {
  return (
    <div className="flex items-center justify-between mb-3 w-full">
      <h3 className="text-xs font-bold text-wg-headline dark:text-white uppercase tracking-wider flex items-center gap-1.5">
        <span>{icon}</span> {title}
      </h3>
      <div className="flex items-center gap-2">
        {badge}
        {rightElement}
      </div>
    </div>
  )
}

function ConnectedBadge({ label, host, onReconnect, onDisconnect, onShowLogs }) {
  return (
    <div className="flex items-center gap-1.5">
      <span className="text-[10px] px-2 py-0.5 rounded-full bg-emerald-500/10 dark:bg-emerald-500/20 text-emerald-600 dark:text-emerald-400 font-medium border border-emerald-500/20">
        ✓ {label}{host ? ` · ${host}` : ''}
      </span>
      {onShowLogs && (
        <button
          onClick={(e) => { e.stopPropagation(); onShowLogs(); }}
          title="SSH loglarını göster"
          className="text-[10px] px-1.5 py-0.5 rounded bg-wg-headline/10 dark:bg-white/10 text-wg-headline dark:text-white/60 hover:bg-wg-headline/20 dark:hover:bg-white/20 transition font-mono border border-wg-headline/20 dark:border-white/10"
        >
          &gt;_
        </button>
      )}
      {onReconnect && (
        <button
          onClick={(e) => { e.stopPropagation(); onReconnect(); }}
          title="Yeniden bağlan"
          className="text-[10px] px-1.5 py-0.5 rounded bg-wg-blue/10 text-wg-blue hover:bg-wg-blue/20 transition font-medium border border-wg-blue/20"
        >
          ↺
        </button>
      )}
      {onDisconnect && (
        <button
          onClick={(e) => { e.stopPropagation(); onDisconnect(); }}
          title="Bağlantıyı kes"
          className="text-[10px] px-1.5 py-0.5 rounded bg-wg-red/10 text-wg-red hover:bg-wg-red/20 transition font-medium border border-wg-red/20"
        >
          ✕
        </button>
      )}
    </div>
  )
}

function BlurOverlay({ label, onEnrich }) {
  return (
    <div className="relative">
      {/* Blurred fake rows */}
      <div className="blur-sm pointer-events-none select-none space-y-2 py-1">
        {[80, 60, 70].map((w, i) => (
          <div key={i} className={`h-3 rounded bg-wg-gray-light dark:bg-wg-headline/40`} style={{ width: `${w}%` }} />
        ))}
      </div>
      {/* CTA overlay */}
      <div className="absolute inset-0 flex items-center justify-center">
        <button
          onClick={(e) => { e.stopPropagation(); onEnrich(); }}
          className="px-4 py-2 rounded-lg bg-wg-red hover:bg-wg-red-hover text-white text-xs font-semibold shadow-lg transition active:scale-95"
        >
          🔑 {label}
        </button>
      </div>
    </div>
  )
}

/* ── Feature Key Section ─────────────────────────────────────────────────────── */

// Fallback for older show feature-key format if show features isn't completely parsed
function FeatureKeySection({ featureKey, t }) {
  if (!featureKey || !featureKey.features?.length) return null

  // Key features to highlight
  const highlight = ['LiveSecurity', 'Gateway AntiVirus', 'WebBlocker', 'Intrusion Prevention', 'APT Blocker', 'Application Control']

  const featured = featureKey.features.filter(f =>
    highlight.some(h => f.name?.toLowerCase().includes(h.toLowerCase()))
  )
  const others = featureKey.features.filter(f =>
    !highlight.some(h => f.name?.toLowerCase().includes(h.toLowerCase()))
  )

  const renderFeature = (f, i) => (
    <div key={i} className="flex items-center justify-between py-1.5 border-b border-wg-gray-light/50 dark:border-wg-headline/20 last:border-0">
      <span className="text-xs text-wg-body dark:text-wg-gray-light/60 truncate pr-2">{f.name}</span>
      <div className="flex items-center gap-2 shrink-0">
        {f.expiration && (
          <span className="text-[10px] text-wg-body dark:text-wg-gray-light/40">
            {t('device.licenseExpiry')}: {f.expiration}
          </span>
        )}
        <span className={`text-[10px] font-semibold px-1.5 py-0.5 rounded ${
          f.active
            ? 'bg-emerald-500/10 text-emerald-600 dark:text-emerald-400'
            : 'bg-wg-red/10 text-wg-red'
        }`}>
          {f.active ? t('device.licenseActive') : t('device.licenseExpired')}
        </span>
      </div>
    </div>
  )

  return (
    <div className="space-y-0">
      {featured.map(renderFeature)}
      {others.length > 0 && (
        <details className="mt-1">
          <summary className="text-[10px] text-wg-body dark:text-wg-gray-light/40 cursor-pointer hover:text-wg-red select-none py-1">
            +{others.length} {t('device.licenseUnlicensed')}...
          </summary>
          <div className="mt-1">{others.map(renderFeature)}</div>
        </details>
      )}
    </div>
  )
}

// New table layout for "show features"
function LicensedFeaturesTable({ features, t }) {
  if (!features || !features.length) return null
  return (
    <div className="overflow-x-auto w-full">
      <table className="w-full text-[11px]">
        <thead>
          <tr className="text-wg-body dark:text-wg-gray-light/40 text-left border-b border-wg-gray-light/30 dark:border-wg-headline/20">
            <th className="pb-1 pr-3 font-medium">Feature</th>
            <th className="pb-1 pr-3 font-medium">Capacity</th>
            <th className="pb-1 pr-3 font-medium">Status</th>
            <th className="pb-1 font-medium">Expiration</th>
          </tr>
        </thead>
        <tbody className="divide-y divide-wg-gray-light/30 dark:divide-wg-headline/20">
          {features.map((f, i) => {
            const isActive = f.status?.toLowerCase() === 'enabled'
            return (
              <tr key={i} className="text-wg-headline dark:text-wg-gray-light/80 hover:bg-wg-gray-light/10 dark:hover:bg-wg-headline/30 transition-colors">
                <td className="py-1.5 pr-3 font-medium">{f.name}</td>
                <td className="py-1.5 pr-3 font-mono">{f.capacity}</td>
                <td className="py-1.5 pr-3">
                  <span className={`inline-block px-1.5 py-0.5 rounded text-[10px] font-semibold ${isActive ? 'bg-emerald-500/10 text-emerald-600 dark:text-emerald-400' : 'bg-wg-body/10 text-wg-body dark:bg-white/10 dark:text-white/60'}`}>
                    {f.status}
                  </span>
                </td>
                <td className="py-1.5 font-mono text-[10px]">{f.expiration}</td>
              </tr>
            )
          })}
        </tbody>
      </table>
    </div>
  )
}

/* ── Main Component ──────────────────────────────────────────────────────────── */

export default function DeviceInfoCard({ info, enrichment, onEnrichRequest, onReconnect, onDisconnect, onShowLogs }) {
  const { t } = useI18n()
  const hasEnrich = !!enrichment
  const [isLicenseOpen, setIsLicenseOpen] = useState(false)

  return (
    <div className="rounded-2xl border border-wg-gray-light dark:border-wg-headline/30 bg-white dark:bg-wg-headline/10 overflow-hidden shadow-sm flex flex-col">

      {/* ── Section 1: Device Identity (from XML) ───────────────────── */}
      <div className="px-5 py-4 border-b border-wg-gray-light dark:border-wg-headline/20">
        <SectionHeader icon="🔷" title={t('device.identitySection')} />
        <div className="space-y-0">
          {info?.model        && <Row label={t('device.model')}      value={info.model} />}
          {info?.firmware_version && <Row label={t('device.firmware')} value={enrichment?.full_version || info.firmware_version} mono />}
          {info?.system_name  && <Row label={t('device.systemName')} value={info.system_name} />}
          {info?.domain_name  && <Row label={t('device.domain')}     value={info.domain_name} />}
          {info?.contact      && <Row label={t('device.contact')}    value={info.contact} />}
          {info?.location     && <Row label={t('device.location')}   value={info.location} />}
          {info?.time_zone    && <Row label={t('device.timezone')}   value={info.time_zone} />}
          {info?.dns_servers?.length > 0 && (
            <Row label={t('device.dns')} value={info.dns_servers.join(', ')} mono />
          )}
          {info?.log_server   && <Row label={t('device.logServer')}  value={info.log_server} mono />}
          {info?.syslog_server && <Row label={t('device.syslogServer')} value={info.syslog_server} mono />}
          {info?.dnswatch     && <Row label={t('device.dnswatch')}   value={info.dnswatch} />}
        </div>

        {/* Interface table */}
        {info?.interfaces?.length > 0 && (
          <div className="mt-3">
            <p className="text-[10px] font-bold text-wg-body dark:text-wg-gray-light/40 uppercase tracking-wider mb-2">
              {t('device.interfaces')}
            </p>
            <div className="overflow-x-auto">
              <table className="w-full text-[11px]">
                <thead>
                  <tr className="text-wg-body dark:text-wg-gray-light/40 text-left">
                    <th className="pb-1 pr-3 font-medium">{t('device.ifName')}</th>
                    <th className="pb-1 pr-3 font-medium">{t('device.ifType')}</th>
                    <th className="pb-1 pr-3 font-medium">{t('device.ifIP')}</th>
                    <th className="pb-1 font-medium">{t('device.ifStatus')}</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-wg-gray-light/30 dark:divide-wg-headline/20">
                  {info.interfaces.map((iface, i) => (
                    <tr key={i} className="text-wg-headline dark:text-wg-gray-light/80">
                      <td className="py-1 pr-3 font-mono">{iface.name}</td>
                      <td className="py-1 pr-3">{iface.type}</td>
                      <td className="py-1 pr-3 font-mono">{iface.ip || '—'}</td>
                      <td className="py-1">
                        <span className={`inline-block w-1.5 h-1.5 rounded-full mr-1 ${iface.enabled ? 'bg-emerald-500' : 'bg-wg-body/30'}`} />
                        {iface.enabled ? 'UP' : 'DOWN'}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>
        )}
      </div>

      {/* ── Section 2: Live Data (from SSH) ─────────────────────────── */}
      <div className="px-5 py-4 border-b border-wg-gray-light dark:border-wg-headline/20">
        <SectionHeader
          icon="📡"
          title={t('device.liveSection')}
          badge={hasEnrich
            ? <ConnectedBadge
                label={t('device.sshConnected')}
                host={enrichment.ssh_host}
                onShowLogs={onShowLogs}
                onReconnect={onReconnect}
                onDisconnect={onDisconnect}
              />
            : <span className="text-[10px] text-wg-body dark:text-wg-gray-light/40">{t('device.sshPending')}</span>
          }
        />
        {hasEnrich ? (
          <div className="space-y-0">
            <Row label={t('device.serial')}  value={enrichment.serial_number} mono />
            <Row label={t('device.uptime')}  value={enrichment.up_time} />
            <Row label={t('device.cpu')}     value={enrichment.cpu_usage} />
            <Row label={t('device.memory')}  value={enrichment.memory_usage} />
          </div>
        ) : (
          <BlurOverlay label={t('device.enrichCta')} onEnrich={onEnrichRequest} />
        )}
      </div>

      {/* ── Section 3: License (Feature Key) ────────────────────────── */}
      <div 
        className={`px-5 py-4 transition-colors ${hasEnrich ? 'cursor-pointer hover:bg-wg-gray-light/20 dark:hover:bg-white/5' : ''}`}
        onClick={() => { if (hasEnrich) setIsLicenseOpen(!isLicenseOpen); }}
      >
        <SectionHeader
          icon="🔑"
          title={t('device.licenseSection')}
          badge={hasEnrich
            ? <ConnectedBadge label={t('device.sshConnected')} host={enrichment.ssh_host} />
            : null
          }
          rightElement={hasEnrich ? (
            <span className="text-wg-body dark:text-wg-gray-light/50 text-xs ml-2">
              {isLicenseOpen ? '▲' : '▼'}
            </span>
          ) : null}
        />
        
        {hasEnrich ? (
          isLicenseOpen && (
            <div className="mt-4 pt-4 border-t border-wg-gray-light dark:border-wg-headline/20 cursor-auto" onClick={(e) => e.stopPropagation()}>
              {enrichment.features?.length > 0 ? (
                <LicensedFeaturesTable features={enrichment.features} t={t} />
              ) : enrichment.feature_key ? (
                <FeatureKeySection featureKey={enrichment.feature_key} t={t} />
              ) : (
                <div className="text-sm text-wg-body">No feature key available.</div>
              )}
            </div>
          )
        ) : (
          <div className="cursor-auto" onClick={(e) => e.stopPropagation()}>
            <BlurOverlay label={t('device.enrichCta')} onEnrich={onEnrichRequest} />
          </div>
        )}
      </div>
    </div>
  )
}

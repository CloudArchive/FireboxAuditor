import { useState, useEffect } from 'react'
import { I18nProvider } from './i18n/I18nContext'
import { AuthProvider, useAuth } from './contexts/AuthContext'
import LoginPage from './pages/LoginPage'
import DashboardPage from './pages/DashboardPage'
import AuditPage from './pages/AuditPage'

/* ── Inner app — uses auth context ─────────────────────────────────────────── */

function InnerApp() {
  const { isLoggedIn, apiFetch } = useAuth()

  // page: 'dashboard' | 'audit'
  const [page, setPage]               = useState('dashboard')
  const [history, setHistory]         = useState([])
  const [activeRecord, setActiveRecord] = useState(null)
  const [loading, setLoading]         = useState(false)
  const [historyLoading, setHistLoading] = useState(false)

  // Load history on mount / login
  useEffect(() => {
    if (!isLoggedIn) return
    loadHistory()
  }, [isLoggedIn])

  const loadHistory = async () => {
    setHistLoading(true)
    try {
      const resp = await apiFetch('/api/history')
      if (resp.ok) {
        const data = await resp.json()
        setHistory(data || [])
      }
    } finally {
      setHistLoading(false)
    }
  }

  // Upload new XML and go to audit page
  const handleNewAudit = async (fetchFn) => {
    setLoading(true)
    try {
      const data = await fetchFn() // { id, report }
      // Build a minimal record to show immediately
      const record = {
        id:          data.id,
        created_at:  new Date().toISOString(),
        file_name:   '—',
        device_name: data.report?.device_info?.system_name || 'Firebox',
        score:       data.report?.score ?? 0,
        report:      data.report,
        enrichment:  null,
      }
      setActiveRecord(record)
      setPage('audit')
      // Reload history in background
      loadHistory()
    } catch (err) {
      console.error('Upload failed:', err)
      alert(err.message)
    } finally {
      setLoading(false)
    }
  }

  // View an existing audit from history
  const handleViewAudit = async (id) => {
    try {
      const resp = await apiFetch(`/api/history/${id}`)
      if (!resp.ok) return
      const record = await resp.json()
      setActiveRecord(record)
      setPage('audit')
    } catch (err) {
      console.error(err)
    }
  }

  const handleDeleteAudit = async (id) => {
    try {
      await apiFetch(`/api/history/${id}`, { method: 'DELETE' })
      setHistory(prev => prev.filter(r => r.id !== id))
    } catch (err) {
      console.error(err)
    }
  }

  const handleRecordUpdate = (updated) => {
    setActiveRecord(updated)
    // Also update in history list if present
    setHistory(prev => prev.map(r => r.id === updated.id ? { ...r, enriched: !!updated.enrichment } : r))
  }

  const handleBack = () => {
    setPage('dashboard')
    setActiveRecord(null)
    loadHistory()
  }

  // Not logged in → show login
  if (!isLoggedIn) return <LoginPage />

  // Audit detail page
  if (page === 'audit' && activeRecord) {
    return (
      <AuditPage
        auditRecord={activeRecord}
        onBack={handleBack}
        onRecordUpdate={handleRecordUpdate}
      />
    )
  }

  // Dashboard
  return (
    <DashboardPage
      history={history}
      loading={loading || historyLoading}
      onView={handleViewAudit}
      onDelete={handleDeleteAudit}
      onNewAudit={handleNewAudit}
    />
  )
}

/* ── Root — provides contexts ───────────────────────────────────────────────── */

export default function App() {
  return (
    <I18nProvider>
      <AuthProvider>
        <InnerApp />
      </AuthProvider>
    </I18nProvider>
  )
}

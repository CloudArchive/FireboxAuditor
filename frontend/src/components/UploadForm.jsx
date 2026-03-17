import { useState, useRef } from 'react'
import { useI18n } from '../i18n/I18nContext'

export default function UploadForm({ onSubmit, loading }) {
  const { t } = useI18n()
  const [fileName, setFileName] = useState('')
  const inputRef = useRef()

  const handleSubmit = (e) => {
    e.preventDefault()
    const file = inputRef.current?.files?.[0]
    if (!file) return
    const formData = new FormData()
    formData.append('config', file)
    onSubmit(() =>
      fetch('/api/audit/upload', { method: 'POST', body: formData })
    )
  }

  return (
    <form onSubmit={handleSubmit} className="wg-card p-8 rounded-xl border border-wg-gray-light dark:border-wg-headline/50 bg-white dark:bg-wg-headline/20 wg-concrete space-y-5" id="upload-form">
      <h3 className="text-lg font-medium text-wg-headline dark:text-white">
        <span className="wg-accent mr-1">&gt;</span>
        {t('upload.title')}
      </h3>
      <div
        onClick={() => inputRef.current?.click()}
        className="border-2 border-dashed border-wg-gray-light dark:border-wg-headline rounded-xl p-10 text-center cursor-pointer hover:border-wg-red/50 dark:hover:border-wg-red/40 transition-colors duration-200 group"
      >
        <input
          ref={inputRef}
          type="file"
          accept=".xml"
          className="hidden"
          onChange={(e) => setFileName(e.target.files?.[0]?.name || '')}
        />
        {fileName ? (
          <div>
            <p className="text-wg-red font-medium text-lg">{fileName}</p>
            <p className="text-wg-body text-xs mt-1">{t('upload.dropzoneHint')}</p>
          </div>
        ) : (
          <div>
            <div className="text-4xl mb-3 opacity-60 group-hover:opacity-100 transition-opacity">📂</div>
            <p className="text-wg-body dark:text-wg-gray-light/70 font-medium">{t('upload.dropzone')}</p>
            <p className="text-xs text-wg-body/60 dark:text-wg-gray-light/40 mt-1">{t('upload.dropzoneHint')}</p>
          </div>
        )}
      </div>
      <button
        type="submit"
        disabled={loading || !fileName}
        className="w-full py-3 rounded-md bg-wg-red hover:bg-wg-red-hover text-white font-semibold transition-colors duration-200 disabled:opacity-50 disabled:cursor-not-allowed"
        id="upload-submit"
      >
        {loading ? t('upload.loading') : t('upload.submit')}
      </button>
    </form>
  )
}

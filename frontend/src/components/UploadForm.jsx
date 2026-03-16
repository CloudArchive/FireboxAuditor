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
    <form onSubmit={handleSubmit} className="p-8 rounded-2xl border border-gray-200 dark:border-gray-800 bg-white dark:bg-gray-900 space-y-5">
      <h3 className="text-lg font-semibold text-gray-900 dark:text-white">{t('upload.title')}</h3>
      <div
        onClick={() => inputRef.current?.click()}
        className="border-2 border-dashed border-gray-300 dark:border-gray-700 rounded-xl p-10 text-center cursor-pointer hover:border-blue-500 transition"
      >
        <input
          ref={inputRef}
          type="file"
          accept=".xml"
          className="hidden"
          onChange={(e) => setFileName(e.target.files?.[0]?.name || '')}
        />
        {fileName ? (
          <p className="text-blue-500 dark:text-blue-400 font-medium">{fileName}</p>
        ) : (
          <div>
            <p className="text-gray-500 dark:text-gray-400">{t('upload.dropzone')}</p>
            <p className="text-xs text-gray-400 dark:text-gray-600 mt-1">{t('upload.dropzoneHint')}</p>
          </div>
        )}
      </div>
      <button
        type="submit"
        disabled={loading || !fileName}
        className="w-full py-3 rounded-lg bg-blue-600 hover:bg-blue-500 text-white font-semibold transition disabled:opacity-50"
      >
        {loading ? t('upload.loading') : t('upload.submit')}
      </button>
    </form>
  )
}

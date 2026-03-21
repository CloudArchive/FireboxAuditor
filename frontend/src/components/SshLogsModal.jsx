export default function SshLogsModal({ logs, host, onClose }) {
  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/60 backdrop-blur-sm animate-fade-in">
      <div className="w-full max-w-2xl rounded-2xl border border-wg-headline/50 bg-[#0d1117] shadow-2xl animate-slide-up flex flex-col" style={{ maxHeight: '80vh' }}>

        {/* Header */}
        <div className="px-5 py-3 border-b border-white/10 flex items-center justify-between shrink-0">
          <div className="flex items-center gap-2">
            <span className="text-base">🖥</span>
            <span className="text-sm font-semibold text-white">SSH Terminal — {host}</span>
          </div>
          <button
            onClick={onClose}
            className="text-white/40 hover:text-white text-lg leading-none transition px-1"
            title="Kapat"
          >
            ✕
          </button>
        </div>

        {/* Log output */}
        <div className="flex-1 overflow-y-auto px-4 py-3 font-mono text-[11px] text-emerald-400 space-y-0.5">
          {logs.length === 0 ? (
            <p className="text-white/40 italic mt-2 text-center">Bellek yenilendiği için geçmiş loglar silindi.<br/>Canlı akışı görmek için <b>Yeniden Bağlan</b> butonunu kullanabilirsiniz.</p>
          ) : (
            logs.map((line, i) => (
              <div key={i} className={
                line.startsWith('[RAW') ? 'text-amber-400 whitespace-pre-wrap mt-2' :
                line.startsWith('[PARSED]') ? 'text-sky-400' :
                line.startsWith('[WARN]') ? 'text-red-400' :
                'text-emerald-400'
              }>
                {line}
              </div>
            ))
          )}
        </div>

        <div className="px-5 py-3 border-t border-white/10 shrink-0">
          <button
            onClick={onClose}
            className="w-full py-2 rounded-md bg-white/5 hover:bg-white/10 text-white/60 hover:text-white text-xs font-medium transition"
          >
            Kapat
          </button>
        </div>
      </div>
    </div>
  )
}

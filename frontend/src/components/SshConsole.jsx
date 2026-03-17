import { useEffect, useRef } from 'react'

export default function SshConsole({ logs, visible }) {
  const scrollRef = useRef(null)

  useEffect(() => {
    if (scrollRef.current) {
      scrollRef.current.scrollTop = scrollRef.current.scrollHeight
    }
  }, [logs])

  if (!visible || !logs || logs.length === 0) return null

  return (
    <div className="mt-4 p-4 rounded-md bg-black text-green-400 font-mono text-xs overflow-hidden border border-wg-headline/50 shadow-2xl">
      <div className="flex items-center justify-between mb-2 border-b border-wg-headline/30 pb-1">
        <span className="text-[10px] uppercase tracking-widest text-wg-gray-light/50">SSH Debug Console</span>
        <div className="flex gap-1.5">
          <div className="w-2.5 h-2.5 rounded-full bg-red-500/50"></div>
          <div className="w-2.5 h-2.5 rounded-full bg-yellow-500/50"></div>
          <div className="w-2.5 h-2.5 rounded-full bg-green-500/50"></div>
        </div>
      </div>
      <div 
        ref={scrollRef}
        className="max-h-48 overflow-y-auto space-y-1 scrollbar-thin scrollbar-thumb-wg-headline scrollbar-track-transparent"
      >
        {logs.map((log, i) => (
          <div key={i} className="flex gap-2">
            <span className="opacity-50 select-none">[{i + 1}]</span>
            <span className="break-all whitespace-pre-wrap">{log}</span>
          </div>
        ))}
      </div>
    </div>
  )
}

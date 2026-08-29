import { useEffect, useState } from 'react'
import { apiBaseUrl, fetchServiceGreeting } from './api'

type ServiceStatus =
  | { state: 'checking' }
  | { state: 'reachable'; greeting: string }
  | { state: 'unreachable'; reason: string }

function describe(status: ServiceStatus): string {
  switch (status.state) {
    case 'checking':
      return 'Checking…'
    case 'reachable':
      return status.greeting
    case 'unreachable':
      return `Unreachable — ${status.reason}`
  }
}

export default function App() {
  const [status, setStatus] = useState<ServiceStatus>({ state: 'checking' })

  useEffect(() => {
    const controller = new AbortController()

    fetchServiceGreeting(controller.signal)
      .then((greeting) => setStatus({ state: 'reachable', greeting }))
      .catch((error: unknown) => {
        if (controller.signal.aborted) return
        setStatus({
          state: 'unreachable',
          reason: error instanceof Error ? error.message : String(error),
        })
      })

    return () => controller.abort()
  }, [])

  return (
    <main className="min-h-dvh bg-slate-50 px-4 py-6 text-slate-900">
      <div className="mx-auto flex max-w-sm flex-col gap-4">
        <h1 className="text-2xl font-semibold tracking-tight">TurfGPS</h1>
        <section className="rounded-xl border border-slate-200 bg-white p-4 shadow-sm">
          <h2 className="text-sm font-medium text-slate-500">Service</h2>
          <p className="mt-1 text-base">{describe(status)}</p>
          <p className="mt-2 break-all font-mono text-xs text-slate-500">{apiBaseUrl}</p>
        </section>
      </div>
    </main>
  )
}

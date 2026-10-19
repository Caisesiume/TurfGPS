import { useEffect, useState } from 'react'
import { apiBaseUrl, fetchServiceGreeting } from './api'

type ServiceStatus =
  | { state: 'checking' }
  | { state: 'reachable'; greeting: string }
  | { state: 'unreachable'; reason: string }

// What the card says in each state. A browser reports a refused cross-origin
// answer and a dead network the same way, so the unreachable state names no
// cause; the error itself is shown beneath it as detail, not as the message.
function describe(status: ServiceStatus): string {
  switch (status.state) {
    case 'checking':
      return 'Checking the service…'
    case 'reachable':
      return status.greeting
    case 'unreachable':
      return 'The service could not be reached.'
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
        <section aria-labelledby="service-heading" className="rounded-xl border border-slate-200 bg-white p-4 shadow-sm">
          <h2 id="service-heading" className="text-sm font-medium text-slate-600">Service</h2>
          <p role="status" className="mt-1 text-base">{describe(status)}</p>
          {status.state === 'unreachable' && (
            <p className="mt-1 text-sm text-slate-600">{status.reason}</p>
          )}
          <p className="mt-2 break-all font-mono text-xs text-slate-600">{apiBaseUrl}</p>
        </section>
      </div>
    </main>
  )
}

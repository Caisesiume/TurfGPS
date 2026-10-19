// @vitest-environment node
//
// NFR-005's acceptance test: issue #26, AC1 and AC2, in the arrangement
// `docs/adr/ADR-0006-deployment-topology.md` deploys. The client is static
// files on one origin (§ D3, Vercel), and the browser calls the service on
// another, under CORS, never through a rewrite (§ D4).
//
// What is real and what stands in, phrase by phrase of NFR-005's Verification:
//
//   "the built client": a production build into a temporary directory. Nothing
//   here imports `web/src`; no dev server, no stale `web/dist`.
//
//   "a plain static file server": a `node:http` server that answers every
//   request by reading a file out of that directory, or 404. It forwards
//   nothing, rewrites nothing, and has no SPA fallback. Under ADR-0006 it needs
//   none, because the client's call goes to the service's own origin.
//
//   "with no application runtime present": no file from `web/src` or from the
//   build output executes on the server side. The harness is a Node process, as
//   any static host is a program; AC1 bounds project code run to serve the
//   client, not the existence of software on the host.
//
//   "a call to the service over HTTP": a stub on a second loopback origin,
//   answering the one route the service serves today. Both hops are real HTTP.
//   The stub admits the client's origin with Access-Control-Allow-Origin; the
//   real service sends no CORS header yet, and that is owed to the backend
//   story that adds it, not to this file.
//
//   "loads": the document and every subresource are fetched over HTTP from the
//   static origin. jsdom does not run `<script type="module">` and has no
//   fetch, so the bundle is evaluated as a classic script in a window whose URL
//   is that origin, and the window is given a fetch that applies the browser's
//   rule for a simple cross-origin GET: it sends Origin, and unless the answer
//   carries Access-Control-Allow-Origin for that origin or `*`, it rejects with
//   the TypeError a browser throws. A request a browser would preflight is
//   refused outright, because no preflight is emulated. This is an emulation,
//   not a browser; the second evaluation, against a service sending no CORS
//   header, is what shows it can refuse.
//
// `vite preview` is not used: it is Vite's own server, an application runtime,
// and not the deployed artefact.

import { execFileSync } from 'node:child_process'
import { createServer, type Server } from 'node:http'
import { existsSync, mkdtempSync, readdirSync, rmSync, statSync } from 'node:fs'
import { readFile } from 'node:fs/promises'
import { tmpdir } from 'node:os'
import { dirname, join } from 'node:path'
import { fileURLToPath } from 'node:url'
import { JSDOM } from 'jsdom'
import { afterAll, beforeAll, expect, test } from 'vitest'

const webDir = join(dirname(fileURLToPath(import.meta.url)), '..')

// Distinctive, so rendering it proves the text crossed the wire from the stub.
const SERVICE_ANSWER = 'TurfGPS service (NFR-005 harness upstream)'
const UNREACHABLE = 'The service could not be reached.'

const CONTENT_TYPES: Record<string, string> = {
  css: 'text/css',
  html: 'text/html; charset=utf-8',
  ico: 'image/x-icon',
  js: 'text/javascript',
  png: 'image/png',
  svg: 'image/svg+xml',
  woff2: 'font/woff2',
}

// Corroborating only: a name list can be evaded by a framework it does not
// name. What decides AC1 is that the load was answered entirely out of files.
const SERVER_SHAPED =
  /(^|\/)(package\.json|server|functions|\.netlify|\.vercel|\.output|\.next)(\/|$)|(^|\/)(entry-server|ssr-manifest)[^/]*$|\.node$/
// What Vercel would execute if it sat in the project: functions under `api/`
// and routing middleware. ADR-0006 lets the client go there as files only.
const VERCEL_EXECUTES = ['api', 'middleware.js', 'middleware.ts', 'middleware.mjs', 'middleware.mts']

type Evaluation = { requested: string[]; heading: string; status: string; paragraphs: string[] }

let outDir = ''
let builtFiles: string[] = []
let staticOrigin = ''
let serviceBase = ''
let serviceSendsCors = true
const servedByStatic: string[] = []
const receivedByService: string[] = []
const subresources: string[] = []
const evaluatedSources: string[] = []
let withCors: Evaluation = { requested: [], heading: '', status: '', paragraphs: [] }
let withoutCors: Evaluation = withCors
const servers: Server[] = []
const windows: { close: () => void }[] = []

function listFiles(root: string, prefix = ''): string[] {
  return readdirSync(root).flatMap((entry) => {
    const rel = prefix === '' ? entry : `${prefix}/${entry}`
    return statSync(join(root, entry)).isDirectory() ? listFiles(join(root, entry), rel) : [rel]
  })
}

function listen(server: Server): Promise<string> {
  servers.push(server)
  return new Promise((resolve) => {
    server.listen(0, '127.0.0.1', () => {
      const address = server.address()
      resolve(`http://127.0.0.1:${typeof address === 'object' && address !== null ? address.port : 0}`)
    })
  })
}

// The static host: a file out of the build output, or 404. Nothing else.
function staticHost(): Server {
  return createServer((request, response) => {
    void (async () => {
      const pathname = new URL(request.url ?? '/', 'http://placeholder').pathname
      const segments = decodeURIComponent(pathname).split('/').filter((s) => s !== '')
      const escapes = segments.some((s) => s === '.' || s === '..' || s.includes('\\'))
      const file = join(outDir, ...(segments.length === 0 ? ['index.html'] : segments))
      const bytes = escapes ? null : await readFile(file).catch(() => null)
      servedByStatic.push(`${request.method} ${pathname} ${bytes === null ? 'absent 404' : 'file 200'}`)
      if (bytes === null) {
        response.writeHead(404).end()
        return
      }
      const extension = file.slice(file.lastIndexOf('.') + 1).toLowerCase()
      response.writeHead(200, { 'content-type': CONTENT_TYPES[extension] ?? 'application/octet-stream' })
      response.end(bytes)
    })()
  })
}

// The service stand-in: its one route, and the CORS header the backend story
// will add, admitting the client's origin and no other.
function serviceStub(): Server {
  return createServer((request, response) => {
    const origin = request.headers.origin ?? '-'
    receivedByService.push(`${request.method} ${request.url} origin=${origin}`)
    const headers: Record<string, string> = { 'content-type': 'text/plain; charset=utf-8' }
    if (serviceSendsCors && origin === staticOrigin) headers['access-control-allow-origin'] = staticOrigin
    response.writeHead(200, headers)
    response.end(`${SERVICE_ANSWER}\n`)
  })
}

// The browser's rule for a cross-origin request, as far as a simple GET needs.
function browserFetch(pageOrigin: string, requested: string[]) {
  return async (input: unknown, init?: RequestInit): Promise<Response> => {
    const url = new URL(String(input), `${pageOrigin}/`)
    requested.push(url.href)
    if (url.origin === pageOrigin) return fetch(url, init)
    const method = (init?.method ?? 'GET').toUpperCase()
    if (method !== 'GET' || [...new Headers(init?.headers).keys()].length > 0) {
      throw new Error(`harness: ${method} ${url.href} would be preflighted, and no preflight is emulated`)
    }
    const response = await fetch(url, { ...init, headers: { origin: pageOrigin } })
    const allowed = response.headers.get('access-control-allow-origin')
    if (allowed !== pageOrigin && allowed !== '*') {
      await response.arrayBuffer()
      throw new TypeError('Failed to fetch')
    }
    return response
  }
}

// Evaluate the built bundle in a window at the static origin, and read what it
// rendered once the status leaves its checking state.
async function evaluate(html: string): Promise<Evaluation> {
  const requested: string[] = []
  const dom = new JSDOM(html, { url: `${staticOrigin}/`, runScripts: 'outside-only', pretendToBeVisual: true })
  windows.push(dom.window)
  // One set, as a browser provides them: undici rejects a foreign AbortSignal.
  Object.assign(dom.window, { AbortController, AbortSignal, fetch: browserFetch(staticOrigin, requested) })
  for (const source of evaluatedSources) dom.window.eval(source)

  const root = () => dom.window.document.getElementById('root')
  const status = () => root()?.querySelector('[role="status"]')?.textContent ?? ''
  const deadline = Date.now() + 5_000
  while (Date.now() < deadline && (status() === '' || status().startsWith('Checking'))) {
    await new Promise((resolve) => setTimeout(resolve, 20))
  }
  return {
    requested,
    heading: root()?.querySelector('h1')?.textContent ?? '',
    status: status(),
    paragraphs: [...(root()?.querySelectorAll('p') ?? [])].map((p) => p.textContent ?? ''),
  }
}

beforeAll(async () => {
  outDir = mkdtempSync(join(tmpdir(), 'turfgps-nfr005-'))
  staticOrigin = await listen(staticHost())
  serviceBase = await listen(serviceStub())

  // The service's address is a build input, as it is on Vercel. NODE_ENV is
  // removed because vitest sets it to `test`, which would build React's
  // development entry: an artefact no deploy produces.
  const env: NodeJS.ProcessEnv = { ...process.env, VITE_API_BASE_URL: serviceBase }
  delete env.NODE_ENV
  execFileSync(
    process.execPath,
    [join(webDir, 'node_modules', 'vite', 'bin', 'vite.js'), 'build', '--outDir', outDir, '--emptyOutDir'],
    { cwd: webDir, env, stdio: 'pipe' },
  )
  builtFiles = listFiles(outDir).sort()

  // Load it as a browser does: the document, then every subresource it names.
  const document = await fetch(`${staticOrigin}/`)
  const html = document.ok ? await document.text() : ''
  const parsed = new JSDOM(html)
  const references = [...parsed.window.document.querySelectorAll('script[src], link[href]')].map(
    (element) => element.getAttribute('src') ?? element.getAttribute('href') ?? '',
  )
  parsed.window.close()
  for (const reference of references) {
    const asset = await fetch(new URL(reference, `${staticOrigin}/`))
    subresources.push(reference)
    if (asset.ok && reference.endsWith('.js')) evaluatedSources.push(await asset.text())
    else await asset.arrayBuffer()
  }
  if (html === '' || evaluatedSources.length === 0) return

  withCors = await evaluate(html)
  serviceSendsCors = false
  withoutCors = await evaluate(html)
}, 120_000)

afterAll(() => {
  for (const window of windows) window.close()
  for (const server of servers) server.close()
  if (outDir !== '') rmSync(outDir, { recursive: true, force: true })
})

test('AC1: every request the static host answered was a file read, and nothing else reached it', () => {
  // Exact on purpose. The client's call goes to the service's origin, so any
  // request it sent here, or anything the host failed to find, is an extra
  // entry; AC1's threshold is none.
  expect(servedByStatic).toEqual(['GET / file 200', ...subresources.map((r) => `GET ${r} file 200`)])
  expect(subresources.length).toBeGreaterThan(0)
})

test('AC1: neither the build output nor the project carries anything a host would execute', () => {
  expect(builtFiles.filter((file) => SERVER_SHAPED.test(file))).toEqual([])
  expect(builtFiles).toContain('index.html')
  expect(VERCEL_EXECUTES.filter((name) => existsSync(join(webDir, name)))).toEqual([])
})

test('AC1: the build output is the one script index.html declares, and no other', () => {
  // Discovery reads index.html only, so a second chunk would be served by the
  // deploy and seen by no assertion here. Pinning the count breaks that silence.
  expect(builtFiles.filter((file) => file.endsWith('.js'))).toHaveLength(1)
  expect(evaluatedSources).toHaveLength(1)
})

test('AC2: nothing evaluated carries a dynamic import, which a classic-script evaluation would not catch', () => {
  expect(evaluatedSources.flatMap((source) => source.match(/\bimport\s*\([^)]{0,60}/g) ?? [])).toEqual([])
})

test('AC2: the built client loads and mounts from the static host, naming the service it was built for', () => {
  // The heading is an element the service's answer cannot become; the base is
  // rendered from the same constant the call is built from.
  expect(withCors.heading).toBe('TurfGPS')
  expect(withCors.paragraphs).toContain(serviceBase)
})

test('AC2: its first call crosses to the service origin over HTTP and completes under CORS', () => {
  // One call, to the absolute base fixed at build time, carrying the client's
  // origin. A second identical call would mean a development build slipped in.
  expect(withCors.requested).toEqual([`${serviceBase}/`])
  expect(receivedByService[0]).toBe(`GET / origin=${staticOrigin}`)
  expect(withCors.status).toBe(SERVICE_ANSWER)
})

test('AC2 depends on CORS: a service sending no CORS header is answered but refused, and the client says so', () => {
  // This is the real service today. The request arrives and is answered; the
  // browser withholds the answer, and the card reports the service unreachable.
  expect(withoutCors.requested).toEqual([`${serviceBase}/`])
  expect(receivedByService).toEqual([`GET / origin=${staticOrigin}`, `GET / origin=${staticOrigin}`])
  expect(withoutCors.status).toBe(UNREACHABLE)
})

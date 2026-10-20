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
//   "a call to the service over HTTP": the service itself. Its binary is built
//   from `service/` with the host's Go toolchain and started as an operator
//   starts it, with TURFGPS_CORS_ALLOWED_ORIGINS naming the static host's
//   origin (#224) and no other TURFGPS_ variable. It listens on :8080, the one
//   address it has (`service/cmd/turfgps/main.go`), so this file waits for that
//   port to be free and is ready only once the service it started says it is
//   listening. A missing Go toolchain or a port that stays taken fails the
//   file; nothing here skips.
//
//   Without a grant: a second static host on another loopback origin serves
//   the same files, and the grant does not name it. Same service process, same
//   bundle; only the page's origin differs, so the refusal is the service's
//   own exact-origin decision.
//
//   "loads": the document and every subresource are fetched over HTTP from the
//   static origin. jsdom does not run `<script type="module">` and has no
//   fetch, so the bundle is evaluated as a classic script in a window whose URL
//   is that origin, and the window is given a fetch that applies the browser's
//   rule for a simple cross-origin GET: it sends Origin, and unless the answer
//   carries Access-Control-Allow-Origin for that origin or `*`, it rejects with
//   the TypeError a browser throws. A request a browser would preflight is
//   refused outright, because no preflight is emulated. The browser is the one
//   part emulated; the headers it judges are the service's own.
//
// `vite preview` is not used: it is Vite's own server, an application runtime,
// and not the deployed artefact.

import { execFileSync, spawn, type ChildProcess } from 'node:child_process'
import { createServer, type Server } from 'node:http'
import { connect } from 'node:net'
import { existsSync, mkdirSync, mkdtempSync, readdirSync, rmSync, statSync } from 'node:fs'
import { readFile } from 'node:fs/promises'
import { tmpdir } from 'node:os'
import { dirname, join } from 'node:path'
import { fileURLToPath } from 'node:url'
import { JSDOM } from 'jsdom'
import { afterAll, beforeAll, expect, test } from 'vitest'

const webDir = join(dirname(fileURLToPath(import.meta.url)), '..')
const serviceDir = join(webDir, '..', 'service')

// The service's one listen address, `defaultAddr` in `service/cmd/turfgps/main.go`.
const SERVICE_PORT = 8080
const serviceBase = `http://127.0.0.1:${SERVICE_PORT}`
const UNREACHABLE = 'The service could not be reached.'
const NO_GRANT = 'no grant'

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

type Page = { html: string; references: string[]; sources: string[] }
type Evaluation = { requested: string[]; answered: string[]; heading: string; status: string; paragraphs: string[] }

const emptyPage: Page = { html: '', references: [], sources: [] }
const emptyEvaluation: Evaluation = { requested: [], answered: [], heading: '', status: '', paragraphs: [] }

let workDir = ''
let outDir = ''
let builtFiles: string[] = []
let staticOrigin = ''
let ungrantedOrigin = ''
let service: ChildProcess | null = null
let serviceLog = ''
let serviceAnswer = ''
const servedByStatic: string[] = []
const servedByUngranted: string[] = []
let page: Page = emptyPage
let withCors: Evaluation = emptyEvaluation
let withoutCors: Evaluation = emptyEvaluation
const servers: Server[] = []
const windows: { close: () => void }[] = []

function sleep(ms: number): Promise<void> {
  return new Promise((resolve) => setTimeout(resolve, ms))
}

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
function staticHost(log: string[]): Server {
  return createServer((request, response) => {
    void (async () => {
      const pathname = new URL(request.url ?? '/', 'http://placeholder').pathname
      const segments = decodeURIComponent(pathname).split('/').filter((s) => s !== '')
      const escapes = segments.some((s) => s === '.' || s === '..' || s.includes('\\'))
      const file = join(outDir, ...(segments.length === 0 ? ['index.html'] : segments))
      const bytes = escapes ? null : await readFile(file).catch(() => null)
      log.push(`${request.method} ${pathname} ${bytes === null ? 'absent 404' : 'file 200'}`)
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

function portTaken(port: number): Promise<boolean> {
  return new Promise((resolve) => {
    const socket = connect(port, '127.0.0.1')
    socket.once('connect', () => {
      socket.destroy()
      resolve(true)
    })
    socket.once('error', () => resolve(false))
  })
}

// Build the service from `service/` and start it granting `origin` alone.
async function startService(origin: string): Promise<void> {
  const binary = join(workDir, process.platform === 'win32' ? 'turfgps.exe' : 'turfgps')
  try {
    execFileSync('go', ['build', '-o', binary, './cmd/turfgps'], { cwd: serviceDir, stdio: 'pipe' })
  } catch (error) {
    const stderr = (error as { stderr?: Buffer }).stderr?.toString() ?? ''
    throw new Error(`go build ./cmd/turfgps failed in ${serviceDir}; this test runs the real service:\n${stderr}`, { cause: error })
  }

  // A second gate run on the same host holds the port for seconds; anything
  // holding it longer is not this file's service, and is not talked to.
  const deadline = Date.now() + 30_000
  while (await portTaken(SERVICE_PORT)) {
    if (Date.now() > deadline) {
      throw new Error(`port ${SERVICE_PORT} stayed taken for 30 s; the service listens there and nowhere else`)
    }
    await sleep(250)
  }

  // Only the grant: an inherited TURFGPS_ variable would configure more than
  // this file states, such as a zone sync reaching a database.
  const env = Object.fromEntries(Object.entries(process.env).filter(([name]) => !name.startsWith('TURFGPS_')))
  env.TURFGPS_CORS_ALLOWED_ORIGINS = origin
  const child = spawn(binary, [], { env, stdio: ['ignore', 'pipe', 'pipe'] })
  service = child

  // Ready when it says so: an answer on the port alone could be another process's.
  await new Promise<void>((resolve, reject) => {
    const timer = setTimeout(() => reject(new Error(`the service did not report listening within 15 s:\n${serviceLog}`)), 15_000)
    const onOutput = (chunk: Buffer) => {
      serviceLog += chunk.toString()
      if (/\blistening\b/.test(serviceLog)) {
        clearTimeout(timer)
        resolve()
      }
    }
    child.stdout?.on('data', onOutput)
    child.stderr?.on('data', onOutput)
    child.once('exit', (code) => {
      clearTimeout(timer)
      reject(new Error(`the service exited ${code} before listening:\n${serviceLog}`))
    })
  })
}

async function stopService(): Promise<void> {
  const child = service
  if (child === null || child.exitCode !== null || child.signalCode !== null) return
  const exited = new Promise((resolve) => child.once('exit', resolve))
  child.kill()
  await exited
}

// Load a page as a browser does: the document, then every subresource it names.
async function load(origin: string): Promise<Page> {
  const document = await fetch(`${origin}/`)
  const html = document.ok ? await document.text() : ''
  const parsed = new JSDOM(html)
  const references = [...parsed.window.document.querySelectorAll('script[src], link[href]')].map(
    (element) => element.getAttribute('src') ?? element.getAttribute('href') ?? '',
  )
  parsed.window.close()
  const sources: string[] = []
  for (const reference of references) {
    const asset = await fetch(new URL(reference, `${origin}/`))
    if (asset.ok && reference.endsWith('.js')) sources.push(await asset.text())
    else await asset.arrayBuffer()
  }
  return { html, references, sources }
}

// The browser's rule for a cross-origin request, as far as a simple GET needs.
// `answered` records each cross-origin answer's status and the grant it carried.
function browserFetch(pageOrigin: string, requested: string[], answered: string[]) {
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
    answered.push(`${response.status} ${allowed ?? NO_GRANT}`)
    if (allowed !== pageOrigin && allowed !== '*') {
      await response.arrayBuffer()
      throw new TypeError('Failed to fetch')
    }
    return response
  }
}

// Evaluate a page's bundle in a window at `origin`, and read what it rendered
// once the status leaves its checking state.
async function evaluate(origin: string, loaded: Page): Promise<Evaluation> {
  const requested: string[] = []
  const answered: string[] = []
  const dom = new JSDOM(loaded.html, { url: `${origin}/`, runScripts: 'outside-only', pretendToBeVisual: true })
  windows.push(dom.window)
  // One set, as a browser provides them: undici rejects a foreign AbortSignal.
  Object.assign(dom.window, { AbortController, AbortSignal, fetch: browserFetch(origin, requested, answered) })
  for (const source of loaded.sources) dom.window.eval(source)

  const root = () => dom.window.document.getElementById('root')
  const status = () => root()?.querySelector('[role="status"]')?.textContent ?? ''
  const deadline = Date.now() + 5_000
  while (Date.now() < deadline && (status() === '' || status().startsWith('Checking'))) {
    await sleep(20)
  }
  return {
    requested,
    answered,
    heading: root()?.querySelector('h1')?.textContent ?? '',
    status: status(),
    paragraphs: [...(root()?.querySelectorAll('p') ?? [])].map((p) => p.textContent ?? ''),
  }
}

beforeAll(async () => {
  workDir = mkdtempSync(join(tmpdir(), 'turfgps-nfr005-'))
  outDir = join(workDir, 'client')
  mkdirSync(outDir)
  staticOrigin = await listen(staticHost(servedByStatic))
  ungrantedOrigin = await listen(staticHost(servedByUngranted))

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

  await startService(staticOrigin)
  // What the service answers its root route, read directly, so the card can be
  // held to it without this file restating the service's text.
  serviceAnswer = (await (await fetch(`${serviceBase}/`)).text()).trim()

  page = await load(staticOrigin)
  if (page.html === '' || page.sources.length === 0) return
  withCors = await evaluate(staticOrigin, page)
  withoutCors = await evaluate(ungrantedOrigin, await load(ungrantedOrigin))
}, 180_000)

afterAll(async () => {
  for (const window of windows) window.close()
  for (const server of servers) server.close()
  await stopService()
  if (workDir !== '') rmSync(workDir, { recursive: true, force: true })
})

test('AC1: every request the static host answered was a file read, and nothing else reached it', () => {
  // Exact on purpose. The client's call goes to the service's origin, so any
  // request it sent here, or anything the host failed to find, is an extra
  // entry; AC1's threshold is none.
  expect(servedByStatic).toEqual(['GET / file 200', ...page.references.map((r) => `GET ${r} file 200`)])
  expect(page.references.length).toBeGreaterThan(0)
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
  expect(page.sources).toHaveLength(1)
})

test('AC2: nothing evaluated carries a dynamic import, which a classic-script evaluation would not catch', () => {
  expect(page.sources.flatMap((source) => source.match(/\bimport\s*\([^)]{0,60}/g) ?? [])).toEqual([])
})

test('AC2: the built client loads and mounts from the static host, naming the service it was built for', () => {
  // The heading is an element the service's answer cannot become; the base is
  // rendered from the same constant the call is built from.
  expect(withCors.heading).toBe('TurfGPS')
  expect(withCors.paragraphs).toContain(serviceBase)
})

test('AC2: its first call reaches the real service over HTTP and the card shows its answer, under the grant', () => {
  // One call, to the absolute base fixed at build time. A second identical
  // call would mean a development build slipped in. The answer is not in the
  // bundle, so rendering it means it crossed the wire.
  expect(serviceAnswer).not.toBe('')
  expect(page.sources.filter((source) => source.includes(serviceAnswer))).toEqual([])
  expect(withCors.requested).toEqual([`${serviceBase}/`])
  expect(withCors.status).toBe(serviceAnswer)
  expect(withCors.answered).toEqual([`200 ${staticOrigin}`])
})

test('AC2 without a grant: the service answers an origin it does not name, the browser withholds it, and the card says so', () => {
  // The same files from a second static host. They load and make the same
  // call; the service answers it with no grant, so the answer is refused.
  expect(servedByUngranted[0]).toBe('GET / file 200')
  expect(withoutCors.heading).toBe('TurfGPS')
  expect(withoutCors.requested).toEqual([`${serviceBase}/`])
  expect(withoutCors.answered).toEqual([`200 ${NO_GRANT}`])
  expect(withoutCors.status).toBe(UNREACHABLE)
})

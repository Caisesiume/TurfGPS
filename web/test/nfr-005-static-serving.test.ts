// @vitest-environment node
//
// NFR-005's acceptance test — issue #26, AC1 and AC2.
//
// NFR-005's `Verification` field reads:
//
//   test — the built client is served by a plain static file server with no
//   application runtime present, and loads and completes a call to the service
//   over HTTP
//
// That clause admits two readings, and which one a harness picks decides
// whether the test can fail on anything. So this file states, phrase by
// phrase, which words it satisfies literally and which it satisfies under the
// arrangement `DEPLOYMENT.md § Deployment architecture` actually deploys.
// Picking whichever reading goes green most easily is the defect
// `docs/DELIVERY.md § Proof that a test can fail` exists to prevent.
//
// LITERALLY
//
//   "the built client" — `beforeAll` runs the production build into a
//   temporary directory and serves nothing else. This file imports nothing
//   from `web/src`; no dev server, no source transform, no stale `web/dist`.
//
//   "a static file server" — every byte of the client is produced by reading a
//   file off that directory and writing it out. The origin holds no route
//   table for the client, no template, no rewrite, and deliberately no SPA
//   fallback to index.html.
//
//   "over HTTP" — two real `node:http` servers on loopback. Nothing in this
//   file stubs `fetch`, which is the whole difference between it and
//   `web/src/App.test.tsx`: that one asserts what the component does with an
//   answer, this one asserts that an answer arrives at all.
//
//   "with no application runtime present" — literally in the sense AC1's
//   metric states it, *server-side application code executed to serve the
//   client: none*, and that count is zero here: no file from `web/src` and no
//   file from the build output is executed on the server side. Not in the
//   sense of *no process exists* — this harness is itself a Node process, as
//   any static host is itself a program. AC1 bounds project code executing to
//   produce a client response, not the existence of software on the host.
//
// UNDER THE DEPLOYED READING
//
//   "plain" — a plain file server does not forward anything, and this origin
//   forwards `/api` to the service. `DEPLOYMENT.md § Deployment architecture`
//   puts one reverse proxy in front of both — it serves the client's files and
//   routes /api to the service, on one origin — so serving the client is
//   plain, and the forward is the second half of that same deployed component.
//   Read literally instead, the client and the service land on different
//   origins, the service sets no CORS headers, and no call completes: the
//   literal reading makes AC2 unsatisfiable by the arrangement this project
//   deploys. The test is written against the deployed arrangement, and that
//   the clause admits a reading its own deployment model cannot satisfy is
//   reported as a defect in the record's wording rather than repaired here.
//
//   "loads" — literal for the HTTP half: the document and every subresource it
//   references are fetched from the origin and must be answered from files.
//   Under the reading for the execution half: the built bundle is evaluated in
//   a jsdom window whose URL is that origin, because jsdom does not execute
//   `<script type="module">`. Today's build is one self-contained chunk with
//   no `import`, no `export` and no `import.meta`, so classic-script and
//   module evaluation coincide; if chunking ever changes, the evaluation
//   throws loudly rather than passing quietly.
//
//   "a call to the service" — the upstream is a stub answering the one route
//   the service serves today. This item's scope excludes `service/`, and the
//   criterion measures whether the client can reach a service from the static
//   arrangement, not what the service replies. What is proved without a
//   reading is that the call leaves the client as a relative URL, resolves to
//   the origin the client was served from, crosses a real HTTP connection, and
//   arrives back in the rendered document.
//
// `vite preview` is not used, and could not stand in for this: it is Vite's
// own server carrying a configured dev proxy, so it is both an application
// runtime and a development-server configuration rather than the deployed
// artefact — it satisfies neither reading of the clause.
//
// This file is outside `tsconfig.json`'s `include`, so `tsc --noEmit` does not
// check it and no `@types/node` or `@types/jsdom` dependency is added for it.
// The harness is exercised end to end by its own assertions, which is what
// would surface a mistake in it.

import { execFileSync } from 'node:child_process'
import { createServer } from 'node:http'
import { mkdtempSync, readdirSync, rmSync, statSync } from 'node:fs'
import { readFile } from 'node:fs/promises'
import { tmpdir } from 'node:os'
import { dirname, join } from 'node:path'
import { fileURLToPath } from 'node:url'
import { JSDOM } from 'jsdom'
import { afterAll, beforeAll, expect, test } from 'vitest'

const webDir = join(dirname(fileURLToPath(import.meta.url)), '..')

// Distinctive on purpose: rendering it proves the text crossed the wire from
// the upstream, rather than coming from anything baked into the bundle.
const SERVICE_ANSWER = 'TurfGPS service (NFR-005 harness upstream)'

const CONTENT_TYPES: Record<string, string> = {
  css: 'text/css',
  html: 'text/html; charset=utf-8',
  ico: 'image/x-icon',
  js: 'text/javascript',
  json: 'application/json',
  png: 'image/png',
  svg: 'image/svg+xml',
  woff2: 'font/woff2',
}

// The shapes a server-rendering or serverless target emits into a build
// output. Corroborating only, and deliberately not the check that decides AC1:
// a name list can always be evaded by a framework it does not name. What
// decides AC1 is that the load below was answered entirely out of files.
const SERVER_SHAPED =
  /(^|\/)(package\.json|server|functions|\.netlify|\.vercel|\.output|\.next)(\/|$)|(^|\/)(entry-server|ssr-manifest)[^/]*$|\.node$/

type Served = { path: string; mode: 'file' | 'forward' | 'absent'; status: number }

let outDir = ''
let builtFiles: string[] = []
const served: Served[] = []
const requestedByClient: string[] = []
const receivedByService: string[] = []
const subresources: string[] = []
let rootText = ''
let servers: { close: () => void }[] = []
let window: { close: () => void } | null = null

function listFiles(root: string, prefix = ''): string[] {
  return readdirSync(root).flatMap((entry) => {
    const full = join(root, entry)
    const rel = prefix === '' ? entry : `${prefix}/${entry}`
    return statSync(full).isDirectory() ? listFiles(full, rel) : [rel]
  })
}

// Resolve a URL path to a file inside the build output, or null if it escapes.
function resolveInOutput(pathname: string): string | null {
  const segments = decodeURIComponent(pathname)
    .split('/')
    .filter((segment) => segment !== '')
  if (segments.some((s) => s === '.' || s === '..' || s.includes('\\'))) return null
  return join(outDir, ...(segments.length === 0 ? ['index.html'] : segments))
}

function listen(server: ReturnType<typeof createServer>): Promise<number> {
  return new Promise((resolve) => {
    server.listen(0, '127.0.0.1', () => {
      const address = server.address()
      resolve(typeof address === 'object' && address !== null ? address.port : 0)
    })
  })
}

beforeAll(async () => {
  outDir = mkdtempSync(join(tmpdir(), 'turfgps-nfr005-'))

  // vite directly rather than `npm run build`: that script is `tsc --noEmit &&
  // vite build`, and the type check emits nothing, so the artefact is the same
  // one and the build gate already runs the check.
  //
  // The two variables below are removed rather than inherited, and NODE_ENV is
  // the one that matters. Vitest sets it to `test`, and a build carrying that
  // resolves React's development entry: measured here, 389 kB against the
  // deployed build's 191 kB, and its StrictMode double-invokes the effect that
  // calls the service. Inheriting it would have left this file asserting
  // against an artefact no deploy produces while still going green.
  // VITE_API_BASE_URL is removed so the build takes the same-origin default
  // `/api` that `web/.env.example` documents and the deployment model deploys.
  const env = { ...process.env }
  delete env.NODE_ENV
  delete env.VITE_API_BASE_URL
  execFileSync(
    process.execPath,
    [join(webDir, 'node_modules', 'vite', 'bin', 'vite.js'), 'build', '--outDir', outDir, '--emptyOutDir'],
    { cwd: webDir, env, stdio: 'pipe' },
  )
  builtFiles = listFiles(outDir).sort()

  // The service. Reachable only over HTTP, and it records what reached it.
  const service = createServer((request, response) => {
    receivedByService.push(`${request.method} ${request.url}`)
    response.writeHead(200, { 'content-type': 'text/plain; charset=utf-8' })
    response.end(`${SERVICE_ANSWER}\n`)
  })
  const servicePort = await listen(service)

  // The origin the client is served from: files for the client, and the /api
  // forward `DEPLOYMENT.md § Deployment architecture` gives the reverse proxy.
  // The prefix is stripped because the service's only route today is its root,
  // which is what `web/vite.config.ts` does for the dev and preview servers.
  const origin = createServer((request, response) => {
    void (async () => {
      const url = new URL(request.url ?? '/', 'http://placeholder')
      if (url.pathname === '/api' || url.pathname.startsWith('/api/')) {
        const upstream = await fetch(`http://127.0.0.1:${servicePort}${url.pathname.slice(4) || '/'}`, {
          method: request.method,
        })
        const body = Buffer.from(await upstream.arrayBuffer())
        served.push({ path: url.pathname, mode: 'forward', status: upstream.status })
        response.writeHead(upstream.status, {
          'content-type': upstream.headers.get('content-type') ?? 'application/octet-stream',
        })
        response.end(body)
        return
      }
      const file = resolveInOutput(url.pathname)
      let bytes: Buffer | null = null
      if (file !== null) {
        bytes = await readFile(file).catch(() => null)
      }
      if (bytes === null) {
        served.push({ path: url.pathname, mode: 'absent', status: 404 })
        response.writeHead(404).end()
        return
      }
      served.push({ path: url.pathname, mode: 'file', status: 200 })
      const extension = file!.slice(file!.lastIndexOf('.') + 1).toLowerCase()
      response.writeHead(200, { 'content-type': CONTENT_TYPES[extension] ?? 'application/octet-stream' })
      response.end(bytes)
    })()
  })
  const base = `http://127.0.0.1:${await listen(origin)}/`
  servers = [service, origin]

  // Load it the way a browser does: the document, then every subresource it
  // references, each over HTTP from that origin. Nothing here throws on a bad
  // answer — a failed load has to reach an assertion to produce a message.
  const document = await fetch(base)
  const html = document.ok ? await document.text() : ''
  const parsed = new JSDOM(html)
  const references = [...parsed.window.document.querySelectorAll('script[src], link[href]')].map(
    (element) => element.getAttribute('src') ?? element.getAttribute('href') ?? '',
  )
  const sources: string[] = []
  for (const reference of references) {
    const asset = await fetch(new URL(reference, base))
    subresources.push(reference)
    if (asset.ok && reference.endsWith('.js')) sources.push(await asset.text())
    else await asset.arrayBuffer()
  }
  parsed.window.close()

  if (html === '' || sources.length === 0) return

  // Execute the built bundle against a document whose URL is that origin.
  // fetch, AbortController and AbortSignal are injected as one set because a
  // browser provides them as one: undici rejects a foreign AbortSignal, and
  // resolving a relative URL against the document is what a browser does with
  // the client's `/api` — an absolute URL baked in at build time would ignore
  // that base and never reach this origin, which is the point of recording it.
  const dom = new JSDOM(html, { url: base, runScripts: 'outside-only', pretendToBeVisual: true })
  window = dom.window
  Object.assign(dom.window, {
    AbortController,
    AbortSignal,
    fetch: (input: unknown, init?: unknown) => {
      requestedByClient.push(String(input))
      return fetch(new URL(String(input), dom.window.location.href), init as RequestInit)
    },
  })
  for (const source of sources) dom.window.eval(source)

  const read = () => dom.window.document.getElementById('root')?.textContent ?? ''
  const deadline = Date.now() + 5_000
  while (Date.now() < deadline && !read().includes(SERVICE_ANSWER)) {
    await new Promise((resolve) => setTimeout(resolve, 20))
  }
  rootText = read()
}, 120_000)

afterAll(() => {
  window?.close()
  for (const server of servers) server.close()
  if (outDir !== '') rmSync(outDir, { recursive: true, force: true })
})

const servingTheClient = () => served.filter((entry) => !entry.path.startsWith('/api'))

test('AC1 — every byte of the client is a file read, and nothing else ran to serve it', () => {
  // The exact list is the assertion, and the exactness is the point: anything
  // the origin forwarded, executed or failed to find while the client loaded
  // would appear here as an extra entry or a mode other than `file`, and AC1's
  // threshold is *none*. The client's own call to the service is excluded
  // because it is AC2's subject rather than part of serving the client.
  expect(servingTheClient().map((e) => `${e.path} ${e.mode} ${e.status}`)).toEqual([
    '/ file 200',
    ...subresources.map((reference) => `${reference} file 200`),
  ])
  expect(subresources.length).toBeGreaterThan(0)
})

test('AC1 — the build output carries no server-shaped artefact', () => {
  expect(builtFiles.filter((file) => SERVER_SHAPED.test(file))).toEqual([])
  expect(builtFiles).toContain('index.html')
})

test('AC2 — the built client loads and mounts from that arrangement', () => {
  expect(rootText).toContain('TurfGPS')
})

test('AC2 — and completes its first call to the service over HTTP, same-origin', () => {
  // Relative, so a browser resolves it against the origin the client was
  // served from. This is what makes the two assertions below attributable to
  // the deployed arrangement rather than to an address baked in at build time.
  //
  // The count is pinned deliberately, and it earned that: a second identical
  // entry here is what exposed the NODE_ENV leak above, the development
  // build's StrictMode calling the service twice. A client that legitimately
  // starts retrying changes what "the first successful call" means and should
  // come back through this assertion rather than past it.
  expect(requestedByClient).toEqual(['/api/'])
  expect(served.filter((entry) => entry.mode === 'forward').map((entry) => entry.path)).toEqual(['/api/'])
  expect(receivedByService).toEqual(['GET /'])
  expect(rootText).toContain(SERVICE_ANSWER)
})

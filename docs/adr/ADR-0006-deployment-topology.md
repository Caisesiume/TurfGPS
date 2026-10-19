# ADR-0006 — Deployment topology: one VPS behind Caddy, the client on Vercel

**Status:** accepted — 2026-10-19, by the Owner on #25 ([#issuecomment-6256368290](https://github.com/Caisesiume/TurfGPS/issues/25#issuecomment-6256368290)); in force from the merge of the PR that adds it.
**Source:** #25 — the question, its options and their reasons, [#issuecomment-6256343407](https://github.com/Caisesiume/TurfGPS/issues/25#issuecomment-6256343407), and the Owner's decision on it, [#issuecomment-6256368290](https://github.com/Caisesiume/TurfGPS/issues/25#issuecomment-6256368290). Where this record and those comments differ, the comments are the Owner's and this record is the defect.
**Relation:** amends `docs/DEPLOYMENT.md` in two respects — the client leaves the host, and the reverse proxy is Caddy — and leaves the rest of its model standing. It changes nothing in `Architecture.md`: `§ D1`, `§ D3` and `§ D4` stand as written.

## Context

`DEPLOYMENT.md` modelled one Linux host: the service as a systemd unit beside PostgreSQL with PostGIS and Valhalla, and an unnamed reverse proxy that terminated TLS **and served the static client**. Certificate renewal was the one operational duty that document created rather than inherited, with a renewal hook to reload the proxy.

The Owner's direction, as the question records it: deploy as easily as possible, with the client on Vercel and its free TLS, and the backend on a persistent server with TLS that costs nothing in certificates. The backend's shape was already fixed elsewhere. `Architecture.md § D1` rules out a serverless target because solve sessions live in the process, `NFR-004` makes that exclusion a `MUST`, and the scheduled zone sync of #142 must keep running between requests.

## Decision

The Owner chose option (a). Verbatim, from [#issuecomment-6256368290](https://github.com/Caisesiume/TurfGPS/issues/25#issuecomment-6256368290):

> **(a) VPS + Caddy.** The backend (service, PostGIS, Valhalla, under systemd) runs on one VPS behind Caddy, which issues and renews Let's Encrypt certificates automatically. The client is deployed on Vercel. The browser calls the API directly on its own subdomain, never through a Vercel rewrite (`DEPLOYMENT.md:60`).

What is now true:

- **D1 — One VPS carries the backend.** The service, PostgreSQL with PostGIS, and Valhalla run on it under systemd, as `DEPLOYMENT.md` already modelled.
- **D2 — Caddy is the reverse proxy and the sole ingress.** It terminates TLS on certificates it obtains and renews itself, from Let's Encrypt by default ([Automatic HTTPS § Issuer fallback](https://caddyserver.com/docs/automatic-https#issuer-fallback)).
- **D3 — The client is deployed on Vercel**, as static files, and is not served from the host.
- **D4 — The browser calls the API on the API's own subdomain, under CORS, and never through a Vercel rewrite**, so that no request line carrying a plan code passes through Vercel.

## Options rejected

The reasons are the ones the question recorded, quoted from [#issuecomment-6256343407](https://github.com/Caisesiume/TurfGPS/issues/25#issuecomment-6256343407).

- **(b) A PaaS — Fly.io, Render, Railway.** Its advantage: *"managed TLS and git-push deploys, the closest to Vercel."* Against it: *"An always-on instance big enough for Valhalla and in-process sessions is likely the dearest option, and free tiers sleep, which breaks the sync. A managed PostGIS would amend D4."*
- **(c) A Cloudflare Tunnel in front of any machine.** Its advantage: *"free TLS and no open ports."* Against it: *"Cloudflare decrypts the traffic, and plan codes travel in the request line (`DEPLOYMENT.md:60`: no host log may record one)."*

## Consequences

**Obliged:**

- `DEPLOYMENT.md` names Caddy and moves the client off the host, in the PR that adds this record.
- Certificate renewal stops being an operator's job and becomes Caddy's. What remains is a check that renewal happened, because Caddy keeps serving while it retries a failed renewal for up to 30 days ([Automatic HTTPS § Errors](https://caddyserver.com/docs/automatic-https#errors)). `DEPLOYMENT.md § Operational duties` carries it.
- The plan-code log constraint reaches two places it did not reach before. Caddy's error log records the request target of a request that fails with a server error, even when no access log is configured, and `DEPLOYMENT.md § Where the deployment configuration lives` states that with its sources. And nothing in this repository configures Vercel's logs, so no plan code may be sent to Vercel: not in an API call through a rewrite, and not in the path or query of a link to the client.
- No deploy artefact is authored here: no Caddyfile, no systemd unit, no push-to-deploy job. `docs/Requirements/DECISIONS.md § RD-009` stands until the analyst re-plans #25, #109 and RD-009 after this record merges.

**Given up, stated rather than implied:**

- **A second operator.** Vercel holds the client, its builds and its logs, and nothing in this repository configures it. The client's share of the plan-code constraint therefore rests on never sending Vercel a code, not on a directive anyone can inspect.
- **Distance from the serverless pressure.** Vercel is a serverless platform, which is the pressure `Architecture.md § D2` names toward the topology `§ D1` rules out. The client goes there as files only. A Vercel function, including a rewrite or proxy to the API, is the step D4 forbids.
- **Same-origin simplicity.** The API is cross-origin from the client, so a CORS policy becomes one more thing to keep correct. Serving both from one origin needed none.
- **Closed ports.** Caddy's automatic issuance needs the domain's A/AAAA records pointing at the host and ports 80 and 443 reachable from outside ([Automatic HTTPS § Overview](https://caddyserver.com/docs/automatic-https#overview)). Option (c) needed no open port.
- **Managed operations.** The host is still the operator's to patch and keep running. Option (b) would have taken part of that over, at its price.
- **Vercel's terms are unchecked.** The question noted that Vercel's free plan is for non-commercial use. That is to be checked against Vercel's terms before TurfGPS ever charges.

**Reversibility: high, and nothing is built yet.** The client is static files and moves to any static host by deploying them there. Caddy can be replaced by any proxy that meets the obligations `DEPLOYMENT.md` places on `deploy/proxy/`. The VPS can be replaced by any single Linux host.

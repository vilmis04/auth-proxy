# auth-proxy

Authentication gate in front of the service. It signs users up and in, keeps the
session in an HttpOnly JWT cookie, and forwards every other request to the
service with the verified username in a `user` header.

```
browser ──▶ Coolify (Traefik) ──▶ auth-proxy ──▶ service
                                   │  /api/auth/*  handled here (sign-up, login, logout, is-authenticated)
                                   │  everything else needs a valid cookie, then is forwarded with
                                   │  `user: <name>` and `X-Internal-Token: <INTERNAL_TOKEN>`
                                   └──▶ Postgres (auth table)
```

The service must only be reachable from the proxy and must reject requests
whose `X-Internal-Token` does not match `INTERNAL_TOKEN`.

## Environment variables

Everything comes from the environment. There are no config files.

| Variable | Required | Description |
|---|---|---|
| `JWT_KEY` | yes | HS256 signing secret, at least 32 bytes. The app refuses to start otherwise. Generate with `openssl rand -base64 48`. |
| `DATABASE_URL` | yes | Postgres URL, e.g. `postgres://user:pass@db:5432/auth?sslmode=disable` |
| `SERVICE_URL` | yes | Base URL of the service, e.g. `http://service:8080` |
| `INTERNAL_TOKEN` | yes | Shared secret sent to the service as `X-Internal-Token` |
| `PORT` | no | Listen port, default `8080` |
| `JWT_TTL` | no | Token and cookie lifetime as a Go duration, default `24h` |
| `COOKIE_DOMAIN` | no | `Domain` of the session cookie. Leave empty for a host-only cookie (layout A) |
| `ALLOWED_ORIGINS` | no | Comma-separated CORS origins. Leave empty when everything is same-origin (layout A) |
| `TRUSTED_PROXIES` | no | Comma-separated IPs/CIDRs allowed to set `X-Forwarded-*`, i.e. the Traefik network. Without it every user appears to come from the proxy's IP, so rate limiting is shared by all users |
| `GIN_MODE` | no | `release` (set in the image). Use `debug` locally |

## Domain layout

### A. One domain, `/api` routed to the proxy (recommended)

```
https://app.example.com/      → client
https://app.example.com/api   → auth-proxy
```

The browser only talks to one origin, so no CORS and no cross-site cookie are
needed. Leave `COOKIE_DOMAIN` and `ALLOWED_ORIGINS` empty. The cookie is
`Secure; HttpOnly; SameSite=Lax` and host-only.

The proxy forwards the path unchanged: `/api/things` reaches the service as
`/api/things`.

### B. Separate subdomains

```
https://app.example.com → client
https://api.example.com → auth-proxy
```

```
COOKIE_DOMAIN=.example.com
ALLOWED_ORIGINS=https://app.example.com
```

The cookie is scoped to the parent domain with `SameSite=Lax`. Requests from the
client need `credentials: "include"`. State-changing requests whose `Origin` is
not in `ALLOWED_ORIGINS` are rejected with 403.

## Deploying on Coolify

Deploys are built and run by Coolify from this repository. Do not build or push
images by hand.

1. Create an application from this GitHub repository with the **Dockerfile** build pack.
2. Set the domain: `https://app.example.com/api` for layout A, or `https://api.example.com` for layout B.
3. Add the environment variables above. Mark `JWT_KEY`, `DATABASE_URL` and `INTERNAL_TOKEN` as secrets.
   Use the internal hostnames of your Postgres and service resources in `DATABASE_URL` and `SERVICE_URL`.
4. Set `TRUSTED_PROXIES` to the network Coolify's proxy connects from
   (`docker network inspect coolify` shows the subnet).
5. Health check: path `/healthz`, port `8080`. The image also carries its own
   `HEALTHCHECK` (`/app -healthcheck`).
6. Enable deploy on push to `master`. Merging a pull request deploys it.

Database tables are created on startup by the embedded migrations in
`internal/storage/migrations`. They run under a Postgres advisory lock, so two
instances starting together are safe, and they are safe to run against a
database that already has the `auth` table.

### Probes

- `GET /healthz`: the process is serving (no database check, so a database blip does not restart the proxy).
- `GET /readyz`: also pings the database.

## Local development

```
export JWT_KEY=$(openssl rand -base64 48)
export INTERNAL_TOKEN=dev-token
export SERVICE_URL=http://localhost:8081
export DATABASE_URL='postgres://postgres:postgres@localhost:5432/auth?sslmode=disable'
export GIN_MODE=debug
make run
```

The session cookie is `Secure`; browsers accept it on `http://localhost`.

| Command | What it does |
|---|---|
| `make build` | Build `bin/app` |
| `make test` | `go test ./...` |
| `make check` | gofmt, `go vet` and tests, as CI does |
| `make docker` | Build the image |

Set `TEST_DATABASE_URL` to a throwaway Postgres database to also run the
migration test; it is skipped otherwise.

## Notes

- `/login` and `/sign-up` are rate limited per IP (in memory, per instance).
- Usernames are stored exactly as typed. Releases before the username fix stored
  them HTML-escaped; `scripts/unescape-usernames.sql` repairs those rows and
  must be run exactly once.

# Paste

Paste is the Yueli site group's anonymous-first, multi-file code sharing product. It uses a Go API, a Nuxt Web app, Foundation
contracts and Identity for browser guest sessions and optional account ownership.

Developer tokens and automated Paste workflows: [开发者令牌 API](docs/developer-tokens.md).

The repository is being rebuilt from the useful product behavior in the legacy CodeShare implementation. Rust, SQLite, local
password handling and the old visual treatment are not compatibility contracts.

## Repository shape

```text
api/                 Public request and response types
cmd/paste/           Go API entrypoint
internal/paste/      Paste lifecycle and access rules
internal/postgres/   PostgreSQL adapter and migrations
internal/httpapi/    HTTP adapter
web/                 Nuxt product app and Playwright acceptance
flightdeck/          Resumable repository work
```

Local lifecycle is owned by the adjacent Workspace CLI rather than ad-hoc background processes.

## Development

Use the adjacent Workspace repository for the complete local composition. It provisions PostgreSQL, starts the shared Identity
provider, applies the Paste schema and launches the API and Web app on loopback:

```powershell
cd ..\workspace
.\environments\paste-local\run.ps1
```

The default local entry points are `http://localhost:3010` for Web and `http://127.0.0.1:8091` for the API. Generated runtime
state belongs to Workspace `.doctor/`; do not launch a second unmanaged copy of either service.

Repository checks can be run independently when the required toolchains are installed:

```powershell
go test ./...
go vet ./...

cd web
pnpm install --frozen-lockfile
pnpm test
pnpm typecheck
pnpm build
pnpm test:e2e
```

The Playwright suite expects the real Workspace composition and covers public creation/viewing on desktop and mobile, protected
access, copy behavior and Identity-backed owner management.

## HTTP surface

- `POST /api/v1/pastes` creates an anonymous or authenticated Paste.
- `GET /api/v1/pastes/{code}` opens an unprotected Paste.
- `POST /api/v1/pastes/{code}/access` submits a password without placing it in the URL.
- `/api/v1/me/pastes` and `/api/v1/me/pastes/{code}` provide owner-scoped list, read, update and delete operations for users and Identity guests.
- `POST /api/v1/pastes/guest-claims` transfers guest content to its signed-in owner; it accepts only an Identity `guest:claim` assertion, never client-supplied owner IDs or ordinary user/PAT credentials.

## Browser guest ownership

Apply migration 0007 before running the guest-aware API/Web pair. The shared Identity Nuxt layer owns the HttpOnly guest cookie and token exchange; Paste requests `paste-api` audience tokens and supplies its resource claim endpoint. `NUXT_PASTE_AUDIENCE` must match the API's configured JWT audience and the Identity client's allowed audiences. Guest sessions request 30 days by default (`NUXT_GUEST_SESSION_TTL_SECONDS`), subject to Identity's TTL policy; cookie security follows the existing product setting.

Guest content remains in the anonymous quota pool and cannot be private. The browser can list, edit and delete its own content until its guest session expires or its cookie is cleared. Signing in claims that content into the account through the shared Identity protocol. A transactional claim ledger prevents stale guest tokens or concurrent writes from leaving new guest-owned rows after transfer. Existing ownerless anonymous content cannot be retroactively claimed.

The composer opens sharing confirmation before creation; permissions, expiry and password are chosen there. Relative expiry starts at submission, not when the editor or confirmation panel opens.

See [PRODUCT.md](PRODUCT.md) for product scope and [DESIGN.md](DESIGN.md) for the interface contract.

## Administrator setup

Administration is decided by Foundation's persistent site authorization, independently of Identity roles.
On an unclaimed instance, sign in and open `/admin` (redirects to `/setup`), then explicitly claim the site.
`PASTE_ADMIN_SUBS` seeds protected administrators only when initializing authorization; removing the environment value does not revoke an existing grant.
Apply migrations 0004 (authorization) and 0005 (audit) before starting the updated API.
`PASTE_INSTANCE_ID` identifies the authorization instance; keep it stable across restarts.

The public setup endpoint is `GET /api/v1/authorization/setup`; claiming requires a real user at
`POST /api/v1/authorization/setup/claim`. Public user profiles are read through the shared Identity BFF
at `/identity-api/api/v1/users`; Paste stores no duplicate user directory.

## License

Apache-2.0.

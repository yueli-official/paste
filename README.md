# Paste

Paste is the Yueli site group's anonymous-first, multi-file code sharing product. It uses a Go API, a Nuxt Web app, Foundation
contracts and Identity for optional authenticated ownership.

The repository is being rebuilt from the useful product behavior in the legacy CodeShare implementation. Rust, SQLite, local
password handling and the old visual treatment are not compatibility contracts.

## Planned repository shape

```text
api/                 Public request and response types
cmd/paste/           Go API entrypoint
internal/paste/      Paste lifecycle and access rules
internal/postgres/   PostgreSQL adapter and migrations
internal/httpapi/    HTTP adapter
web/                 Nuxt product app and Playwright acceptance
flightdeck/          Resumable repository work
```

Development commands will be documented when the first runnable slice lands. Local lifecycle is owned by the adjacent Workspace
CLI rather than ad-hoc background processes.

## License

Apache-2.0.


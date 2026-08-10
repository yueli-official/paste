# Plan

## P0 — Product and repository contract

- [x] Confirm anonymous/authenticated behavior, multi-file first release and public repository visibility.
- [x] Create the public repository, AGENTS, PRODUCT, README and resumable Flightdeck.
- [ ] Add toolchain, licensing and Workspace Repository Lock contracts.

## P1 — Go lifecycle

- [ ] Implement Paste creation, access, update, expiry and deletion semantics with memory tests. Creation, access and owner
  listing are complete; update/delete remain.
- [ ] Add PostgreSQL schema and adapter integration tests.
- [ ] Add Foundation HTTP/auth adapters and deterministic public errors.

## P2 — Nuxt product surface

- [ ] Establish the visual world and creation/viewing surface brief.
- [ ] Build the multi-file editor and public/protected/terminal viewing states.
- [ ] Add Identity-backed history and owner management.

## P3 — Workspace and acceptance

- [ ] Add Repository Lock entry and Environment Contract with shared Identity.
- [ ] Pass Go, Nuxt, unit, integration, typecheck and production build gates.
- [ ] Run CLI Playwright desktop/mobile acceptance against the real local composition.

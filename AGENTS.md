# Repository instructions

- When adjacent `../workspace` exists and its `repos.lock.yaml` includes this repository, read and follow
  `../workspace/docs/multi-project-development.md` before cross-repository work or any local process lifecycle action. Product
  sessions treat Workspace contracts and `.doctor/` as read-only, use the Workspace CLI for generated state, use a separate Git
  worktree for concurrent writes to this repository, and never own shared Provider processes.
- Route long-running or resumable work through `flightdeck/deck.md` and the focused
  `flightdeck/work/*/index.md`; keep the Markdown handoff aligned with repository reality.
- This repository owns the Paste product, its Go API, Nuxt Web app and published consumer contracts. Do not move product
  implementation into the Workspace orchestration repository.
- Identity is the only user identity authority. Anonymous Paste creation must not create a shadow account or local credential.
- Web acceptance must use the repository CLI Playwright suite against the real local composition, including desktop and mobile
  rendering plus creation, protected access, copy/share and authenticated management flows.
- Local backend listeners bind to loopback. A frontend may listen on a non-loopback address only when the Workspace Environment
  Contract explicitly declares LAN exposure.
- Interface icons come from Tabler through Iconify; do not introduce a second icon family.


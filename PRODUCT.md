# Product

<!-- impeccable:product-schema 1 -->

## Platform

web

## Users

- Developers and technical collaborators create a shareable code artifact while debugging, reviewing or handing work to
  another person. They need the first useful link without being forced through registration.
- Authenticated Yueli users return to their own Pastes to find, edit, expire or delete them.
- Recipients open a Paste link, understand its files and language immediately, and copy one file or the complete Paste without
  needing an account.
- Explicitly configured Paste site administrators govern full-site records and public presentation settings. Identity control-plane
  administration does not implicitly grant this product-local role.

## Product Purpose

Paste turns one or more small text files into a durable, readable share link. Success means creating a Paste is faster than
preparing an attachment, protected content never leaks through its URL, and a recipient can inspect or copy the material without
learning the product first.

## Positioning

Paste combines an anonymous-first sharing path with authenticated continuity inside the Yueli site group. A Paste may contain a
small ordered file set rather than forcing every share into one undifferentiated code block, while the public reading experience
remains as direct as a single-file paste.

## Operating Context

- The primary action begins with code already on a clipboard or in an editor.
- Anonymous creators receive a share link but no management record. Signing in is required for personal history, editing and
  deletion.
- Recipients commonly arrive from chat, an issue, an email or a short link on desktop or mobile.
- Yueli Identity supplies OIDC login. The Paste API and Web app use the current Foundation runtime and Workspace orchestration.

## Capabilities and Constraints

- The first release supports one to twenty ordered text files per Paste. Every file has a display path, language and content.
- A Paste has a title, optional description and optional tags. Server-owned size and count limits fail explicitly.
- Unlisted is the default visibility. Private Pastes require an authenticated owner and authenticated access.
- A Paste may have an expiry and may be protected by a password. Passwords are never placed in share URLs, logs or plaintext
  storage.
- Anonymous Pastes are immutable after creation. Authenticated owners may list, edit, expire and delete their own Pastes.
- A creation suspension applies to one authenticated subject and blocks only new Paste creation. It is not an Identity account
  ban and does not delete, disable or otherwise change that subject's existing shares.
- Signed-in creation quotas count successful creations per subject over a UTC calendar day. The site-wide default is 1–10000 and
  an administrator may set an independent per-user override.
- Anonymous creation uses one shared site-wide UTC daily total, not an IP-, device- or inferred-user identity. Its configured range
  is 0–100000; zero pauses anonymous creation for everyone.
- Only a successfully committed Paste consumes quota. Counter consumption and Paste insertion are one atomic database operation,
  so rejected or rolled-back attempts are not counted.
- Exhausted quotas return `429 common.rate_limited`; creation suspension returns `403 paste.creation_suspended`; a disabled
  anonymous pool returns `403 paste.anonymous_creation_disabled`.
- Administrator user batches change only Paste-local creation state or daily quota. They never delete Identity users; partial
  failures remain explicit and retryable.
- Display and governance settings have independent revisions. A unified save may partially succeed and must identify which group
  failed rather than presenting stale settings as one atomic result.
- Administrator summaries never return code content or password material. Batch modification and deletion report partial failures
  instead of presenting a false all-or-nothing result.
- Expired and deleted locators are terminal states and are not silently reallocated.
- Syntax highlighting, per-file copy, whole-Paste copy, share-link copy and responsive keyboard operation are first-release
  behavior.
- Public discovery, comments, collaborative live editing, executable sandboxes, binary uploads and commercial claims are outside
  the first release.
- The production hostname remains an open deployment decision.

## Brand Commitments

- The user-facing site name defaults to `代码片段`; a Paste site administrator may change it and the public site description.
- Repository, service, API, route and persistence identifiers remain `paste`; editable presentation never renames technical
  contracts.
- Product copy is concise Chinese and speaks in familiar developer terms.
- Interface icons use Tabler only.
- The old CodeShare purple landing page is source evidence, not a visual authority.

## Evidence on Hand

- Legacy repository `E:\projects\yozya\code-paste` at clean commit `ccf235f` contains working Rust/Vue behavior and screenshots
  for single-file creation, password access, expiry, syntax highlighting, copy and share.
- Foundation, Identity and Workspace contracts already exist in adjacent Yueli repositories.
- There are no verified customer counts, availability figures, testimonials, pricing claims or production traffic measurements;
  surfaces must not invent them.

## Product Principles

- The shortest path creates a useful share, not an account.
- Protection is a server contract, never a secret embedded in a URL.
- Multiple files should read like one compact artifact, not a miniature file manager.
- Ownership stays explicit: anonymous is immutable; authenticated ownership is manageable and auditable.
- Terminal lifecycle states are honest and stable.

## Accessibility & Inclusion

Creation, file switching, protected access, copy/share and owner management must work by keyboard, retain visible focus, expose
status in text rather than color alone, and remain usable on narrow mobile viewports. Code typography may be monospaced, but all
controls and explanatory content must remain readable under browser zoom and user font scaling.

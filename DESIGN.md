# Paste interface contract

Paste is an editor-first workbench, not a marketing page or a miniature file manager. The creation and editing route (`/`) begins
directly with a full-height editor shell: ordered file tabs, active code, transient share controls and document status. It does not
place a hero, explanatory copy or a separate public website Header ahead of the task. Recipients get the same file vocabulary in a
quieter, read-only surface; authenticated users get the same application shell around a compact management ledger.

## Visual language

- Cool porcelain canvases, mist-blue chrome and ink-navy dark surfaces give the editor a focused technical atmosphere without
  falling into generic charcoal. Mineral blue owns selection, focus, commit actions and the editor status bar; red and green stay
  reserved for destructive and successful states.
- Borders establish structure. Shadows are limited to genuinely elevated overlays, while the workbench itself remains flat.
- Interface text uses the system Chinese sans-serif stack. Monospace typography is reserved for source code and file-oriented
  metadata.
- Tabler is the only icon family. Icons support labels and do not replace unfamiliar action names.

The canonical tokens live in `web/app/assets/css/main.css`. Light and dark modes must preserve the same hierarchy, rather than
reinterpreting the layout with unrelated colors or effects.

## Layout and interaction

- Creation and editing on `/` use one edge-to-edge editor shell. A single editor chrome merges the product entry, horizontal file
  tabs, share, My Paste, theme and account controls into one 40px desktop row; it is the route's only title and
  navigation layer. The code surface consumes the remaining viewport above a 28px status bar, which owns the language selector.
- Sharing opens from the editor chrome in a transient Nuxt UI slideover and never permanently consumes editing width.
- File tabs support direct switching, drag reordering, `Alt + ←/→` keyboard reordering, double-click rename and direct deletion
  from the tab close control. Inline rename uses the tab's flat editor-native field rather than a rounded form control, and adding
  a file does not force rename mode. On
  narrow viewports the editor chrome becomes 44px and the status bar 30px; the product entry collapses to an icon, My Paste moves
  into the account context, and the horizontally scrolling file tabs remain the primary file carrier.
- The authenticated `/mine` route uses the same edge-to-edge application shell: one editor-like titlebar, one flat search toolbar,
  an internally scrolling file ledger and a readable status bar. It does not render the shared public Header or card-page hero.
- The shared `/p/:code` route is the read-only expression of that same editor shell. Its desktop frame is one 40px file bar, a
  full-height code surface and a 28px status bar; narrow viewports use the matching 44px file bar and 30px status bar. The file bar
  carries file selection plus the primary read, copy and inspect actions, without adding a page title, card wrapper or second
  navigation row.
- Shared-Paste metadata and secondary copy actions live in a transient information rail opened from the file bar. The rail may
  explain title, description, tags, visibility, expiry and file details, but it must not reserve permanent width beside the code
  or make reading contingent on opening it.
- `/mine` supports explicit row selection and selection of the current filtered result set. Once anything is selected, the search
  toolbar becomes a contextual command bar for batch visibility/expiry changes, destructive deletion and clearing selection.
  Batch changes patch only the chosen fields, run with bounded concurrency and keep failed items selected after partial failure;
  they must never overwrite file content, passwords or unrelated metadata.
- The protected `/admin` route is an editor-native Operate master-detail ledger, not a card dashboard. Compact application chrome
  leads to full-site governance as the primary workspace and site settings as a deliberate secondary workspace. On desktop, the
  safe record inspector shares the work plane with the ledger; at narrower widths it becomes an overlay, and on mobile it occupies
  the full work plane.
- User governance within `/admin` searches exact or partial user subjects and filters active or creation-suspended states. Its
  ledger keeps today's successful creations, effective quota, active and lifetime Paste counts, and latest creation visible;
  administrators can suspend or resume creation directly without confusing the action with an Identity account ban. Rows support
  explicit selection and current-page selection; selection turns the toolbar into batch state and daily-quota commands.
- User batch operations report partial success, keep failed subjects selected for retry and never imply deletion of an Identity
  user.
- The user inspector edits creation state, a per-user quota override and governance notes, and links into that subject's Paste
  records. It occupies the desktop side rail and becomes the complete governance plane at 320px; access state remains visible and
  mobile actions retain targets of at least 44px.
- Admin rows keep access state visible at mobile widths instead of hiding it as expendable table metadata. Mobile navigation,
  selection, filtering, paging and inspector actions provide actionable targets of at least 44px.
- Admin batch modification and deletion report partial success and keep failed records available for retry. The inspector is a
  governance summary only: it never returns code content or password material.
- Site settings use one command bar over a flat settings catalog, without a decorative preview pane. The display group edits the
  public site name and description; the governance group edits the signed-in users' default daily creation limit (1–10000) and
  shared anonymous daily total (0–100000). Each group exposes its own revision. Unified save reports partial success by group, and
  the mobile command bar keeps abandoning unsaved changes reachable. Presentation values may replace the user-facing product
  label, but do not rename the technical `paste` identity, API or routes.
- Other public routes retain the shared public Header. Do not reintroduce that Header—or any second titlebar—above the application
  shell on `/`, `/mine`, `/p/:code` or `/admin`.
- A Paste contains 1–20 ordered text files. Multiple files should feel like one shareable artifact, not a directory browser.
- Passwords are submitted in the protected access form and never appear in a query string or copied share URL.
- A single file may use the full 1 MiB Paste budget; all files together remain limited to 1 MiB. The status bar shows the active
  file against that real boundary, client validation names an offending file before submission, and API validation preserves
  field violations.
- Anonymous creators receive a usable link immediately. Identity login adds history and management but must not obstruct public
  creation or reading.
- Every control retains a visible focus state, text status and a practical touch target. Reduced-motion preferences suppress
  nonessential transitions.

## Reuse boundaries

Foundation supplies shared runtime, Identity integration and common UI primitives. Paste owns its workbench composition, code
editor and product-specific states. Standard form controls—including inputs, textareas, tags and dropdowns—use Nuxt UI; custom
controls are reserved for editor-native interactions such as the file explorer and code surface. New pages should reuse the
established tokens and shell before introducing another card, navigation pattern or visual motif.

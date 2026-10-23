# Paste interface contract

Paste is a fast text and code sharing workspace. The core path is paste content, optionally adjust access, and generate a link.
The public editor, reader and content library share a graphite-and-lime visual language with a global account control in the top
right and a persistent left content navigator on desktop.

## Visual language

- Dark mode is the initial preference: cold graphite surfaces, restrained separators, readable muted text and one lime primary
  action. A cool neutral light mode retains the same hierarchy.
- The serif `paste.` wordmark is the product signature. The configurable site description appears beside it as a tagline; an empty description hides it. The site name remains in page
  metadata. The theme switch and shared Identity account control live at the top right on every workspace route.
- System Chinese sans-serif type carries interface copy; monospace is reserved for file names, source code and technical metadata.
- Tabler is the only icon family. Labels explain unfamiliar actions; compact mobile icon actions retain accessible names.
- Product theme tokens live in `web/app/assets/css/main.css`. Layouts and component states belong in template Tailwind utilities.

## Creation and editing

- `/` opens directly into an empty `untitled.txt` file with focused CodeMirror input. A title is optional. Generating a link is
  disabled until at least one file contains content; anonymous creation remains available without login.
- The left navigator contains a secondary New action, library search, the current editing state and real recent owned content.
  Anonymous users see a sign-in action without promotional filler. Mobile navigation opens from the header.
- A document command row contains the optional title, one access-and-expiry summary, and the primary Generate link / Save changes
  action. The summary opens one sharing panel for visibility, expiry, password, description and tags. Sharing does not require
  visiting that panel first.
- There is no automatic clipboard read or implied autosave. Current content is marked unshared or unsaved; abandoning changed
  content requires confirmation. Anonymous creators are informed that their content cannot subsequently be edited.
- A Paste contains 1–20 ordered text files and a combined 1 MiB content budget. File tabs retain switching, drag reordering,
  `Alt + Left/Right` reordering, double-click rename and deletion. File names infer languages; selecting a language updates the extension while preserving the path stem and avoiding duplicate names.
- Plain text and code use the full-width CodeMirror canvas with language selection, wrapping and raw-copy controls. Source syntax
  colors are theme-aware. The editor status row reports the real total byte budget and file count.
- Markdown uses the shared Foundation ContentEditor (Tiptap) for formatted editing, CodeMirror for source editing, and shared
  ContentProse for preview. At widths of at least 1100px, Markdown starts with editing on the left and live preview on the right.
  Edit / Split / Preview controls choose the layout; the source/formatted menu chooses the editor. Narrow screens default to
  editing and offer a separate Preview view.
- Markdown starts in source mode. Switching a nonempty source into formatted editing explains possible normalization of HTML,
  custom syntax and whitespace when edits are made. Merely opening preview or formatted mode must preserve the original source
  byte for byte; the API continues to receive Markdown strings. Shared editor draft storage is disabled for this workspace.
- Markdown body content has a bounded reading width; code keeps the available workspace width. Mobile tool rows keep file tabs
  and editing actions accessible without overflowing the page. The shared rich editor owns its own scrollable formatting toolbar.

## Reading and library

- `/p/:code` uses the same shell, a clear title, one primary Copy link action, file tabs and raw-copy controls. Markdown opens in
  reading mode and can reveal its original source. Code remains read-only.
- Authenticated ownership is verified before showing Edit and Delete on the reading page. Deletion requires confirmation,
  uses the current revision, and returns to the library after success.
- Successful creation or saving displays an inline confirmation and selectable share URL. Copy success appears at the action;
  no success Toast is emitted. Details and multi-file copy actions remain in the information panel.
- Password access, loading and unavailable-content states retain the surrounding workspace. Passwords never enter the URL.
- `/mine` is named 内容库. It retains server search, pagination, explicit row selection, batch visibility/expiry changes and
  deletion. Batch changes affect only selected fields; partial failures remain selected and never overwrite file content.
- The library uses compact file cards on graphite surfaces, with language, file count, access state and update date grouped
  beneath the title. Long titles truncate inside the sidebar; file tabs leave room for letter descenders.
- The future website-cluster launcher belongs beside the top-right account control. Do not display an inert launcher or invent
  destinations before a real site navigation contract is supplied.

## Administration

The protected administration area continues to use the shared Foundation admin shell and existing governance flows.

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

## Reuse and ownership

Foundation owns the shared content renderer/editor, runtime, account integration and common controls. Paste owns its composition,
CodeMirror adapter, file model and sharing states. Consume the Workspace-approved packages and verify frozen artifacts before
claiming a production build; local source overlays are a separate integration check.

- Ordinary layout, spacing, sizing, responsive behavior and states belong in Vue Tailwind utilities, Nuxt UI configuration or
  shared Foundation modules. Product pages do not own `<style>` blocks.
- Main CSS owns semantic light/dark tokens, document base/focus behavior, minimal CodeMirror host rules and reduced-motion behavior.
- Keep administrator save feedback, compact collections and pagination within the existing shared contracts. The new public shell
  must not replace or duplicate the administration shell.

## Verification

CLI Playwright covers real desktop/mobile creation, protected access, copy/share, login and owned content management. The redesign
suite additionally checks direct creation, complete multiline clipboard input, exact Markdown preservation across mode switches,
and content created in the shared formatted editor. Screenshot collection is not visual approval: inspect actual rendered states,
fix alignment/spacing/overflow issues and inspect again before delivery.

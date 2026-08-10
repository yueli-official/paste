# Paste interface contract

Paste is an editor-first workbench, not a marketing page or a miniature file manager. The primary surface keeps the ordered file
set, active code and publication settings visible as one continuous task. Recipients get the same file vocabulary in a quieter,
read-only surface; authenticated users get a compact management ledger rather than a second product shell.

## Visual language

- Warm paper page backgrounds and near-white working surfaces connect Paste to the Yueli product family without making code feel
  decorative.
- Mineral blue is the single primary accent for selection, focus and commit actions. Red and green are reserved for destructive
  and successful states.
- Borders establish structure. Shadows are limited to genuinely elevated overlays, while the workbench itself remains flat.
- Interface text uses the system Chinese sans-serif stack. Monospace typography is reserved for source code and file-oriented
  metadata.
- Tabler is the only icon family. Icons support labels and do not replace unfamiliar action names.

The canonical tokens live in `web/app/assets/css/main.css`. Light and dark modes must preserve the same hierarchy, rather than
reinterpreting the layout with unrelated colors or effects.

## Layout and interaction

- Desktop creation uses three connected regions: file rail, code editor and publication rail. This keeps switching, editing and
  publishing within one scan path.
- Narrow viewports convert the file rail to a horizontal chooser, retain a substantial editor viewport and stack publication
  controls below it.
- A Paste contains 1–20 ordered text files. Multiple files should feel like one shareable artifact, not a directory browser.
- Passwords are submitted in the protected access form and never appear in a query string or copied share URL.
- Anonymous creators receive a usable link immediately. Identity login adds history and management but must not obstruct public
  creation or reading.
- Every control retains a visible focus state, text status and a practical touch target. Reduced-motion preferences suppress
  nonessential transitions.

## Reuse boundaries

Foundation supplies shared runtime, Identity integration and common UI primitives. Paste owns its workbench composition, code
editor and product-specific states. New pages should reuse the established tokens and shell before introducing another card,
navigation pattern or visual motif.

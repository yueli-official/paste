---
version: 1
slug: "web-app-pages-admin-vue"
primary_target: "web/app/pages/admin.vue"
related_targets: ["web/app/assets/css/main.css","web/app/layouts/default.vue"]
---

## THESIS

Site governance is a focused inspection desk: administrators search the whole service, select precise records, act in batches, and move to site presentation settings without leaving the editor-native workspace.

## OWN-WORLD

Use the existing cold porcelain and ink-navy editor world, mineral-blue chrome, flat one-pixel boundaries, compact monospace metadata, and the same light/dark behavior as the composer and reader.

## STORY

The primary path is search or filter, scan ownership/access/state, select records, then modify or delete. Inspecting one record reveals a safe summary only. Site settings are a deliberate secondary workspace for public name and description.

## FIRST VIEWPORT

Desktop shows compact top chrome, a two-item section rail, toolbar, dense ledger, optional summary inspector, and fixed status bar. Mobile keeps the same order as stacked rows with permanent selection controls and no horizontal page overflow.

## FORM

The structure is the fourth grounded Operate candidate chosen from concept seed `eb959727`: a master-detail explorer, not a card dashboard. Governance owns the broad working plane; settings use one restrained form without a decorative preview pane.

## FINISH

Keep radii small and reserved for controls, never container decoration. State colors must remain readable in both themes. Transient batch controls use Nuxt UI overlays, keyboard focus returns to the trigger, and no summary exposes code content or password material.

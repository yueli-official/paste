# Context

- Legacy source is `E:\projects\yozya\code-paste` at clean commit `ccf235f`.
- Repository and product name are `paste` and `Paste`.
- The GitHub repository is public at `https://github.com/yueli-official/paste`.
- Anonymous users may create but do not receive owner management. Authenticated users manage only their own Pastes.
- First release supports one to twenty ordered text files and does not execute submitted code.
- Passwords never travel in URLs or persist as plaintext. Private Pastes require Identity ownership.
- Product implementation stays in this repository. Workspace owns local composition and generated runtime state.
- UI direction is the editor-first continuous workbench documented in `DESIGN.md`: ordered file rail, code surface and publish
  rail on desktop, with horizontal file switching and stacked settings on narrow viewports.
- Local acceptance uses the shared Identity provider, Paste API on `127.0.0.1:8091` and Paste Web on `localhost:3010`.

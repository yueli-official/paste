# Repository instructions
AGENTS version: 1.0.2

通用约束遵循用户级根 AGENTS.md；本文件只保留本仓库职责和入口。

- When adjacent `../workspace` exists and its `repos.lock.yaml` includes this repository, read and follow
  `../workspace/docs/multi-project-development.md` before cross-repository work or any local process lifecycle action. Product
  sessions treat Workspace contracts and `.doctor/` as read-only, use the Workspace CLI for generated state, use a separate Git
  worktree for concurrent writes to this repository, and never own shared Provider processes.

- This repository owns the Paste product, its Go API, Nuxt Web app and published consumer contracts. Do not move product
  implementation into the Workspace orchestration repository.
- Identity is the only user identity authority. Anonymous Paste creation must not create a shadow account or local credential.
- Web acceptance must use the repository CLI Playwright suite against the real local composition, including desktop and mobile
  rendering plus creation, protected access, copy/share and authenticated management flows.
- Local backend listeners bind to loopback. A frontend may listen on a non-loopback address only when the Workspace Environment
  Contract explicitly declares LAN exposure.
- Interface icons come from Tabler through Iconify; do not introduce a second icon family.

- 生产部署必须先读取[前后端配套发布合同](../workspace/docs/deployment-policy.md)，以 API/Web 同一源码快照、双端清单和 `deploy/verify_release.py` 完成门禁。

- 本产品全部页面使用 SPA（`ssr: false`），由用户于2026-11-14明确批准。这是共享后台SSR默认值的产品例外。保留现有样式、Nitro认证/BFF及后端权限验证。

## 共用入口

- 共用规则、技能和知识的归属见 [Workspace 内容归属](../workspace/docs/knowledge-placement.md)；产品实现与特有操作留本仓库，不复制根流程。
- 代码部署使用 [Workspace 部署技能](../workspace/.agents/skills/yueli-deploy/SKILL.md)和[发布合同](../workspace/docs/deployment-policy.md)，服务参数与更新命令仍由本仓库拥有。
- 使用共享 UI 前，先读实际前端目录中已安装的 `node_modules/@yueli/ui/AGENTS.md`，按该包执行 inspect、search/docs；不拿邻仓新源码冒充已安装接口。

## Flightdeck

长期工作使用已安装的 Flightdeck 技能，从 `flightdeck/deck.md` 恢复。`/flightdeck/` 保持 Git 忽略，维护工作台不授予提交、推送或发布权限。

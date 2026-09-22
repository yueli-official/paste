# Paste 开发者令牌

## Goal

让用户可以在 Account 创建绑定 `paste-yueli-web` 的细粒度开发者令牌，通过 Paste 现有 API 创建片段并读取、编辑、删除自己的片段；当前 Paste 管理员可额外选择片段治理与公开站点设置权限。每次请求都重新验证 PAT，并取令牌 scope 与账号当前 Paste 权限的交集。

完成时应满足：普通用户只能申请和使用自己的片段权限；管理员 scope 不会授予非管理员；首位管理员认领、管理员会话、用户额度/封禁与反滥用治理不向 PAT 开放；显式 PAT 失败不回退浏览器 Cookie；产品指南、OpenAPI、运行配置和真实本地浏览器/API 验收一致。

## Status

待查收

## Current

2026-09-22 本地实现与真实组合验收已完成。Paste 已升级到 Foundation Go `v0.5.0`，接入 PersonalTokenVerifier、可信权限目录、六项细粒度 capability 与显式路由 allowlist；普通用户只有自己的片段权限，当前管理员额外获得片段治理与公开站点设置，认领、管理员会话、用户治理和反滥用设置保持拒绝。

持久授权目录已从 v1 升级到 v2，并新增 `0006_authorization_catalog_v2` 迁移。真实 PostgreSQL 集成测试覆盖 v1 digest 初始化、执行升级、现有管理员获得新 capability、普通用户仍不能治理；避免生产已有数据库因 catalog digest 漂移而拒绝启动。

验证通过：`go test ./...`、`go vet ./...`、20 项 HTTP Result/OpenAPI freshness、PAT 专项严格 TypeScript 检查；Workspace Isolated 最终 Session `20260922T053447Z-43064` ready，CLI Playwright 桌面/手机 2/2 通过，覆盖 Account 六项权限、UI 创建令牌、Paste CRUD、管理员查询与设置、敏感路由拒绝、撤销立即失效和显式 PAT 不回退 Cookie。桌面/手机权限截图已目检，隔离组合验收后已停止，既有 Commerce Session 未受影响。

Workspace 全量 environment 测试被既有 Distribution 合同失败阻断：`commerce browser product must bind identity.oidc with oidc-clients provisioning`。该失败与本轮 Paste Environment 无关；Paste 实际 prepare/up、迁移、五进程 ready 与浏览器闭环均通过，本轮没有修改 Distribution 旁支。

本 Work 只负责 Paste。BVideo 已完成本地 PAT 交付但尚未生产部署；Gallery、Nav、Distribution 后续分别建立单站 Work，不在这里混合实施。当前等待用户确认 Paste 单站交付完成；生产部署不在本 Work 范围。

## Next

等待用户查收 Paste 开发者令牌交付；收到反馈后处理对应问题。2026-09-22 用户已授权本地提交，尚未推送或生产部署。

## References

- [上下文与权限边界](context.md)
- [实施与验收计划](plan.md)
- [Foundation 个人令牌合同](../../../../foundation/flightdeck/knowledge/authorization/personal-access-tokens.md)

# Paste Go 重建

## Goal

以 Go、Nuxt、Foundation、Identity 和 Workspace 合同交付可运行、可测试、可验收的多文件 Paste 产品，替代旧
CodeShare 实现而不继承其技术债和无证据内容。

## Status

Open

## Current

公开 GitHub 仓库 `yueli-official/paste` 已创建并克隆。产品事实已确认：匿名可创建但不可管理；登录用户拥有历史与
管理能力；首版支持一个 Paste 内多文件。PRODUCT 已记录真实能力、非目标和无证据限制。

首个 Go 领域切片已落地：`internal/paste` 用 Foundation `CompactURLV1` 与 UUIDv7 建立公开定位符和内部标识，固定
1–20 文件、路径/标签去重、1 MiB 总内容上限、unlisted/private、bcrypt 密码、过期和 owner-only private 访问。
内存 Store 隐藏可变状态并支持 owner 列表；六组测试覆盖匿名多文件、private owner、密码不泄漏、过期、非法文件
与 owner 过滤，`go test ./...` 和 `git diff --check` 通过。

## Next

提交并推送初始仓库基线；用该 revision 接入 Workspace Repository Lock。随后补齐更新/删除生命周期和 PostgreSQL
adapter。

## Progress

- 2026-08-11：创建公开仓库，确认匿名/登录与多文件范围，写入 PRODUCT、AGENTS 和 Flightdeck；完成首个带测试的
  Go 领域 module 与内存 Store。

## References

- [稳定上下文](context.md)
- [执行计划](plan.md)
- [产品事实](../../../PRODUCT.md)

# Paste Go 重建

## Goal

以 Go、Nuxt、Foundation、Identity 和 Workspace 合同交付可运行、可测试、可验收的多文件 Paste 产品，替代旧
CodeShare 实现而不继承其技术债和无证据内容。

## Status

Finished

## Current

多文件 Paste 已完整交付：GoFrame API、PostgreSQL schema/store、Foundation Problem/trace/rate-limit/readiness、
Identity JWT 所有权，以及匿名创建、密码访问、登录用户列表/编辑/删除的生命周期均已落地。Nuxt 端提供 editor-first
工作台、公开/受保护阅读页和个人管理页，并复用 Foundation UI 与 Identity Account Control。

Workspace `paste-local` 组合已通过共享 Identity provision，API `127.0.0.1:8091` 与 Web `localhost:3010` ready。Go
测试与 vet、Web unit/typecheck/build 已通过；CLI Playwright 已真实覆盖桌面/移动匿名创建、密码错误/正确访问、无密码
URL、OIDC 登录后的创建、列表、编辑和删除，公开页面 axe 扫描无违规。

## Next

None.

## Progress

- 2026-08-11：创建公开仓库，确认匿名/登录与多文件范围，写入 PRODUCT、AGENTS 和 Flightdeck；完成首个带测试的
  Go 领域 module 与内存 Store。
- 2026-08-11：完成 PostgreSQL、HTTP/Identity、Nuxt workbench、Workspace 组合与真实 Playwright 验收。

## References

- [稳定上下文](context.md)
- [执行计划](plan.md)
- [产品事实](../../../PRODUCT.md)

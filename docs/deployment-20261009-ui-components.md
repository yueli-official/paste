# 共享 UI 组件部署记录

2026-10-09 用户授权推广；正式入口 https://paste.yuelili.com。仅更新 Web 至 `yueli/paste-web:server-20261009-ai-components-2`，healthy、RestartCount=0；API 容器 ID 和运行镜像保持原值，未执行迁移或生产管理写入。

源码提交 `82aef7112b0abe5bac0b9f945cf5c13737b5d4e6`；镜像 ID `sha256:0e0a02667c15f15284d86739aa1b5a7070b7d324ece178d0707806b35c48d79f`；Web 制品 SHA-256 `9d2996c78928b16cffe9f3dcc6f838112ad098adeb53d2691c9b4b6c113890f0`。

固定采用 @yueli/ui `0.8.0-preview.20261009.ai-discovery.6`，SHA-256 `06407c95454b215c61b5cf3ef432476de6677d3a78542b07fe706a7c52106c0a`。独立 Git archive 冻结安装、前端单测、类型、生产构建和已安装内容校验通过；源码覆盖为空。原版私有发布依赖字节对照后纳入 vendor，仅更改获取方式。

备份 `/projects/yuelili.com/backups/ui-components-20261009-2/paste`：数据库 custom dump 经 pg_restore --list 核对；原 Compose、配置、迁移和运行身份可追溯。制品及服务器回执 `/projects/yuelili.com/.deploy/ui-components-20261009-2/`。
CLI Playwright light/dark × 390/1440 共 4 个生产场景通过，包含页面、图片（如有）、公开交互和匿名登录门禁；无横向溢出、pageerror、hydration mismatch 或站点 5xx。没有生产管理员会话，认证后的管理行为采用本站已有隔离组合证据，不宣称线上写入通过。

本机证据 `C:/Users/yl/.codex/task-artifacts/sites-ui-upgrade/20261009-2`：production-result.json、deployment-result.json、source-manifest.json、截图及原依赖对照报告。临时源码与 node_modules 已限定清理；保留可追溯制品、日志和验收证据。未 push、发布 SDK 或改动 Distribution。

# 2026-10-16 前端依赖基线部署

已部署 [paste](https://paste.yuelili.com)：`yueli/paste-web:server-20261016-dependency-baseline-1`，镜像 `sha256:76bdee53d90e6c0c828fa749b2b0364d65278cba7e884b4b6430e2d0a2495d0e`。

本轮在实际线上源码快照上更新批准的共享前端包组合，保留已上线页面修复，排除工作区其他未上线功能。用户授权先部署查看，再决定提交；没有 commit、push 或公开包发布。

- 基线：`0556d5134f0d158ec40fee5a4feb5519bc632feb6f95c065c99d6f7d4f832079`；共享包各自保留版本，以版本及最终 tarball SHA-256 准入。
- 完整源码快照：`fddc9ef8ad11bd31f45d81abcc2de4051734bf41fa4a7bd6cbf10836fe69586e`；文件摘要、API/Web来源、制品与依赖回执在服务器 `/projects/yuelili.com/.deploy/paste-dependency-baseline-20261016-1/release.json`。
- API复用 `yueli/paste-api:server-20260911-redeploy-1`，来源文件、运行镜像及原交付制品中的二进制摘要匹配；API容器保持，无迁移、无生产密钥轮换。
- 候选与正式域名 CLI Playwright：390/1440、明暗主题、渲染/水合、无溢出、无未捕获脚本错误及5xx，匿名管理入口正确进入Account。正式制品关键JS逐文件SHA验证见同目录证据。本轮没有生产管理员会话，未执行生产管理写入或真实交易。
- 配置与数据库备份：`/projects/yuelili.com/backups/paste-dependency-baseline-20261016-1`，数据库备份已验证可读；已有持久配置、API与其他容器保持。

构建范围：本轮固定生产快照重新执行冻结安装、prepare、typecheck与production build。Linux候选、实际运行输出及正式HTTPS分别验证，不能把本地验收当作线上管理员交互证据。

本地证据：`D:/coding/yueli-official/task-artifacts/workspace/dependency-baseline-20261016/production/paste`。新站接入及后续升级遵守 [Workspace前端依赖基线](../../workspace/docs/frontend-dependency-baseline.md)。

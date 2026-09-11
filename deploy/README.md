# Paste 服务器部署

目标 `https://paste.yuelili.com`，目录 `/projects/yuelili.com/paste`。Compose 运行独立 PostgreSQL18、一次性迁移、Paste API/Web，复用 `yueli-services` 中的 Identity；头像由已有 Account/Asset 负责。

- `data/postgres` 持久化全部 Paste 数据；`releases/<version>` 保存镜像输入、源码快照、SHA256和发布脚本，`backups` 保存切换前代理配置。
- `.env` 仅在服务器生成，0600，保存数据库密码、连接串和稳定的会话密钥，不输出或提交。
- Web仅主机回环13010，API回环18291；数据库不发布端口。OpenResty现有域名/TLS代理Web。
- OIDC客户端 `paste-yueli-web`，callback `https://paste.yuelili.com/auth/callback`，登出回调正式根路径，audiences `paste-api`、`identity-api`。
- 不配置本地测试账号为管理员，不迁移本地样本。首次登录 `/admin` 后显式认领。
- 镜像基于验收源码的Linux amd64静态Go二进制和Nuxt production `.output`；非root运行，带健康检查、内存与日志限制。

2026-09-10 首次制品 `yueli/paste-api:server-20260910-1` 与 `yueli/paste-web:server-20260910-1`；使用本地已验收 Foundation/Identity源码组合，不宣称新的上游正式Release已发布。证据与最终上线结果见本Work的 server-deployment.md。

# Paste 首次正式部署 — 2026-09-10

用户确认本地验收通过，并授权部署 https://paste.yuelili.com，创建 /projects/yuelili.com/paste 保存部署和数据。部署成功，未执行 Git 提交、推送或上游包发布。

## 运行合同

- 发布：server-20260910-1；镜像 yueli/paste-api:server-20260910-1、yueli/paste-web:server-20260910-1。
- 服务器目录：/projects/yuelili.com/paste；Compose 与 .env 位于目录根部，.env 权限0600。
- 独立 PostgreSQL 18.4，数据库 paste、用户 paste_app；data/postgres 挂载 /var/lib/postgresql，不开放主机端口，使用内部 paste_private 网络。V1–V5 迁移已执行。
- API 仅回环18291→8091，Web仅回环13010→3000；三个常驻容器均 healthy。
- 复用线上 Identity、Account、Asset；未重启或升级其他站点与共享服务。
- 使用既有 paste.yuelili.com TLS 与 OpenResty站点，root.conf 代理改为127.0.0.1:13010；nginx -t通过后reload。
- Identity devprovision仅注册 paste-yueli-web，回调 https://paste.yuelili.com/auth/callback，登出回调根路径，audiences为paste-api和identity-api；未创建测试用户或授予管理员。

## 制品与备份

releases/server-20260910-1 保存API/Web归档、源码快照、逐文件SHA256清单、部署脚本和production.json；归档校验后构建镜像。制品包含已验收的本地Foundation/Identity源码组合，不能解释为上游新版本已正式发布。

backups/root.conf.before-server-20260910-1 保存切换前代理配置。backups/initial-deployment.sql.gz 为清理验收样本后的初始数据库备份，权限0600；本次未配置定时备份。

本地证据：E:/tmp/yueli-paste-deploy-20260910/production.json、home-1440.png、home-390.png、public-snippet.png、login-redirect.png。

## 正式环境验收

CLI Playwright：1440桌面浅色、390手机深色首页200，编辑器和添加文件交互正常，无横向溢出；真实创建201，公开分享页读取内容成功；密码保护挑战423，正确密码解锁200；pageerror为空。

/admin正确跳转Account登录页，客户端与正式回调正确。初始化接口200，claimed=false、canClaim=true。未使用生产账号登录，完整登录回调、管理员认领及认证后后台操作未在生产执行；用户需首次登录后显式认领。

公开资料代理返回线上昵称、头像标识，Account真实头像图片交付200；/healthz正常。验收创建的两个匿名片段ehqeTZAG、A8ottU71已按精确代码、标题、匿名归属条件删除，未导入或保留本地样本。

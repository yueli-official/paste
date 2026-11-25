# 2026-11-23 代码分享九宫格部署

[代码分享](https://paste.yuelili.com/)已更新至 `site-launcher-20261123-1`。页头已采用共享九宫格。九个入口以浅色底和彩色图标呈现。站点 Logo 和 favicon 可独立配置，公开接口 `GET /api/v1/site/branding` 返回当前地址。

- 桌面和手机公网浏览器均通过九宫格验证，56 项脚本与样式的摘要与制品一致。两端运行镜像已核对，重启次数为 0。
- 执行了 `0006_authorization_catalog_v2` 和 `0008_site_branding`。恢复旧 Web 时须保留支持 v2 授权目录的 API。
- API 镜像 `sha256:48a3f36f68b3ac0d6012ccfd5a24cebbe988169c3a28f3dcb9c7419f8fa533a1`，Web 镜像 `sha256:56dee14a1c744811fefb873498f8d12c1ad4cebf5079c878b784f678b8df92f1`。
- 源码快照 `fd74e68132a740debf60b1113c85311a6602c84218c17cb5bf6dadc81e918b2a`。双端验收回执见 `/projects/yuelili.com/paste/deployment-release.json`，恢复配置见 `/projects/yuelili.com/backups/paste-site-launcher-20261123-1`。

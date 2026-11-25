# 2026-11-24 站群入口与联系方式部署

页头九宫格放在账号入口左侧，已有主题按钮时放在主题按钮右侧。共享菜单保留站点官方图标，在底部提供 QQ 群、GitHub 和哔哩哔哩入口。

| 服务 | 版本 | 运行镜像 |
| --- | --- | --- |
| Web | `launcher-contacts-20261124-1` | `sha256:457798907f154dd7c83eb642fff8f5724dbbfe0459f0554c071758ec9652f46b` |
| API | `site-launcher-20261123-1` | `sha256:48a3f36f68b3ac0d6012ccfd5a24cebbe988169c3a28f3dcb9c7419f8fa533a1` |

[正式站点](https://paste.yuelili.com/)的桌面和手机浏览器检查通过，70 项静态资源摘要与发布制品一致。共享 UI 使用 `0.10.4-preview.20261124.contacts.1`。API 复用原镜像，数据库无需迁移。

生效回执位于 `/projects/yuelili.com/paste/deployment-release.json`，制品与核验记录位于 `/projects/yuelili.com/.deploy/paste-launcher-contacts-20261124-1/`。恢复配置及旧回执位于 `/projects/yuelili.com/backups/paste-launcher-contacts-20261124-1/`，上一版 Web 镜像为 `yueli/paste-web:mobile-navigation-20261124-3`。

# 2026-11-24 代码分享移动导航部署

手机底栏提供「首页」「内容库」「我的」。[「我的」](https://paste.yuelili.com/me)是独立页面，站群九宫格留在右上角。手机打开编辑器时保留底栏，进入编辑后底栏为键盘让位。

| 服务 | 版本 | 运行镜像 ID |
| --- | --- | --- |
| Web | `mobile-navigation-20261124-2` | `sha256:01ef3ceda22741275cd784f4112a006cebf34d3e849e1019db54fb45858001d3` |
| API | `site-launcher-20261123-1` | `sha256:48a3f36f68b3ac0d6012ccfd5a24cebbe988169c3a28f3dcb9c7419f8fa533a1` |

候选通过手机与桌面浏览器检查。公网复查了九宫格、手机导航和「我的」页面，70 项脚本及样式摘要与制品一致。本次切换 `paste-web`，复用原 API，数据库无需迁移。

当前回执为 `/projects/yuelili.com/paste/deployment-release.json`，制品和验证证据在 `/projects/yuelili.com/.deploy/paste-mobile-navigation-20261124-2/`。恢复配置和前版回执位于 `/projects/yuelili.com/backups/paste-mobile-navigation-20261124-2/`，前版 Web 镜像为 `yueli/paste-web:site-launcher-20261123-1`。恢复时使用备份的 Web 镜像配置重建 `paste-web`，核对公网后恢复对应回执。

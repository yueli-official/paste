# 2026-11-24 代码片段手机导航调整

「内容库」保留在手机底栏。「我的」继续使用共享账户面板。

候选和公网通过手机与桌面浏览器复查，70 项脚本及样式与制品摘要一致。仅更新 `paste-web` 到 `mobile-navigation-20261124-3`，API 复用 `site-launcher-20261123-1`，数据库无需迁移。

当前回执为 `/projects/yuelili.com/paste/deployment-release.json`，制品和证据位于 `/projects/yuelili.com/.deploy/paste-mobile-navigation-20261124-3/`。恢复配置和前版回执位于 `/projects/yuelili.com/backups/paste-mobile-navigation-20261124-3/`。恢复时使用 `yueli/paste-web:mobile-navigation-20261124-2` 的备份配置重建 `paste-web`，核验后恢复对应回执。

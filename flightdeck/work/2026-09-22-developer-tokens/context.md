# 上下文与权限边界

用户要求按 BVideo、Paste（代码片段）、Gallery、Nav、Distribution 的方向让更多站点支持开发者令牌。仓库规则禁止建立笼统的多站推广 Work，因此本 Work 只交付 Paste；后续产品各自拥有独立 Goal、权限模型和验收。

Paste 是匿名优先的多文件代码片段产品。PAT 只为已登录用户提供可撤销的自动化身份，不改变匿名创建、密码访问、所有者约束、修订冲突和管理员授权模型。Identity 仍是凭证权威，Paste 仍是业务权限权威。

计划声明六项可委托能力：

- `paste.paste.create`：以当前用户身份创建片段。
- `paste.paste.read`：列出并读取当前用户自己的片段。
- `paste.paste.update`：编辑当前用户自己的片段。
- `paste.paste.delete`：删除当前用户自己的片段。
- `paste.paste.moderate`：管理员查询、治理和删除站内片段，仍需当前 Paste 管理员权限。
- `paste.settings.manage`：管理员修改公开站点设置，仍需当前 Paste 管理员权限。

公开读取无需 PAT。`/api/v1/authorization/setup/claim`、`/api/v1/admin/session`、`/api/v1/admin/users*`、`/api/v1/admin/governance-settings` 与其他未声明路由默认拒绝 PAT。权限目录只接受 `identity-svc` 且带 `personal-token:permissions` scope 的服务身份；目录展示与实际调用都重新读取当前 Paste 授权。

本地接入使用现有 `paste-yueli-web` OIDC client 作为 PAT application/site ID，资源 audience 继续是 `paste-api`。Workspace Environment 负责把 Identity catalog 与 Paste 在线验证端点连通，产品仓库不复制 Identity 的令牌存储或 Cookie 行为。

# Paste 开发者 API

> 自动生成，请勿手改；从运行时 PAT 声明和产品 OpenAPI 派生。

[使用指南](../developer-tokens.md) · [机器能力目录](developer-api.json) · [PAT OpenAPI](developer-openapi.json)

本文件描述当前源码允许的范围。生产部署：**未知**；真实调用验证：**未知**。实际生效情况查产品部署记录。

请求携带 `Authorization: Bearer <PAT>` 与 `X-Yueli-Token-Site: <实际站点 ID>`。
同一分支内 scope 全部需要，分支之间任选其一；仍须通过当前账号与资源权限检查。目录不是授权凭证。

## 可申请权限

| Scope | 名称 | 当前业务能力 | 说明 |
| --- | --- | --- | --- |
| paste.paste.create | 创建片段 | paste.paste.create | 以当前账号身份创建多文件代码片段，仍受账号创建状态与每日额度限制。 |
| paste.paste.delete | 删除自己的片段 | paste.paste.delete | 删除当前账号拥有的片段，继续遵守修订号并发保护。 |
| paste.paste.moderate | 治理站内片段 | paste.paste.moderate | 查询、修改和删除站内片段；仍需当前 Paste 管理员权限。 |
| paste.paste.read | 读取自己的片段 | paste.paste.read | 列出并读取当前账号拥有的片段，不包含其他用户或匿名片段的后台治理。 |
| paste.paste.update | 编辑自己的片段 | paste.paste.update | 编辑当前账号拥有的片段，继续遵守修订号并发保护。 |
| paste.settings.manage | 管理公开站点设置 | paste.settings.manage | 修改 Paste 名称与公开说明；仍需当前 Paste 管理员权限。 |

## 媒体 Profile

| Namespace / Profile | 所需 Scope |
| --- | --- |

## 允许调用的操作

| 方法 / 路由 | 操作 | 所需 Scope |
| --- | --- | --- |
| `GET /api/v1/admin/pastes` | List Pastes for site governance | `paste.paste.moderate` |
| `DELETE /api/v1/admin/pastes/{code}` | Delete a Paste as site administrator | `paste.paste.moderate` |
| `PATCH /api/v1/admin/pastes/{code}` | Govern a Paste | `paste.paste.moderate` |
| `PATCH /api/v1/admin/settings` | Update public site settings | `paste.settings.manage` |
| `GET /api/v1/me/pastes` | List the current user's Pastes | `paste.paste.read` |
| `DELETE /api/v1/me/pastes/{code}` | Delete an owned Paste | `paste.paste.delete` |
| `GET /api/v1/me/pastes/{code}` | Get an owned Paste for management | `paste.paste.read` |
| `PATCH /api/v1/me/pastes/{code}` | Update an owned Paste | `paste.paste.update` |
| `POST /api/v1/pastes` | Create an anonymous or user-owned Paste | `paste.paste.create` |
| `GET /api/v1/pastes/{code}` | Open a Paste that does not require a password | `paste.paste.read` |
| `POST /api/v1/pastes/{code}/access` | Open a password-protected Paste | `paste.paste.read` |
| `GET /api/v1/settings` | Read public site settings | `paste.paste.read` |
| `GET /api/v1/site/branding` | Get public site branding | `paste.paste.read` |
| `PATCH /api/v1/site/branding` | Update site branding | `paste.settings.manage` |

字段、请求体、响应与引用 Schema 见生成的 OpenAPI。未列入的接口不代表允许 PAT 调用。

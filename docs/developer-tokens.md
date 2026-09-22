# 开发者令牌 API

Paste 复用 Identity 的开发者令牌。账户中心实时查询当前账号在 Paste 的权限；脚本向 Paste API 发送 `Authorization: Bearer <PAT>`，不复制浏览器 Cookie。

## 权限

| 权限 | 能力 | 当前账号要求 |
| --- | --- | --- |
| `paste.paste.create` | 以当前账号身份创建多文件片段 | 有效用户；仍受创建状态和每日额度限制 |
| `paste.paste.read` | 列出、读取自己的片段；读取公开片段与站点设置 | 有效用户 |
| `paste.paste.update` | 编辑自己的片段 | 所有者；仍需正确 `expectedRevision` |
| `paste.paste.delete` | 删除自己的片段 | 所有者；仍需正确 `expectedRevision` |
| `paste.paste.moderate` | 查询、治理和删除站内片段 | 当前 Paste 管理员 |
| `paste.settings.manage` | 修改公开站点名称与说明 | 当前 Paste 管理员 |

实际权限始终是“令牌勾选 scope ∩ 当前 Paste 授权”。管理员撤权或 Token 撤销会在下一次请求生效。PAT 不开放首位管理员认领、管理员会话、用户额度/封禁、反滥用设置或其他未明确列出的路由；失败的显式 PAT 不会回退浏览器登录状态。

## 本地环境

从 Workspace 启动 Paste 隔离组合，避免改变其他站点正在使用的共享 Identity catalog：

```powershell
cd ..\workspace
.\environments\paste-local\run.ps1 -Mode Isolated
```

默认入口：

- Paste：`http://paste.dev.yuelili.test:3010`
- 账户中心：`http://account-paste.dev.yuelili.test:3000/developer-tokens`
- Paste API：`http://127.0.0.1:8093`

本地 Environment 登记 `paste-yueli-web` 为 PAT application，资源 audience 保持 `paste-api`。Paste API 使用 `PASTE_PERSONAL_TOKEN_SITE_ID`、`PASTE_PERSONAL_TOKEN_VERIFY_URL` 和 `PASTE_PERSONAL_TOKEN_ALLOW_HTTP` 接入 Identity 在线验证；非回环与生产环境默认要求 HTTPS。

## API 示例

完整字段与响应见 [OpenAPI](../contracts/openapi/paste.json)。以下示例假设调用者已把 Token 放入 `PASTE_TOKEN`，且不把它写进仓库或命令输出。

```powershell
$base = 'http://127.0.0.1:8093'
$headers = @{ Authorization = "Bearer $env:PASTE_TOKEN" }

$payload = @{
  title = '示例'
  visibility = 'private'
  files = @(@{ path = 'main.go'; language = 'go'; content = 'package main' })
} | ConvertTo-Json -Depth 8

$created = Invoke-RestMethod "$base/api/v1/pastes" -Method Post -Headers $headers `
  -ContentType 'application/json; charset=utf-8' -Body $payload

$mine = Invoke-RestMethod "$base/api/v1/me/pastes" -Headers $headers

$title = @{ title = '更新后的标题'; expectedRevision = $created.paste.revision } | ConvertTo-Json
$updated = Invoke-RestMethod "$base/api/v1/me/pastes/$($created.paste.code)" -Method Patch `
  -Headers $headers -ContentType 'application/json' -Body $title
```

创建成功返回 `201 + Paste DTO` 和 `Location`；删除成功返回 204。失败使用 HTTP Problem，以 HTTP 状态和稳定 `code` 判断。更新、治理和删除必须先读取当前 revision；遇到 409 时重新读取，不强行覆盖。

## 服务接入与发布状态

Identity 的 `pat.applications` 登记 `id=paste-yueli-web`、`audience=paste-api`，权限目录为 Paste 的 `/api/v1/internal/personal-token/permissions`。该目录只接受 `identity-svc` 且带 `personal-token:permissions` scope 的服务身份。Nuxt BFF 保留显式 PAT，不将其换成浏览器 Cookie。

2026-09-22 本功能处于本地源码交付与验收阶段；`https://paste.yuelili.com` 当前生产版本尚未因此更新。生产部署必须同时配置 Identity catalog 与 Paste 在线验证参数，并完成真实令牌闭环后，才能把本节改为已上线。

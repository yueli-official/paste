# 2026-10-17 共享组件默认样式部署

已部署 [paste](https://paste.yuelili.com)：`yueli/paste-web:server-20261017-component-styles-1`，实际镜像 `sha256:1d9a2b984be9fb0df1726ac139caceef02d2b504b3e00c1990af4bca5aba1796`。

用户确认WWW/Docs/Blog通过后授权继续其余已上线新版。本产品从实际线上源码归档固定快照，仅叠加批准组件和自动样式接入。保留已上线修复，排除其他尚未上线业务。未执行Git提交、推送或公开包发布。

- 依赖基线：`6ef835526d6e357e14a638594c1ae3fc03826d53fff61a05f5fd7c3973892199`。UI 0.9.0 styles.1、Content 0.3.0 styles.1、Asset 0.4.0 styles.2、Identity Nuxt 0.2.3 styles.1，包按产品实际闭包采用；均为固定候选制品。必需CSS由模块/layer提供，品牌和可选定制保留。
- 来源：`a31fdc69ebc286b936b738b673fd4e77f150bef02aaa22c6171ea43317e77135`；逐文件摘要、API/Web来源、制品、依赖回执和候选/正式验收见 `/projects/yuelili.com/.deploy/paste-component-styles-20261017-1/release.json`。
- 固定快照独立冻结安装、prepare、类型检查、所属单元测试及生产构建通过。没有通用test脚本的Distribution/Forum直接运行已有Node测试文件，不更改产品脚本或跳过检查。
- API复用 `yueli/paste-api:server-20260911-redeploy-1`：源码、原始制品/二进制、运行镜像及接口探测匹配，API容器保持；没有迁移、生产密钥轮换或共享Provider重启。
- 配置与数据库备份：`/projects/yuelili.com/backups/paste-component-styles-20261017-1`；数据库custom dump经pg_restore清单确认可读。持久配置保持，仅精确替换本产品Web。
- Linux候选和正式HTTPS的CLI Playwright通过390/1440、明暗主题、渲染/水合、无溢出/脚本异常/5xx、匿名登录门禁；关键JS及全部CSS的HTTP SHA256与制品一致。公开内容及可见图片采用真实线上数据。本轮没有生产管理员会话或真实资金交易，不宣称后台写入、支付或上传生产内容已验证；软件编辑真实上传/插入/再编辑有本轮本地真实组合与独立制品证据。

本地证据：`E:/codex-task-artifacts/component-styles-20261017/production-more-20261017/paste`。本任务只切换六个Web，均healthy、restart count 0；已通过的WWW/Docs/Blog、各站API在各自发布基准后的容器、Asset和数据库保持。期间另一会话更新Commerce API为account-center-20261017及外部Registry；其来源核验和原始/刷新基准保留，Commerce已采用该最新源码/运行证明重新构建Web。已确认的registry-account-center-candidate及commerce-account-center-candidate为他会话临时候选，不停止它，也不将其生命周期当成本任务修改。其余正式容器按复核后的基准保持。候选容器和服务器image-web目录回收，源码归档、制品、备份、截图、日志和回执保留。

本地依赖与Nuxt/构建输出缓存已清理，源码快照、原始/刷新基准、制品、回执、生成样式图、日志和截图保留；SSH隧道已停止。候选代理验收区分换页/关闭时取消的请求与真实失败，并保留原始失败记录；BVideo按实际桌面侧栏/手机弹层分别验收。

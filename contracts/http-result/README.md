# Paste HTTP Result

从产品根运行Foundation CLI：`go run github.com/yueli-official/foundation/go/httpcontract/cmd/httpcontract -project contracts/http-result/project.json`，附加`-check`拒绝生成漂移。Project编排使用显式本地Foundation；CI用已发布v0.4.1分别验证OpenAPI、目录生成与兼容性差异。

17个operation的错误集合由operation-errors.json拥有，错误定义由error-catalog.json拥有。OpenAPI producer不依赖数据库或登录服务。成功直接返回端点DTO，创建201并提供资源Location，无正文删除204。我的片段、后台片段/用户均使用items/page/size/total，page从1起、size默认50上限100；旧pastes/users与limit/offset不再是公开集合合同。单资源仍用端点自有paste/settings/user DTO字段，没有通用code/data/message Envelope。

我的片段的用户范围在数据库查询前确定，q匹配标题、标识、语言与标签。域层保存内部Limit/Offset不暴露到HTTP。每日业务额度使用paste.daily_limit_reached；传输频控使用common.rate_limited。

前端只用Foundation结构化failure，标题/说明/密码等字段接收inline错误，未映射violations进入摘要，traceId位于技术详情。原始message/detail不对用户展示，typed cause保留errors.Is/As。不会自动重试创建。

Go依赖升级至已发布v0.4.1，Web使用js-v0.7.2/Identity v0.3.3正式URL；当前组合仍使用明确的Workspace源码overlay。生产部署正式制品的组合验证与本地门禁分别记录，不推断发布授权。

import { resolveFailureFeedback } from "@yueli/http-runtime";
import {
  pasteFailurePresentation,
  type PasteFailureCode,
} from "../generated/pasteFailure";

const messages: Record<PasteFailureCode, string> = {
  "paste.not_authenticated": "请先登录，再继续管理你的代码片段。",
  "paste.forbidden": "当前账号没有执行此操作的权限。",
  "paste.creation_suspended":
    "你的代码片段创建权限已被暂停；已有内容仍可查看和删除。",
  "paste.anonymous_creation_disabled": "本站暂时停止匿名创建，请登录后重试。",
  "paste.not_found": "此代码片段不存在，请检查分享链接。",
  "paste.gone": "此代码片段已过期或已被删除。",
  "paste.password_required": "此代码片段需要密码才能打开。",
  "paste.password_invalid": "密码不正确，请重新输入。",
  "paste.conflict": "内容已在别处更新，请刷新后重试。",
  "paste.daily_limit_reached":
    "今天的代码片段创建额度已经用完，请在下一个 UTC 自然日再试。",
};

export function pasteFailureFeedback(
  reason: unknown,
  fallback: string,
  fields?: Readonly<Record<string, string>>,
) {
  return resolveFailureFeedback(reason, {
    fallback,
    fields,
    resolveText(code) {
      if (Object.hasOwn(pasteFailurePresentation, code))
        return { message: messages[code as PasteFailureCode] };
      const common: Record<string, string> = {
        "common.validation_failed": "填写的信息不符合要求，请检查后重试。",
        "common.rate_limited": "操作过于频繁，请稍后重试。",
        "foundation.request.network": "网络连接失败，请检查网络后重试。",
        "foundation.request.timeout": "请求超时，请稍后重试。",
        "validation.required": "请填写此项。",
        "validation.max_bytes":
          "单个文件不能超过 1 MiB，请拆分文件或删减内容后重试。",
        "validation.max_total_bytes":
          "全部文件合计不能超过 1 MiB，请删减内容后重试。",
        "validation.max_length": "内容超过长度限制，请缩短后重试。",
        "validation.max_items": "项目数量超过限制，请减少后重试。",
        "validation.range": "数值超出允许范围，请检查后重试。",
        "validation.invalid": "此项内容不符合要求，请检查后重试。",
      };
      return common[code] ? { message: common[code] } : undefined;
    },
  });
}

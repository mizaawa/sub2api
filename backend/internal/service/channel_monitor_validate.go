package service

import (
	"net/url"
	"strings"
)

// 渠道监控参数校验与归一化辅助函数。
// 校验失败一律返回 channel_monitor_const.go 中预定义的 Err* 错误，错误信息不含具体 IP/hostname，避免泄露内网拓扑。

// validateProvider 校验 provider 字符串。
// 唯一来源于 providerAdapters：新增 provider 只需要在 channel_monitor_checker.go 注册 adapter。
func validateProvider(p string) error {
	if !isSupportedProvider(p) {
		return ErrChannelMonitorInvalidProvider
	}
	return nil
}

// validateAPIMode 校验 provider 与 api_mode 的组合。
// responses 只对 OpenAI 有意义；其它 provider 使用 chat_completions 作为默认占位。
func validateAPIMode(provider, apiMode string) error {
	apiMode = defaultAPIMode(apiMode)
	switch apiMode {
	case MonitorAPIModeChatCompletions:
		return nil
	case MonitorAPIModeResponses:
		if provider == "" || provider == MonitorProviderOpenAI {
			return nil
		}
		return ErrChannelMonitorInvalidAPIMode
	default:
		return ErrChannelMonitorInvalidAPIMode
	}
}

// validateInterval 校验 interval_seconds 范围。
func validateInterval(sec int) error {
	if sec < monitorMinIntervalSeconds || sec > monitorMaxIntervalSeconds {
		return ErrChannelMonitorInvalidInterval
	}
	return nil
}

// validateJitter 校验 jitter_seconds（调度 ± 随机抖动）：
// 非负，且 interval - jitter 不得低于最小检测间隔，防止随机偏移后实际间隔过短打爆上游。
func validateJitter(jitterSec, intervalSec int) error {
	if jitterSec < 0 || intervalSec-jitterSec < monitorMinIntervalSeconds {
		return ErrChannelMonitorInvalidJitter
	}
	return nil
}

// validateEndpoint 校验 endpoint：
//   - scheme 支持 https 和 http（允许在本地环境或测试环境使用 HTTP）
//   - 必须为 origin（无 path/query/fragment），防止填写 https://api.openai.com/v1
//     导致 joinURL 拼出 /v1/v1/chat/completions
//
// 注意：SSRF 防护已移除，因为渠道监控是管理员专用功能（非用户提交）。
// 管理员可以配置任意地址，包括：
//   - 公网 IP/域名（如 https://1.2.3.4 或 https://api.example.com）
//   - 内网 IP（如 http://192.168.1.100 或 http://10.0.0.5）
//   - 本地地址（如 http://localhost:8080 或 http://127.0.0.1）
func validateEndpoint(ep string) error {
	ep = strings.TrimSpace(ep)
	if ep == "" {
		return ErrChannelMonitorInvalidEndpoint
	}
	u, err := url.Parse(ep)
	if err != nil {
		return ErrChannelMonitorInvalidEndpoint
	}
	if u.Scheme != "https" && u.Scheme != "http" {
		return ErrChannelMonitorEndpointScheme
	}
	if u.Host == "" {
		return ErrChannelMonitorInvalidEndpoint
	}
	if u.Path != "" && u.Path != "/" {
		return ErrChannelMonitorEndpointPath
	}
	if u.RawQuery != "" || u.Fragment != "" {
		return ErrChannelMonitorEndpointPath
	}

	// SSRF 防护已禁用：管理员功能，信任所有配置
	// 如需恢复防护，参考 git 历史记录中的 SSRF 检查代码
	return nil
}

// normalizeEndpoint 去除前后空白与末尾 `/`，保证存储统一为 origin。
// validateEndpoint 已确保格式合法（仅 origin），这里只做最终归一化。
func normalizeEndpoint(ep string) string {
	ep = strings.TrimSpace(ep)
	ep = strings.TrimRight(ep, "/")
	return ep
}

// normalizeModels 去除空白、重复模型名。保留输入顺序（map 的迭代顺序无关）。
func normalizeModels(in []string) []string {
	if len(in) == 0 {
		return []string{}
	}
	seen := make(map[string]struct{}, len(in))
	out := make([]string, 0, len(in))
	for _, m := range in {
		m = strings.TrimSpace(m)
		if m == "" {
			continue
		}
		if _, ok := seen[m]; ok {
			continue
		}
		seen[m] = struct{}{}
		out = append(out, m)
	}
	return out
}

// normalizeMonitorPrimaryModel applies the Grok health-check default while
// preserving the existing required-model behavior for every other provider.
func normalizeMonitorPrimaryModel(provider, model string) string {
	model = strings.TrimSpace(model)
	if model == "" && provider == MonitorProviderGrok {
		return MonitorDefaultGrokModel
	}
	return model
}

// defaultAPIMode 空串归一为 chat_completions，保证历史数据与旧客户端兼容。
func defaultAPIMode(apiMode string) string {
	if strings.TrimSpace(apiMode) == "" {
		return MonitorAPIModeChatCompletions
	}
	return strings.TrimSpace(apiMode)
}

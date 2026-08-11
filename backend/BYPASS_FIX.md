# Bypass 功能修复说明

## 问题描述

关闭 bypass 后，Claude 模型映射无法被下游检测到：
- ✅ GPT 模型映射正常（如 gpt-5.6 → gpt-5.5）
- ❌ Claude 模型映射失效（如 claude-opus-5 → claude-fable-5）

## 根本原因

`openai_gateway_messages.go` 缺少 bypass 逻辑：
- 其他文件（`gateway_anthropic_passthrough.go`, `openai_gateway_passthrough.go` 等）都有 bypass 支持
- 但 `openai_gateway_messages.go` 始终返回 `originalModel`，忽略了 bypass 开关状态

## 修复内容

在 `internal/service/openai_gateway_messages.go` 中添加了 bypass 逻辑：

### 1. 非流式响应 (handleAnthropicBufferedStreamingResponse)
```go
// 根据 bypass 设置决定响应中显示的模型
responseModel := originalModel
bypassModelConsistency := downstreamModelConsistencyBypassEnabled(c.Request.Context(), s.settingService)
if !bypassModelConsistency {
    // bypass 关闭时，暴露真实的上游模型
    if observedModel := observedUpstreamResponseModel(c); strings.TrimSpace(observedModel) != "" {
        responseModel = observedModel
    }
}
anthropicResp := apicompat.ResponsesToAnthropic(finalResponse, responseModel)
```

### 2. 流式响应初始化 (handleAnthropicStreamingResponse)
```go
// 初始化 state.Model，bypass 关闭时使用观察到的上游模型
responseModel := originalModel
bypassModelConsistency := downstreamModelConsistencyBypassEnabled(c.Request.Context(), s.settingService)
if !bypassModelConsistency {
    if observedModel := observedUpstreamResponseModel(c); strings.TrimSpace(observedModel) != "" {
        responseModel = observedModel
    }
}
state.Model = responseModel
```

### 3. 流式响应动态更新 (processDataLine)
```go
// 每次观察到上游事件后，动态更新 state.Model
if !bypassModelConsistency {
    if observedModel := observedUpstreamResponseModel(c); strings.TrimSpace(observedModel) != "" {
        state.Model = observedModel
    }
}
```

### 4. 结果模型字段 (resultWithUsage)
```go
// 最终结果中也使用正确的模型
finalModel := originalModel
if !bypassModelConsistency {
    if observedModel := observedUpstreamResponseModel(c); strings.TrimSpace(observedModel) != "" {
        finalModel = observedModel
    }
}
out.Model = finalModel
```

## Bypass 逻辑说明

- **bypass 开启** (默认行为) = 隐藏映射，下游看到请求的模型
  - 请求 `claude-fable-5` → 响应显示 `claude-fable-5`
  - 实际调用 `claude-opus-5`（对下游透明）

- **bypass 关闭** (新增行为) = 暴露映射，下游看到真实上游模型
  - 请求 `claude-fable-5` → 响应显示 `claude-opus-5`
  - 下游可以检测到"模型不一致"

## 测试验证

测试场景：
1. **bypass 开启** + opus-5 映射 fable-5
   - 预期：响应显示 `claude-fable-5`
   
2. **bypass 关闭** + opus-5 映射 fable-5
   - 预期：响应显示 `claude-opus-5`
   - 下游检测到模型不一致

## 修改文件

- `internal/service/openai_gateway_messages.go` (+40 行)

## 相关函数

- `downstreamModelConsistencyBypassEnabled()` - 检查 bypass 开关状态
- `observedUpstreamResponseModel()` - 获取观察到的上游模型
- `observedUpstreamResponseModelConflict()` - 检查模型冲突

## 兼容性

- ✅ 向后兼容：默认行为（bypass 开启）不变
- ✅ 与其他文件逻辑一致
- ✅ 不影响计费模型 (`billingModel`)
- ✅ 保留完整的模型追踪信息

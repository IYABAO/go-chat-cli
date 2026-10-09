# go-chat-cli

最小化的 OpenAI 兼容 API 命令行聊天工具。用 Go 标准库实现，零第三方依赖。

## 特性

- 🚀 单文件，零依赖，`go run .` 即跑
- 💬 多轮对话历史（内存）
- ⚡️ 流式输出（SSE）
- 🔌 兼容所有 OpenAI 协议的端点（OpenAI / 豆包 / DeepSeek / 通义 / Ollama）
- 🛠 支持 `/reset` 清空对话、`/exit` 退出

## 快速开始

```bash
# 1. 设置 API Key（以豆包为例）
export OPENAI_API_KEY="你的key"
export OPENAI_BASE_URL="https://ark.cn-beijing.volces.com/api/v3"
export OPENAI_MODEL="doubao-pro-4k"

# 2. 运行
go run .
```

## 命令行参数

| 参数 | 环境变量 | 默认值 | 说明 |
|---|---|---|---|
| `--key` | `OPENAI_API_KEY` | （必填） | API Key |
| `--base-url` | `OPENAI_BASE_URL` | `https://api.openai.com/v1` | API 端点 |
| `--model` | `OPENAI_MODEL` | `doubao-pro-4k` | 模型名称 |

## 构建

```bash
go build -o go-chat-cli .
./go-chat-cli
```

## 支持的服务商

只要兼容 `/v1/chat/completions` 接口都能用：

| 服务商 | base-url | 示例 model |
|---|---|---|
| OpenAI | `https://api.openai.com/v1` | `gpt-4o-mini` |
| 豆包（火山方舟） | `https://ark.cn-beijing.volces.com/api/v3` | `doubao-pro-4k` |
| DeepSeek | `https://api.deepseek.com/v1` | `deepseek-chat` |
| 通义千问 | `https://dashscope.aliyuncs.com/compatible-mode/v1` | `qwen-turbo` |
| Ollama（本地） | `http://localhost:11434/v1` | `llama3` |

## License

MIT

# 用 Go 标准库写一个大模型命令行聊天工具，零第三方依赖

## 为什么不用现成的？

市面上聊天客户端很多，但自己写一个的好处：
1. **理解底层**：OpenAI 兼容协议到底怎么工作的，写一遍就懂
2. **零依赖**：不引入任何 SDK，出问题只能看 RFC 和源码
3. **简历项目**：一个能跑、能编译、有 README 的 Go 项目

## 核心：OpenAI 兼容协议

不管是 OpenAI、豆包、DeepSeek、通义还是本地 Ollama，都是同一套接口：

```
POST /v1/chat/completions
Authorization: Bearer sk-xxx
Content-Type: application/json

{
  "model": "deepseek-chat",
  "messages": [{"role":"user","content":"你好"}],
  "stream": true
}
```

流式返回是 SSE（Server-Sent Events），格式：

```
data: {"choices":[{"delta":{"content":"你"}}]}
data: {"choices":[{"delta":{"content":"好"}}]}
data: [DONE]
```

## 代码实现

整个核心不到 150 行，全部用标准库：

### 1. 数据结构

```go
type Message struct {
    Role    string `json:"role"`
    Content string `json:"content"`
}
```

### 2. 发起请求

```go
func streamChat(client *http.Client, baseURL, apiKey, model string, history []Message) (string, error) {
    url := strings.TrimRight(baseURL, "/") + "/chat/completions"
    body := chatRequest{Model: model, Messages: history, Stream: true}
    buf, _ := json.Marshal(body)

    req, _ := http.NewRequest("POST", url, bytes.NewReader(buf))
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("Authorization", "Bearer "+apiKey)

    resp, err := client.Do(req)
    // ...
}
```

### 3. 解析 SSE 流

这是关键：不要等全部返回再打印，而是逐行读 `data: ` 前缀：

```go
scanner := bufio.NewScanner(resp.Body)
for scanner.Scan() {
    line := scanner.Text()
    if !strings.HasPrefix(line, "data: ") { continue }
    data := strings.TrimPrefix(line, "data: ")
    if data == "[DONE]" { break }

    var chunk chatResponse
    json.Unmarshal([]byte(data), &chunk)
    if len(chunk.Choices) > 0 {
        delta := chunk.Choices[0].Delta.Content
        fmt.Print(delta)  // 实时打印
        full.WriteString(delta)
    }
}
```

### 4. 多轮对话

就是维护一个 messages 数组，每轮把 user 和 assistant 的回复都 append 进去：

```go
history = append(history, Message{Role: "user", Content: input})
// 调 API...
history = append(history, Message{Role: "assistant", Content: full})
```

## 踩坑记录

### 坑 1：Windows 装 Go

winget 装不上（网络问题），msi 静默安装也失败。最后用 zip 版解压即用：

```powershell
# 下载 zip 版，解压到 C:\Users\xxx\go-sdk\
# 然后把 C:\Users\xxx\go-sdk\bin 加到 PATH
go version
```

### 坑 2：git push 认证

Windows 上 git push 报 `could not read Username`，是因为没配 credential helper：

```powershell
git config --global credential.helper manager
```

之后 push 会自动弹浏览器授权一次，以后就不用再登了。

### 坑 3：不同服务商的 base_url 不一样

| 服务商 | base_url |
|---|---|
| OpenAI | `https://api.openai.com/v1` |
| DeepSeek | `https://api.deepseek.com/v1` |
| 豆包 | `https://ark.cn-beijing.volces.com/api/v3` |
| Ollama | `http://localhost:11434/v1` |

代码里做成参数，一行切换：

```powershell
go run . --base-url https://api.deepseek.com/v1 --model deepseek-chat
```

## 项目地址

GitHub: https://github.com/IYABAO/go-chat-cli

零第三方依赖，纯标准库，`go run .` 即跑。支持多轮对话、流式输出、`/reset` 清空、`/exit` 退出。

---

*我是 IYABAO，10 年全栈工程师，关注 AI 工程化与云原生后端。欢迎交流。*

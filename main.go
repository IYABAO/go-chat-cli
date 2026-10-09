// go-chat-cli: 一个最小的 OpenAI 兼容 API 命令行聊天工具
// 用法:
//   export OPENAI_API_KEY="sk-xxx"
//   export OPENAI_BASE_URL="https://ark.cn-beijing.volces.com/api/v3"  # 可选, 默认 OpenAI
//   go run .
package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

// Message 对应 OpenAI chat completions 的单条消息
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
	Stream   bool      `json:"stream"`
}

type chatResponse struct {
	Choices []struct {
		Delta struct {
			Content string `json:"content"`
		} `json:"delta"`
		Message Message `json:"message"`
	} `json:"choices"`
}

func main() {
	model := flag.String("model", getenv("OPENAI_MODEL", "doubao-pro-4k"), "模型名称")
	baseURL := flag.String("base-url", getenv("OPENAI_BASE_URL", "https://api.openai.com/v1"), "API base URL")
	apiKey := flag.String("key", getenv("OPENAI_API_KEY", ""), "API Key (也可用 OPENAI_API_KEY 环境变量)")
	flag.Parse()

	if *apiKey == "" {
		fmt.Fprintln(os.Stderr, "错误: 请设置 OPENAI_API_KEY 或用 --key 传入")
		os.Exit(1)
	}

	client := &http.Client{}
	history := []Message{{Role: "system", Content: "你是一个有用的助手。"}}

	fmt.Printf("go-chat-cli  (模型: %s, 端点: %s)\n", *model, *baseURL)
	fmt.Println("输入 /reset 清空对话, /exit 退出")
	fmt.Println("---")

	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("你> ")
		if !scanner.Scan() {
			break
		}
		input := strings.TrimSpace(scanner.Text())
		if input == "" {
			continue
		}
		if input == "/exit" || input == "/quit" {
			fmt.Println("再见!")
			return
		}
		if input == "/reset" {
			history = []Message{{Role: "system", Content: "你是一个有用的助手。"}}
			fmt.Println("[对话已清空]")
			continue
		}

		history = append(history, Message{Role: "user", Content: input})

		fmt.Print("AI> ")
		full, err := streamChat(client, *baseURL, *apiKey, *model, history)
		if err != nil {
			fmt.Fprintf(os.Stderr, "\n请求失败: %v\n", err)
			// 失败时回退刚加的 user 消息,避免历史污染
			history = history[:len(history)-1]
			continue
		}
		fmt.Println()
		history = append(history, Message{Role: "assistant", Content: full})
	}
}

// streamChat 发起流式请求,把 delta 打印到 stdout,返回完整回复
func streamChat(client *http.Client, baseURL, apiKey, model string, history []Message) (string, error) {
	url := strings.TrimRight(baseURL, "/") + "/chat/completions"
	body := chatRequest{Model: model, Messages: history, Stream: true}
	buf, _ := json.Marshal(body)

	req, err := http.NewRequest("POST", url, bytes.NewReader(buf))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(b))
	}

	var full strings.Builder
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		data := strings.TrimPrefix(line, "data: ")
		if data == "[DONE]" {
			break
		}
		var chunk chatResponse
		if err := json.Unmarshal([]byte(data), &chunk); err != nil {
			continue
		}
		if len(chunk.Choices) > 0 {
			delta := chunk.Choices[0].Delta.Content
			if delta != "" {
				fmt.Print(delta)
				full.WriteString(delta)
			}
		}
	}
	return full.String(), scanner.Err()
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

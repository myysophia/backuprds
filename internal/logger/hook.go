package logger

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"go.uber.org/zap/zapcore"
)

type WecomHook struct {
	Levels     []string
	WebhookURL string
}

func NewWecomHook(levels []string, webhookURL string) *WecomHook {
	return &WecomHook{
		Levels:     levels,
		WebhookURL: webhookURL,
	}
}

func (h *WecomHook) Fire(entry zapcore.Entry) error {
	// 检查是否需要处理该级别的日志
	levelStr := strings.ToLower(entry.Level.String())
	shouldProcess := false
	for _, l := range h.Levels {
		if levelStr == strings.ToLower(l) {
			shouldProcess = true
			break
		}
	}

	if !shouldProcess {
		return nil
	}

	// 构造消息
	timeStr := entry.Time.Format("2006-01-02 15:04:05")
	message := fmt.Sprintf(`【%s告警】
时间：%s
文件：%s
函数：%s
行号：%d
消息：%s`,
		strings.ToUpper(levelStr),
		timeStr,
		entry.Caller.File,
		entry.Caller.Function,
		entry.Caller.Line,
		entry.Message)

	// 如果有错误堆栈，添加到消息中
	if entry.Stack != "" {
		message += fmt.Sprintf("\n堆栈信息：\n%s", entry.Stack)
	}

	// 构造企业微信消息
	wecomMsg := map[string]interface{}{
		"msgtype": "markdown", // 使用 markdown 格式以获得更好的显示效果
		"markdown": map[string]string{
			"content": message,
		},
	}

	// 转换为JSON
	jsonData, err := json.Marshal(wecomMsg)
	if err != nil {
		return fmt.Errorf("failed to marshal wecom message: %v", err)
	}

	// 发送请求
	resp, err := http.Post(h.WebhookURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to send wecom message: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("wecom API returned non-200 status code: %d", resp.StatusCode)
	}

	return nil
}

package extractor

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	"cs_extractor/model"
)

// LLMClient 定义 LLM 客户端接口
type LLMClient interface {
	Extract(ctx context.Context, conv model.Conversation) (*model.ExtractionResult, error)
}

// DoubaoConfig 豆包 API 配置
type DoubaoConfig struct {
	APIKey     string
	BaseURL    string
	Model      string
	Temperature float64
	TopP       float64
	MaxTokens  int
	MaxRetries int
}

// DoubaoClient 豆包 API 客户端
type DoubaoClient struct {
	config     DoubaoConfig
	httpClient *http.Client
}

// NewDoubaoClient 创建豆包客户端
func NewDoubaoClient(config DoubaoConfig) *DoubaoClient {
	return &DoubaoClient{
		config: config,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// ChatRequest 聊天请求结构
type ChatRequest struct {
	Model          string          `json:"model"`
	Messages       []ChatMessage   `json:"messages"`
	Temperature    float64         `json:"temperature,omitempty"`
	TopP           float64         `json:"top_p,omitempty"`
	MaxTokens      int             `json:"max_tokens,omitempty"`
	ResponseFormat *ResponseFormat `json:"response_format,omitempty"`
	Stream         bool            `json:"stream,omitempty"`
}

// ChatMessage 聊天消息
type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ResponseFormat 响应格式
type ResponseFormat struct {
	Type       string                 `json:"type"`
	JSONSchema map[string]interface{} `json:"json_schema,omitempty"`
}

// ChatResponse 聊天响应结构
type ChatResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index   int `json:"index"`
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

// Extract 提取对话信息
func (c *DoubaoClient) Extract(ctx context.Context, conv model.Conversation) (*model.ExtractionResult, error) {
	userPrompt := BuildUserPrompt(conv)

	req := ChatRequest{
		Model: c.config.Model,
		Messages: []ChatMessage{
			{Role: "system", Content: SystemPrompt},
			{Role: "user", Content: userPrompt},
		},
		Temperature: c.config.Temperature,
		TopP:        c.config.TopP,
		MaxTokens:   c.config.MaxTokens,
		Stream:      false,
	}

	// 尝试使用 json_schema 格式
	req.ResponseFormat = c.buildJSONSchemaFormat()

	var lastErr error
	for attempt := 0; attempt <= c.config.MaxRetries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(time.Duration(attempt*attempt) * time.Second):
				// 指数退避
			}
		}

		result, tokens, err := c.doRequest(ctx, req)
		if err != nil {
			lastErr = err
			// 如果是 400 错误，尝试降级到 json_object
			if strings.Contains(err.Error(), "400") {
				req.ResponseFormat = &ResponseFormat{Type: "json_object"}
				result, tokens, err = c.doRequest(ctx, req)
				if err != nil {
					lastErr = err
					continue
				}
			} else {
				continue
			}
		}

		result.ConversationID = conv.ID
		result.Channel = conv.Channel
		result.AgentName = conv.Agent
		result.TurnCount = len(conv.Turns)
		result.ExtractedAt = time.Now()

		// 记录 token 使用（可以扩展）
		_ = tokens

		return result, nil
	}

	return nil, fmt.Errorf("重试 %d 次后仍然失败: %w", c.config.MaxRetries, lastErr)
}

// buildJSONSchemaFormat 构建 JSON Schema 格式
func (c *DoubaoClient) buildJSONSchemaFormat() *ResponseFormat {
	return &ResponseFormat{
		Type: "json_schema",
		JSONSchema: map[string]interface{}{
			"name":        "cs_extraction_result",
			"description": "客服对话结构化信息提取结果",
			"strict":      true,
			"schema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"conversation_id":    map[string]interface{}{"type": "string"},
					"channel":            map[string]interface{}{"type": "string"},
					"agent_name":         map[string]interface{}{"type": "string"},
					"user_query":         map[string]interface{}{"type": "string"},
					"main_intent":        map[string]interface{}{"type": "string", "enum": []string{"refund", "inquiry", "complaint", "technical", "account", "billing", "other"}},
					"issue_category":     map[string]interface{}{"type": "string", "enum": []string{"product_quality", "logistics", "account_mgmt", "technical_issue", "billing_payment", "complaint_suggest", "general_inquiry"}},
					"sub_issues":         map[string]interface{}{"type": "array", "items": map[string]interface{}{"type": "string"}},
					"is_resolved":        map[string]interface{}{"type": "boolean"},
					"resolution_type":    map[string]interface{}{"type": "string", "enum": []string{"self_service", "agent_resolved", "escalated", "transferred_human", "pending", "follow_up_required"}},
					"resolution_text":    map[string]interface{}{"type": "string"},
					"requires_follow_up": map[string]interface{}{"type": "boolean"},
					"user_sentiment":     map[string]interface{}{"type": "string", "enum": []string{"positive", "neutral", "negative"}},
					"sentiment_evidence": map[string]interface{}{"type": "string"},
					"turn_count":         map[string]interface{}{"type": "integer"},
					"topic_switches":     map[string]interface{}{"type": "integer"},
					"contains_transfer":  map[string]interface{}{"type": "boolean"},
					"has_info_missing":   map[string]interface{}{"type": "boolean"},
					"extracted_at":        map[string]interface{}{"type": "string"},
					"confidence":         map[string]interface{}{"type": "number"},
				},
				"required": []string{
					"conversation_id", "channel", "agent_name", "user_query", "main_intent",
					"issue_category", "is_resolved", "resolution_type", "requires_follow_up",
					"user_sentiment", "turn_count", "topic_switches", "contains_transfer",
					"has_info_missing", "extracted_at", "confidence",
				},
			},
		},
	}
}

// doRequest 执行 HTTP 请求
func (c *DoubaoClient) doRequest(ctx context.Context, req ChatRequest) (*model.ExtractionResult, int, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, 0, fmt.Errorf("序列化请求失败: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST",
		c.config.BaseURL+"/chat/completions",
		bytes.NewReader(body))
	if err != nil {
		return nil, 0, fmt.Errorf("创建请求失败: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.config.APIKey)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, 0, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, 0, fmt.Errorf("读取响应失败: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, 0, fmt.Errorf("API 错误 (状态码 %d): %s", resp.StatusCode, string(respBody))
	}

	var chatResp ChatResponse
	if err := json.Unmarshal(respBody, &chatResp); err != nil {
		return nil, 0, fmt.Errorf("解析响应失败: %w", err)
	}

	if len(chatResp.Choices) == 0 {
		return nil, 0, fmt.Errorf("响应中没有选项")
	}

	content := chatResp.Choices[0].Message.Content
	content = cleanJSONResponse(content)

	var result model.ExtractionResult
	if err := json.Unmarshal([]byte(content), &result); err != nil {
		return nil, 0, fmt.Errorf("解析提取结果失败: %w, 原始内容: %s", err, content)
	}

	return &result, chatResp.Usage.TotalTokens, nil
}

// cleanJSONResponse 清理 JSON 响应
func cleanJSONResponse(content string) string {
	// 去除 markdown 代码块标记
	re := regexp.MustCompile("```json\\s*")
	content = re.ReplaceAllString(content, "")
	re = regexp.MustCompile("```\\s*")
	content = re.ReplaceAllString(content, "")

	// 去除首尾空白
	content = strings.TrimSpace(content)

	// 如果有多余内容，截取最后一个完整的 JSON 对象
	lastBrace := strings.LastIndex(content, "}")
	if lastBrace > 0 {
		firstBrace := strings.Index(content, "{")
		if firstBrace >= 0 && firstBrace < lastBrace {
			content = content[firstBrace : lastBrace+1]
		}
	}

	return content
}
package extractor

import (
	"context"
	"time"

	"cs_extractor/model"
)

// MockClient Mock 客户端，用于测试
type MockClient struct {
	delay time.Duration
}

// NewMockClient 创建 Mock 客户端
func NewMockClient() *MockClient {
	return &MockClient{
		delay: 200 * time.Millisecond,
	}
}

// Extract 模拟提取对话信息
func (c *MockClient) Extract(ctx context.Context, conv model.Conversation) (*model.ExtractionResult, error) {
	// 模拟延迟
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-time.After(c.delay):
	}

	// 根据 ID 生成不同的模拟结果
	result := &model.ExtractionResult{
		ConversationID: conv.ID,
		Channel:        conv.Channel,
		AgentName:      conv.Agent,
		TurnCount:      len(conv.Turns),
		ExtractedAt:    time.Now(),
		Confidence:     0.95,
	}

	// 根据对话内容简单判断
	content := ""
	for _, turn := range conv.Turns {
		content += turn.Content
	}

	// 简单的关键词匹配
	if containsAny(content, []string{"退款", "退货", "换货"}) {
		result.UserQuery = "用户申请退款"
		result.MainIntent = model.MainIntentRefund
		result.IssueCategory = model.IssueCategoryProductQuality
	} else if containsAny(content, []string{"登录", "密码", "账号"}) {
		result.UserQuery = "账号登录问题"
		result.MainIntent = model.MainIntentAccount
		result.IssueCategory = model.IssueCategoryAccountMgmt
	} else if containsAny(content, []string{"卡", "慢", "加载", "闪退"}) {
		result.UserQuery = "APP使用问题"
		result.MainIntent = model.MainIntentTechnical
		result.IssueCategory = model.IssueCategoryTechnicalIssue
	} else if containsAny(content, []string{"投诉", "不满", "差评"}) {
		result.UserQuery = "用户投诉"
		result.MainIntent = model.MainIntentComplaint
		result.IssueCategory = model.IssueCategoryComplaintSuggest
	} else if containsAny(content, []string{"会员", "咨询", "了解"}) {
		result.UserQuery = "咨询会员政策"
		result.MainIntent = model.MainIntentInquiry
		result.IssueCategory = model.IssueCategoryGeneralInquiry
	} else {
		result.UserQuery = "一般咨询"
		result.MainIntent = model.MainIntentOther
		result.IssueCategory = model.IssueCategoryGeneralInquiry
	}

	// 判断解决情况
	if containsAny(content, []string{"谢谢", "好的", "可以", "解决"}) {
		result.IsResolved = true
		result.ResolutionType = model.ResolutionTypeAgentResolved
		result.RequiresFollowUp = false
	} else {
		result.IsResolved = false
		result.ResolutionType = model.ResolutionTypePending
		result.RequiresFollowUp = true
	}

	// 判断情绪
	if containsAny(content, []string{"谢谢", "好的", "满意", "太好了"}) {
		result.UserSentiment = model.UserSentimentPositive
	} else if containsAny(content, []string{"投诉", "不满", "差评", "太差", "垃圾"}) {
		result.UserSentiment = model.UserSentimentNegative
	} else {
		result.UserSentiment = model.UserSentimentNeutral
	}

	// 其他字段
	result.TopicSwitches = 1
	result.ContainsTransfer = false
	result.HasInfoMissing = false

	return result, nil
}

// containsAny 检查字符串是否包含任意关键词
func containsAny(s string, keywords []string) bool {
	for _, keyword := range keywords {
		if len(keyword) > 0 && len(s) >= len(keyword) {
			for i := 0; i <= len(s)-len(keyword); i++ {
				if s[i:i+len(keyword)] == keyword {
					return true
				}
			}
		}
	}
	return false
}
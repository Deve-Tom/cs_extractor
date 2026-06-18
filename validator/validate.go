package validator

import (
	"fmt"
	"strings"

	"cs_extractor/model"
)

// Validate 校验并修正提取结果
func Validate(result *model.ExtractionResult, conv model.Conversation) {
	// 1. 校验 conversation_id
	if result.ConversationID == "" {
		result.ConversationID = conv.ID
	}

	// 2. 校验 agent_name
	if result.AgentName == "" {
		result.AgentName = conv.Agent
	}

	// 3. 校验 channel
	if result.Channel == "" {
		result.Channel = conv.Channel
	}

	// 4. 校验 user_query（不超过50字）
	if result.UserQuery == "" {
		result.UserQuery = "未知诉求"
	}
	if len(result.UserQuery) > 50 {
		result.UserQuery = result.UserQuery[:50]
	}

	// 5. 校验 main_intent
	if !model.IsValidMainIntent(result.MainIntent) {
		fmt.Printf("⚠️ 警告: conv_%s 非法 main_intent '%s', 已修正为 'other'\n",
			result.ConversationID, result.MainIntent)
		result.MainIntent = model.MainIntentOther
	}

	// 6. 校验 issue_category
	if !model.IsValidIssueCategory(result.IssueCategory) {
		fmt.Printf("⚠️ 警告: conv_%s 非法 issue_category '%s', 已修正为 'general_inquiry'\n",
			result.ConversationID, result.IssueCategory)
		result.IssueCategory = model.IssueCategoryGeneralInquiry
	}

	// 7. 校验 resolution_type
	if !model.IsValidResolutionType(result.ResolutionType) {
		fmt.Printf("⚠️ 警告: conv_%s 非法 resolution_type '%s', 已修正为 'pending'\n",
			result.ConversationID, result.ResolutionType)
		result.ResolutionType = model.ResolutionTypePending
	}

	// 8. 校验 user_sentiment
	if !model.IsValidUserSentiment(result.UserSentiment) {
		fmt.Printf("⚠️ 警告: conv_%s 非法 user_sentiment '%s', 已修正为 'neutral'\n",
			result.ConversationID, result.UserSentiment)
		result.UserSentiment = model.UserSentimentNeutral
	}

	// 9. 校验 turn_count（优先信任代码计算）
	expectedTurnCount := len(conv.Turns)
	if result.TurnCount != expectedTurnCount {
		fmt.Printf("⚠️ 警告: conv_%s turn_count 不匹配 (LLM: %d, 实际: %d), 已修正\n",
			result.ConversationID, result.TurnCount, expectedTurnCount)
		result.TurnCount = expectedTurnCount
	}

	// 10. 校验 confidence 范围
	if result.Confidence < 0 {
		result.Confidence = 0
	}
	if result.Confidence > 1 {
		result.Confidence = 1
	}

	// 11. 检查是否包含转人工关键词
	transferKeywords := []string{"转人工", "人工客服", "转接"}
	result.ContainsTransfer = containsAny(conv, transferKeywords)
}

// containsAny 检查对话是否包含关键词
func containsAny(conv model.Conversation, keywords []string) bool {
	for _, turn := range conv.Turns {
		content := strings.ToLower(turn.Content)
		for _, keyword := range keywords {
			if strings.Contains(content, strings.ToLower(keyword)) {
				return true
			}
		}
	}
	return false
}
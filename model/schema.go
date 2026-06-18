package model

import (
	"time"
)

// 枚举常量定义
const (
	// MainIntent 主意图
	MainIntentRefund     = "refund"
	MainIntentInquiry    = "inquiry"
	MainIntentComplaint  = "complaint"
	MainIntentTechnical  = "technical"
	MainIntentAccount    = "account"
	MainIntentBilling    = "billing"
	MainIntentOther      = "other"

	// IssueCategory 问题分类
	IssueCategoryProductQuality  = "product_quality"
	IssueCategoryLogistics       = "logistics"
	IssueCategoryAccountMgmt     = "account_mgmt"
	IssueCategoryTechnicalIssue  = "technical_issue"
	IssueCategoryBillingPayment  = "billing_payment"
	IssueCategoryComplaintSuggest = "complaint_suggest"
	IssueCategoryGeneralInquiry  = "general_inquiry"

	// ResolutionType 解决方式
	ResolutionTypeSelfService     = "self_service"
	ResolutionTypeAgentResolved   = "agent_resolved"
	ResolutionTypeEscalated       = "escalated"
	ResolutionTypeTransferredHuman = "transferred_human"
	ResolutionTypePending         = "pending"
	ResolutionTypeFollowUpRequired = "follow_up_required"

	// UserSentiment 用户情绪
	UserSentimentPositive = "positive"
	UserSentimentNeutral  = "neutral"
	UserSentimentNegative = "negative"
)

// 枚举校验映射
var (
	ValidMainIntents = map[string]bool{
		MainIntentRefund:    true,
		MainIntentInquiry:  true,
		MainIntentComplaint: true,
		MainIntentTechnical: true,
		MainIntentAccount:   true,
		MainIntentBilling:   true,
		MainIntentOther:     true,
	}

	ValidIssueCategories = map[string]bool{
		IssueCategoryProductQuality:  true,
		IssueCategoryLogistics:        true,
		IssueCategoryAccountMgmt:      true,
		IssueCategoryTechnicalIssue:   true,
		IssueCategoryBillingPayment:   true,
		IssueCategoryComplaintSuggest: true,
		IssueCategoryGeneralInquiry:   true,
	}

	ValidResolutionTypes = map[string]bool{
		ResolutionTypeSelfService:     true,
		ResolutionTypeAgentResolved:   true,
		ResolutionTypeEscalated:       true,
		ResolutionTypeTransferredHuman: true,
		ResolutionTypePending:         true,
		ResolutionTypeFollowUpRequired: true,
	}

	ValidUserSentiments = map[string]bool{
		UserSentimentPositive: true,
		UserSentimentNeutral:  true,
		UserSentimentNegative: true,
	}
)

// ExtractionResult 表示单条对话的提取结果
type ExtractionResult struct {
	// 标识字段
	ConversationID string `json:"conversation_id"`
	Channel        string `json:"channel"`
	AgentName      string `json:"agent_name"`

	// 核心字段
	UserQuery     string   `json:"user_query"`
	MainIntent    string   `json:"main_intent"`
	IssueCategory string   `json:"issue_category"`
	SubIssues     []string `json:"sub_issues,omitempty"`

	// 解决字段
	IsResolved       bool   `json:"is_resolved"`
	ResolutionType   string `json:"resolution_type"`
	ResolutionText   string `json:"resolution_text,omitempty"`
	RequiresFollowUp bool   `json:"requires_follow_up"`

	// 体验字段
	UserSentiment     string `json:"user_sentiment"`
	SentimentEvidence string `json:"sentiment_evidence,omitempty"`

	// 元信息字段
	TurnCount        int  `json:"turn_count"`
	TopicSwitches    int  `json:"topic_switches"`
	ContainsTransfer bool `json:"contains_transfer"`
	HasInfoMissing   bool `json:"has_info_missing"`

	// 系统字段
	ExtractedAt time.Time `json:"extracted_at"`
	Confidence  float64   `json:"confidence"`
}

// OutputMetadata 输出文件的元数据
type OutputMetadata struct {
	TaskName              string    `json:"task_name"`
	TotalCount            int       `json:"total_count"`
	SuccessCount          int       `json:"success_count"`
	ErrorCount            int       `json:"error_count"`
	ModelUsed             string    `json:"model_used"`
	ExtractedAt           time.Time `json:"extracted_at"`
	ProcessingTimeSeconds float64   `json:"processing_time_seconds"`
	TotalTokensUsed       int       `json:"total_tokens_used"`
}

// OutputError 输出错误记录
type OutputError struct {
	ConversationID string `json:"conversation_id"`
	Error          string `json:"error"`
}

// OutputFile 输出文件结构
type OutputFile struct {
	Metadata OutputMetadata     `json:"metadata"`
	Errors   []OutputError      `json:"errors,omitempty"`
	Results  []ExtractionResult `json:"results"`
}

// IsValidMainIntent 检查主意图是否合法
func IsValidMainIntent(intent string) bool {
	return ValidMainIntents[intent]
}

// IsValidIssueCategory 检查问题分类是否合法
func IsValidIssueCategory(category string) bool {
	return ValidIssueCategories[category]
}

// IsValidResolutionType 检查解决方式是否合法
func IsValidResolutionType(resolutionType string) bool {
	return ValidResolutionTypes[resolutionType]
}

// IsValidUserSentiment 检查用户情绪是否合法
func IsValidUserSentiment(sentiment string) bool {
	return ValidUserSentiments[sentiment]
}
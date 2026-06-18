package extractor

import (
	"fmt"
	"strings"

	"cs_extractor/model"
)

// SystemPrompt 系统提示词
const SystemPrompt = `你是一名专业的客服对话数据分析助手，擅长从客服对话中提取结构化信息。你的任务是分析客服对话内容，提取用户诉求、意图、情绪、解决情况等关键信息。

## 输出格式要求

请严格按照以下 JSON Schema 输出结果：

{
  "conversation_id": "string - 对话唯一标识",
  "channel": "string - 渠道（如：在线、电话）",
  "agent_name": "string - 客服姓名",
  "user_query": "string - 用户核心诉求，不超过50字",
  "main_intent": "string - 主意图，必须是以下值之一：refund, inquiry, complaint, technical, account, billing, other",
  "issue_category": "string - 问题分类，必须是以下值之一：product_quality, logistics, account_mgmt, technical_issue, billing_payment, complaint_suggest, general_inquiry",
  "sub_issues": ["string - 子诉求列表，可选"],
  "is_resolved": "boolean - 是否已解决",
  "resolution_type": "string - 解决方式，必须是以下值之一：self_service, agent_resolved, escalated, transferred_human, pending, follow_up_required",
  "resolution_text": "string - 解决详情，可选",
  "requires_follow_up": "boolean - 是否需要跟进",
  "user_sentiment": "string - 用户情绪，必须是以下值之一：positive, neutral, negative",
  "sentiment_evidence": "string - 情绪判断依据，可选",
  "turn_count": "integer - 对话轮次总数",
  "topic_switches": "integer - 话题切换次数",
  "contains_transfer": "boolean - 是否包含转人工",
  "has_info_missing": "boolean - 是否存在信息缺失",
  "extracted_at": "string - 提取时间，ISO8601格式",
  "confidence": "number - AI置信度，范围0-1"
}

## 提取规则

1. **用户诉求提取**：
   - 提取用户最核心的问题或需求
   - 用简洁的语言概括，不超过50字
   - 如果有多个诉求，记录在 sub_issues 中

2. **意图识别**：
   - refund：退款、退货、换货相关
   - inquiry：咨询、查询、了解信息
   - complaint：投诉、不满、抱怨
   - technical：技术问题、APP使用问题
   - account：账号相关问题
   - billing：账单、支付、费用问题
   - other：其他无法归类的意图

3. **问题分类**：
   - product_quality：产品质量问题
   - logistics：物流、配送问题
   - account_mgmt：账号管理问题
   - technical_issue：技术故障问题
   - billing_payment：账单支付问题
   - complaint_suggest：投诉建议
   - general_inquiry：一般咨询

4. **解决情况判断**：
   - is_resolved：用户问题是否得到解决
   - resolution_type：根据解决方式选择对应枚举值
   - requires_follow_up：是否需要后续跟进处理

5. **情绪分析**：
   - positive：积极情绪（满意、感谢、开心）
   - neutral：中性情绪（平静、一般）
   - negative：消极情绪（不满、愤怒、失望）
   - sentiment_evidence：提供判断依据的具体语句

## 枚举值约束

**main_intent 可选值**：refund, inquiry, complaint, technical, account, billing, other

**issue_category 可选值**：product_quality, logistics, account_mgmt, technical_issue, billing_payment, complaint_suggest, general_inquiry

**resolution_type 可选值**：self_service, agent_resolved, escalated, transferred_human, pending, follow_up_required

**user_sentiment 可选值**：positive, neutral, negative

## 重要提醒

1. 必须输出合法的 JSON 格式，不要包含任何其他文字
2. 不要编造信息，如果某些字段无法确定，使用合理的默认值
3. turn_count 必须准确反映对话轮次数量
4. confidence 反映你对提取结果的确定性程度
5. extracted_at 使用当前时间的 ISO8601 格式`

// BuildUserPrompt 构建用户提示词
func BuildUserPrompt(conv model.Conversation) string {
	var sb strings.Builder

	sb.WriteString("请分析以下客服对话并提取结构化信息：\n\n")
	sb.WriteString(fmt.Sprintf("【对话ID】%s\n", conv.ID))
	sb.WriteString(fmt.Sprintf("【渠道】%s\n", conv.Channel))
	sb.WriteString(fmt.Sprintf("【客服姓名】%s\n\n", conv.Agent))

	// 处理对话内容，支持超长对话截断
	turns := conv.Turns
	if len(turns) > 20 {
		turns = truncateTurns(turns)
	}

	sb.WriteString("【对话内容】\n")
	for i, turn := range turns {
		role := "👤用户"
		if turn.Role == "agent" {
			role = "💁客服"
		}
		sb.WriteString(fmt.Sprintf("第%d轮 %s: %s\n", i+1, role, turn.Content))
	}

	sb.WriteString("\n请按规定的 JSON 格式输出提取结果：")

	return sb.String()
}

// truncateTurns 截断超长对话
// 策略：保留前5轮 + 中间采样5轮 + 最后10轮
func truncateTurns(turns []model.Turn) []model.Turn {
	if len(turns) <= 20 {
		return turns
	}

	var result []model.Turn

	// 前5轮
	result = append(result, turns[:5]...)

	// 中间采样5轮
	middleStart := 5
	middleEnd := len(turns) - 10
	if middleEnd > middleStart {
		step := (middleEnd - middleStart) / 5
		if step < 1 {
			step = 1
		}
		for i := 0; i < 5 && middleStart+i*step < middleEnd; i++ {
			result = append(result, turns[middleStart+i*step])
		}
	}

	// 最后10轮
	result = append(result, turns[len(turns)-10:]...)

	return result
}
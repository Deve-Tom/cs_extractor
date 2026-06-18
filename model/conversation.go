package model

// Turn 表示对话中的单轮消息
type Turn struct {
	Role    string `json:"role"`    // 角色：user 或 agent
	Content string `json:"content"` // 消息内容
}

// Conversation 表示一条完整的客服对话
type Conversation struct {
	ID      string `json:"id"`      // 对话唯一标识
	Channel string `json:"channel"` // 渠道（如"在线"、"电话"）
	Agent   string `json:"agent"`   // 客服姓名
	Turns   []Turn `json:"turns"`   // 对话轮次列表
}
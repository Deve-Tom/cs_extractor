# 客服对话结构化信息提取工具

基于 Golang + 火山方舟豆包大模型 Chat API 开发的自动化工具，从客服对话中提取结构化信息（用户诉求、意图、情绪、解决情况等），输出标准 JSON。

## 技术栈

- Go 1.21+
- 火山方舟豆包模型
- 火山方舟 Chat API

## 项目结构

```
cs_extractor/
├── main.go                    # CLI 入口
├── go.mod
├── config.yaml                # 配置文件
├── model/
│   ├── conversation.go        # 输入数据结构
│   └── schema.go              # 输出 Schema 定义
├── extractor/
│   ├── prompt.go              # Prompt 模板构建
│   ├── doubao_client.go       # 豆包 API 客户端
│   ├── mock.go                # Mock 模式客户端
│   └── pipeline.go            # 批量处理流水线
├── validator/
│   └ validate.go              # 结果校验
├── data/
│   ├── input/
│   │   └ conversation.json    # 输入对话文件
│   └ output/
│       └ results.json         # 输出结果文件
```

## 输入数据格式

`conversation.json` 文件格式：

```json
[
  {
    "id": "conv_001",
    "channel": "在线",
    "agent": "小王",
    "turns": [
      {"role": "user", "content": "你好，我想退款"},
      {"role": "agent", "content": "您好，请问订单号是多少？"}
    ]
  }
]
```

## 输出数据格式

`results.json` 文件格式：

```json
{
  "metadata": {
    "task_name": "cs-dialogue-extraction",
    "total_count": 5,
    "success_count": 5,
    "error_count": 0,
    "model_used": "ep-xxxxxxxx",
    "extracted_at": "2024-01-08T15:30:00Z",
    "processing_time_seconds": 12.5,
    "total_tokens_used": 5000
  },
  "results": [
    {
      "conversation_id": "conv_001",
      "channel": "在线",
      "agent_name": "小王",
      "user_query": "用户申请退款",
      "main_intent": "refund",
      "issue_category": "product_quality",
      "is_resolved": true,
      "resolution_type": "agent_resolved",
      "requires_follow_up": false,
      "user_sentiment": "positive",
      "turn_count": 9,
      "topic_switches": 1,
      "contains_transfer": false,
      "has_info_missing": false,
      "extracted_at": "2024-01-08T15:30:00Z",
      "confidence": 0.95
    }
  ]
}
```

## 提取字段说明

| 字段 | 类型 | 说明 |
|------|------|------|
| conversation_id | string | 对话唯一标识 |
| channel | string | 渠道（在线/电话） |
| agent_name | string | 客服姓名 |
| user_query | string | 用户核心诉求（≤50字） |
| main_intent | string | 主意图（枚举） |
| issue_category | string | 问题分类（枚举） |
| is_resolved | bool | 是否已解决 |
| resolution_type | string | 解决方式（枚举） |
| requires_follow_up | bool | 是否需要跟进 |
| user_sentiment | string | 用户情绪（枚举） |
| turn_count | int | 对话轮次总数 |
| topic_switches | int | 话题切换次数 |
| contains_transfer | bool | 是否转人工 |
| confidence | float | AI 置信度 (0~1) |

## 枚举值定义

**main_intent**: `refund`, `inquiry`, `complaint`, `technical`, `account`, `billing`, `other`

**issue_category**: `product_quality`, `logistics`, `account_mgmt`, `technical_issue`, `billing_payment`, `complaint_suggest`, `general_inquiry`

**resolution_type**: `self_service`, `agent_resolved`, `escalated`, `transferred_human`, `pending`, `follow_up_required`

**user_sentiment**: `positive`, `neutral`, `negative`

## 快速开始

### 1. 安装依赖

```bash
go mod download
```

### 2. 配置 API 信息

编辑 `config.yaml` 文件，填入豆包 API 的配置：

```yaml
doubao:
  api_key: "your-api-key"              # 豆包 API Key
  base_url: "https://ark.cn-beijing.volces.com/api/v3"
  model: "ep-xxxxxxxx"                 # 豆包模型 Endpoint ID
  temperature: 0.1
  top_p: 0.7
  max_tokens: 2000
  max_retries: 3

pipeline:
  concurrency: 3
  input_path: "data/input/conversation.json"
  output_path: "data/output/results.json"
```

### 3. 准备输入数据

将 `conversation.json` 文件放入 `data/input/` 目录。

### 4. 运行程序

**Mock 模式（测试）**：

```bash
go build -o cs-extractor .
./cs-extractor --mode=mock
```

**LLM 模式（真实 API）**：

```bash
./cs-extractor --mode=llm
```

### 5. 命令行参数

| 参数 | 默认值 | 说明 |
|------|--------|------|
| --mode | mock | 运行模式：mock / llm |
| --config | config.yaml | 配置文件路径 |
| --input | 配置文件中的值 | 输入文件路径 |
| --output | 配置文件中的值 | 输出文件路径 |
| --concurrency | 配置文件中的值 | 并发数 |

## AI 工具使用说明

### 豆包 API response_format 高级用法

本项目利用火山方舟豆包 API 的 `response_format` 参数强制规范 JSON 输出，提升结果稳定性：

```json
{
  "response_format": {
    "type": "json_schema",
    "json_schema": {
      "name": "cs_extraction_result",
      "strict": true,
      "schema": {
        "type": "object",
        "properties": { ... }
      }
    }
  }
}
```

**两种模式**：

| 模式 | type 值 | 稳定性 |
|------|---------|--------|
| JSON Object | json_object | ⭐⭐⭐ |
| JSON Schema | json_schema | ⭐⭐⭐⭐⭐ |

**开发建议**：
1. 优先使用 `json_schema` 模式，传入完整的 Schema 定义
2. 如遇到 400 错误（部分模型不支持），自动降级为 `json_object`
3. System Prompt 中仍保留完整字段说明作为双重保障

## 特性

- 支持任意数量对话处理（动态计算总数）
- 并发控制，避免 API 限流
- 自动重试机制（指数退避）
- 结果校验与自动修正
- 超长对话截断策略（>20轮时保留首尾+中间采样）
- Mock 模式便于测试开发

---
## 附带说明

## Schema 定义

### 1. MainIntent（主意图）
**定义**：用户发起本次对话的**最核心目的**（单选，提取最主要的意图）。

| 枚举值 | 中文名称 | 详细说明与适用场景 | 典型用户话术示例 |
|--------|---------|-------------------|-----------------|
| `refund` | 退款/退货 | 用户明确要求退回款项、取消订单并退款、或申请售后退换。 | “我要退款”、“怎么申请退货？”、“钱什么时候退给我？” |
| `inquiry` | 咨询/询问 | 用户询问产品详情、活动规则、政策、营业时间、使用说明等**非故障类**信息。 | “你们这个套餐包含什么？”、“双十一活动怎么参加？” |
| `complaint` | 投诉/抱怨 | 用户对服务态度、产品质量、业务流程等表达**强烈不满**，要求给出说法或赔偿。 | “你们这什么态度？”、“太差劲了，我要投诉！”、“必须给我个交代！” |
| `technical` | 技术支持 | 用户遇到软件、硬件、网络等**技术故障**，需要排查、修复或操作指导。 | “APP一直闪退”、“路由器连不上网”、“怎么重置密码？” |
| `account` | 账户管理 | 涉及账户层面的基础操作，如登录异常、账户解锁、信息修改、注销等。 | “我的号被封了”、“怎么换绑手机号？”、“帮我注销账号。” |
| `billing` | 账单/计费 | 涉及资金扣费异常、发票开具、账单查询、资费争议等**财务相关**问题。 | “为什么多扣了我50块钱？”、“怎么申请开发票？” |
| `other` | 其他 | 无法归类到以上类别的意图（如纯闲聊、无意义测试、打招呼等）。 | “在吗？”、“测试一下”、“谢谢”。 |

---

### 2. IssueCategory（问题分类）
**定义**：用户问题所属的**具体业务领域或模块**（单选，侧重于问题发生的客观领域）。

| 枚举值 | 中文名称 | 详细说明与适用场景 | 与 MainIntent 的区别 |
|--------|---------|-------------------|---------------------|
| `product_quality` | 产品质量 | 商品破损、瑕疵、与描述不符、过期、缺斤少两等实体或虚拟产品本身的问题。 | Intent 是“退款”，Category 是“产品质量” |
| `logistics` | 物流/配送 | 快递丢件、延迟发货、派送错误、物流信息停滞、包装破损等履约环节问题。 | Intent 是“投诉”，Category 是“物流” |
| `account_mgmt` | 账户管理 | 登录失败、实名认证、权限问题、资料修改、账号安全等。 | Intent 是“账户”，Category 是“账户管理” |
| `technical_issue` | 技术故障 | APP闪退、网页打不开、设备无法连接、系统报错、白屏等IT系统问题。 | Intent 是“技术”，Category 是“技术故障” |
| `billing_payment` | 账单与支付 | 支付失败、重复扣款、退款未到账、发票问题、优惠券无法使用等。 | Intent 是“账单”，Category 是“账单与支付” |
| `complaint_suggest` | 投诉与建议 | 对客服态度不满、对业务流程提出改进建议、对规则不合理进行反馈。 | Intent 是“投诉”，Category 是“投诉与建议” |
| `general_inquiry` | 综合咨询 | 售前咨询、使用说明、政策询问、门店地址查询等非特定故障类问题。 | Intent 是“咨询”，Category 是“综合咨询” |

---

### 3. ResolutionType（解决方式）
**定义**：本次对话结束时，客服对该问题的**最终处理状态和方式**（单选）。

| 枚举值 | 中文名称 | 详细说明与适用场景 | 判定依据（对话特征） |
|--------|---------|-------------------|---------------------|
| `self_service` | 引导自助 | 客服未直接操作，而是引导用户通过APP、官网、自助菜单、FAQ自行解决。 | 客服发送了链接、操作路径，用户回复“好的我试试”、“弄好了”。 |
| `agent_resolved` | 客服解决 | 客服在当前对话中**直接操作**或**准确解答**，彻底解决了用户的问题。 | 客服回复“已为您办理”、“已退款”，用户确认“收到了”、“解决了”。 |
| `escalated` | 升级处理 | 问题超出当前客服权限，已提交工单或升级给高级专员/二线支持/管理层。 | 客服回复“已为您提交工单”、“会由高级专员在24小时内联系您”。 |
| `transferred_human`| 转接人工 | 从机器人/智能客服成功转接至人工客服，或转接给其他特定部门的专家。 | 对话中出现“正在为您转接人工”、“转接至技术部”。 |
| `pending` | 待处理/中断 | 对话结束时问题**仍未解决**，且没有明确的后续跟进计划（如用户突然离开/失联）。 | 客服还在提问，但用户长时间未回复，对话自然结束。 |
| `follow_up_required`| 需后续跟进 | 问题已记录，需要其他部门核实或等待系统处理，**后续会再次联系用户**。 | 客服回复“财务核实后3个工作日内退款”、“技术排查后短信通知您”。 |

---

### 4. UserSentiment（用户情绪）
**定义**：用户在**整个对话过程中**表现出的整体情绪倾向（单选，需综合判断）。

| 枚举值 | 中文名称 | 详细说明与判定标准 | 典型特征/话术示例 |
|--------|---------|-------------------|------------------|
| `positive` | 积极/正面 | 用户表现出满意、感谢、赞赏、理解或愉悦的情绪。 | “太感谢了”、“效率真高”、“辛苦了”、“好的没问题”。 |
| `neutral` | 中立/平静 | 用户情绪平稳，仅客观陈述问题或进行正常的信息交互，无明显情感色彩。 | “我的订单号是123”、“怎么修改密码？”、“好的，知道了”。 |
| `negative` | 消极/负面 | 用户表现出愤怒、焦虑、失望、不耐烦、讽刺或强烈不满的情绪。 | “你们搞什么鬼？”、“太慢了！”、“等了半个小时了”、“垃圾软件”。 |

> **💡 情绪判定特殊规则（需写入 Prompt）**：
> 1. **动态变化**：如果用户一开始愤怒（Negative），但问题解决后表达感谢（Positive），整体情绪应判定为 `neutral` 或 `positive`（以最终情绪和整体基调为准）。
> 2. **证据支撑**：当判定为 `negative` 或 `positive` 时，必须在 `sentiment_evidence` 字段中引用用户的原话作为证据。

## Schema 设计思路

Schema 的设计核心在于 **“高内聚、低耦合、防幻觉”**。将数据结构分为输入（Input）和输出（Output）两部分，输出 Schema 采用 **“六维分组法”**，确保提取的信息既全面又具备业务分析价值。

输出结构包含 19 个字段，分为 6 个维度，设计考量如下：

| 维度 | 包含字段 | 设计考量与业务价值 |
| :--- | :--- | :--- |
| **标识维度** | `conversation_id`, `channel`, `agent_name` | **数据血缘**：确保提取结果能与原始数据 100% 关联，支持后续的客服绩效考核和按渠道分析。 |
| **核心诉求** | `user_query`, `main_intent`, `issue_category`, `sub_issues` | **意图与领域分离**：`main_intent`（如退款）代表用户的**主观目的**，`issue_category`（如物流）代表问题的**客观领域**。这种二维分类能支持更细粒度的交叉分析（如：物流领域的退款诉求）。`sub_issues` 采用数组，兼容“一次咨询多件事”的复杂场景。 |
| **解决情况** | `is_resolved`, `resolution_type`, `resolution_text`, `requires_follow_up` | **服务闭环评估**：不仅判断“是否解决”，还通过 `resolution_type` 记录“如何解决的”（如自助、客服解决、升级）。`requires_follow_up` 独立出来，方便生成每日“待跟进工单”列表。 |
| **体验情绪** | `user_sentiment`, `sentiment_evidence` | **防幻觉设计**：大模型在判断情绪时容易产生主观偏差。强制要求输出 `sentiment_evidence`（情绪证据/原话引用），利用“思维链（CoT）”原理，让模型先找证据再下结论，大幅提升情绪判断的准确率。 |
| **对话元信息**| `turn_count`, `topic_switches`, `contains_transfer`, `has_info_missing` | **对话质量特征**：`topic_switches`（话题切换次数）可用于识别用户是否“东拉西扯”或“问题复杂”；`contains_transfer` 和 `has_info_missing` 是评估客服业务熟练度和流程合理性的重要指标。 |
| **系统审计** | `extracted_at`, `confidence` | **工程容错**：`confidence`（置信度 0~1）让大模型进行“自我评估”。在工程上，可以设置阈值（如 <0.8），将低置信度的数据自动路由到人工复核队列。 |

---

## 任务拆解方式

项目采用 **“分层架构 + 接口隔离 + 渐进式交付”** 的原则进行任务拆解，确保开发过程可控、易测试。

### 1. 分层架构拆解
将系统拆分为 4 个核心层，每层职责单一：

*   **数据模型层 (`model/`)**：
    *   **职责**：定义输入输出的数据结构、枚举常量、基础校验函数。
    *   **原则**：不包含任何业务逻辑和 I/O 操作，纯数据定义。
*   **核心提取层 (`extractor/`)**：
    *   **职责**：Prompt 构建、LLM 通信、并发调度。
    *   **拆解**：
        *   `prompt.go`：专职负责字符串拼接和超长文本截断。
        *   `doubao_client.go`：专职负责 HTTP 请求、鉴权、响应解析。
        *   `pipeline.go`：专职负责读取文件、并发控制（Worker Pool）、进度统计。
*   **校验修复层 (`validator/`)**：
    *   **职责**：作为 LLM 输出后的“守门员”，进行数据清洗、枚举值纠正、缺失字段兜底。
*   **接入控制层 (`main.go` & `config.yaml`)**：
    *   **职责**：CLI 参数解析、环境变量读取、组件组装与启动。

### 2. 接口隔离设计 (依赖倒置)
在 `extractor` 包中定义统一的 `LLMClient` 接口：
```go
type LLMClient interface {
    Extract(ctx context.Context, systemPrompt, userPrompt string) (*model.ExtractionResult, error)
}
```
*   **拆解优势**：通过接口隔离，开发初期可以注入 `MockClient`（返回硬编码 JSON），跑通整个 Pipeline 和 Validator 逻辑；开发后期无缝切换为 `DoubaoClient`，极大降低联调成本。

### 3. 渐进式开发路径 (Milestones)
*   **M1 (骨架)**：完成 Model 定义和 MockClient，Pipeline 能跑通假数据并输出 JSON。
*   **M2 (大脑)**：完成 Prompt 工程和 DoubaoClient，单条真实 API 调用成功。
*   **M3 (肌肉)**：完成并发控制、重试机制、Validator 校验，跑通全量数据。
*   **M4 (皮肤)**：完成 CLI 参数、日志美化、README 编写。

---

## 边界处理策略

大模型应用（LLM App）的稳定性取决于对边界情况（Corner Cases）的处理。以下是针对本项目的核心边界处理策略矩阵：

### 1. 输入数据边界 (Input Boundaries)

| 边界场景 | 触发条件 | 处理策略 | 代码实现位置 |
| :--- | :--- | :--- | :--- |
| **空对话/极短对话** | `turns` 数组为空，或只有 1 轮且内容极短（如“在吗”）。 | **短路返回**：不调用 LLM，直接由代码生成默认结果（`main_intent=other`, `has_info_missing=true`, `confidence=0.1`），节省 API 费用。 | `pipeline.go` (预处理阶段) |
| **超长对话 (Token 溢出)**| `turns` 轮次 > 30 轮，或字符数超过模型 Context Window 的 80%。 | **智能截断**：保留前 5 轮（了解起因）+ 最后 10 轮（了解结果）+ 中间随机采样 5 轮（了解过程）。并在 User Prompt 中追加提示：*“注：对话过长，已进行采样截断。”* | `prompt.go` (`BuildUserPrompt`) |
| **非法角色 (Role)** | `turns` 中出现非 `user`/`agent` 的角色（如 `system` 或拼写错误）。 | **角色清洗**：在预处理阶段，将未知角色统一映射为 `user`，或直接过滤掉，防止 LLM API 报 400 错误。 | `pipeline.go` (预处理阶段) |

### 2. LLM 输出边界 (Output Boundaries)

| 边界场景 | 触发条件 | 处理策略 | 代码实现位置 |
| :--- | :--- | :--- | :--- |
| **Markdown 包裹** | LLM 返回 ````json \n {...} \n ```` 格式。 | **正则清洗**：使用正则表达式 `(?s)```(?:json)?\s*(.*?)\s*``` ` 提取内部内容，或直接去除首尾的 Markdown 标记。 | `doubao_client.go` (响应解析) |
| **尾部冗余解释** | LLM 在 JSON 后附加了“*以上是提取结果...*”等废话。 | **边界截断**：从字符串中查找第一个 `{` 和最后一个 `}`，仅截取中间的子串进行 `json.Unmarshal`。 | `doubao_client.go` (响应解析) |
| **枚举值幻觉** | LLM 输出了不在定义范围内的枚举值（如将 `refund` 写成 `refunding`）。 | **后置纠正 (Fallback)**：在 `validator` 中维护合法枚举字典。若发现非法值，尝试进行编辑距离匹配（Levenshtein）纠正；若纠正失败，则降级为默认值（如 `other` / `neutral`），并记录 Warning。 | `validator/validate.go` |
| **关键字段缺失** | LLM 漏掉了 `required` 字段（如没输出 `conversation_id`）。 | **代码兜底**：`validator` 检查必填项，若缺失，使用输入数据中的原始值（如 `conv.ID`）进行强制覆盖补充。 | `validator/validate.go` |

### 3. 网络与工程边界 (Network & Engineering Boundaries)

| 边界场景 | 触发条件 | 处理策略 | 代码实现位置 |
| :--- | :--- | :--- | :--- |
| **API 限流 (429)** | 并发过高触发火山方舟 RPM/TPM 限制。 | **动态退避**：捕获 429 状态码，读取响应头 `Retry-After`；若无该头，则采用指数退避（1s -> 2s -> 4s -> 8s），最多重试 4 次。 | `doubao_client.go` (HTTP 拦截器) |
| **单条数据 Panic** | 某条数据格式极其诡异，导致 Goroutine 崩溃。 | **Recover 兜底**：在 Pipeline 的每个 Worker Goroutine 顶部添加 `defer func() { if r := recover(); r != nil { ... } }()`，捕获 panic，记录错误日志，确保主流程不中断。 | `pipeline.go` (Worker 函数) |
| **JSON 解析失败** | 经过清洗后，LLM 输出的依然是非法 JSON（如缺少引号）。 | **重试机制**：将解析失败视为一次“网络错误”，触发重试机制。在重试的 User Prompt 中追加严厉警告：*“你上次输出的 JSON 格式错误，请务必严格检查语法！”* | `pipeline.go` (重试逻辑) |

### 4. 业务逻辑边界 (Business Logic Boundaries)

| 边界场景 | 触发条件 | 处理策略 |
| :--- | :--- | :--- |
| **情绪动态反转** | 用户开头大骂（Negative），结尾问题解决后说“谢谢，态度真好”（Positive）。 | **Prompt 约束**：在 System Prompt 中明确规定：*“如果用户情绪发生反转，请以**对话结束时**的最终情绪为准，并在 `sentiment_evidence` 中说明情绪变化过程。”* |
| **多意图冲突** | 用户既要求退款，又投诉态度，还询问新产品。 | **主次分离**：Prompt 规定：*“提取最核心、占用篇幅最大的诉求作为 `main_intent` 和 `user_query`，其余次要诉求放入 `sub_issues` 数组中。”* |


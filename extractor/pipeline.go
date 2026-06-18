package extractor

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"cs_extractor/model"
	"cs_extractor/validator"
)

// Pipeline 批量处理流水线
type Pipeline struct {
	client      LLMClient
	concurrency int
	modelUsed   string
}

// NewPipeline 创建流水线
func NewPipeline(client LLMClient, concurrency int, modelUsed string) *Pipeline {
	return &Pipeline{
		client:      client,
		concurrency: concurrency,
		modelUsed:   modelUsed,
	}
}

// Run 执行批量处理
func (p *Pipeline) Run(ctx context.Context, inputPath, outputPath string) error {
	startTime := time.Now()

	// 1. 加载输入数据
	data, err := os.ReadFile(inputPath)
	if err != nil {
		return fmt.Errorf("读取输入文件失败: %w", err)
	}

	var conversations []model.Conversation
	if err := json.Unmarshal(data, &conversations); err != nil {
		return fmt.Errorf("解析输入数据失败: %w", err)
	}

	total := len(conversations)
	fmt.Printf("📂 成功加载 %d 条对话数据\n", total)

	// 处理空数据情况
	if total == 0 {
		output := model.OutputFile{
			Metadata: model.OutputMetadata{
				TaskName:              "cs-dialogue-extraction",
				TotalCount:            0,
				SuccessCount:          0,
				ErrorCount:            0,
				ModelUsed:             p.modelUsed,
				ExtractedAt:           time.Now(),
				ProcessingTimeSeconds: 0,
				TotalTokensUsed:       0,
			},
			Errors:  []model.OutputError{},
			Results: []model.ExtractionResult{},
		}
		return p.writeOutput(outputPath, &output)
	}

	// 2. 并发处理
	results := make([]*model.ExtractionResult, total)
	errors := make([]model.OutputError, 0)
	var mu sync.Mutex
	var wg sync.WaitGroup

	// 使用 semaphore 控制并发
	semaphore := make(chan struct{}, p.concurrency)

	for i, conv := range conversations {
		wg.Add(1)
		go func(idx int, conversation model.Conversation) {
			defer wg.Done()

			// 获取信号量
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			fmt.Printf("🔄 [%d/%d] 处理 %s...\n", idx+1, total, conversation.ID)

			result, err := p.client.Extract(ctx, conversation)
			if err != nil {
				mu.Lock()
				errors = append(errors, model.OutputError{
					ConversationID: conversation.ID,
					Error:          err.Error(),
				})
				mu.Unlock()
				fmt.Printf("❌ [%d/%d] 失败 %s: %s\n", idx+1, total, conversation.ID, err.Error())
				return
			}

			// 3. 校验结果
			validator.Validate(result, conversation)

			mu.Lock()
			results[idx] = result
			mu.Unlock()

			fmt.Printf("✅ [%d/%d] 完成 %s\n", idx+1, total, conversation.ID)
		}(i, conv)
	}

	wg.Wait()

	// 4. 统计结果
	successCount := 0
	validResults := make([]model.ExtractionResult, 0)
	for _, r := range results {
		if r != nil {
			successCount++
			validResults = append(validResults, *r)
		}
	}

	processingTime := time.Since(startTime).Seconds()

	// 5. 构建输出
	output := model.OutputFile{
		Metadata: model.OutputMetadata{
			TaskName:              "cs-dialogue-extraction",
			TotalCount:            total,
			SuccessCount:          successCount,
			ErrorCount:            len(errors),
			ModelUsed:             p.modelUsed,
			ExtractedAt:           time.Now(),
			ProcessingTimeSeconds: processingTime,
			TotalTokensUsed:       0, // 可以从客户端累计
		},
		Errors:  errors,
		Results: validResults,
	}

	// 6. 写入输出文件
	if err := p.writeOutput(outputPath, &output); err != nil {
		return err
	}

	fmt.Printf("🎉 处理完成！成功 %d 条，失败 %d 条，耗时 %.2f 秒\n",
		successCount, len(errors), processingTime)

	return nil
}

// writeOutput 写入输出文件
func (p *Pipeline) writeOutput(outputPath string, output *model.OutputFile) error {
	// 确保输出目录存在
	dir := outputPath[:len(outputPath)-len("/results.json")]
	if dir != "" {
		os.MkdirAll(dir, 0755)
	}

	data, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化输出失败: %w", err)
	}

	if err := os.WriteFile(outputPath, data, 0644); err != nil {
		return fmt.Errorf("写入输出文件失败: %w", err)
	}

	return nil
}
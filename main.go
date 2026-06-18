package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"cs_extractor/extractor"

	"gopkg.in/yaml.v3"
)

// Config 配置结构
type Config struct {
	Doubao struct {
		APIKey      string  `yaml:"api_key"`
		BaseURL     string  `yaml:"base_url"`
		Model       string  `yaml:"model"`
		Temperature float64 `yaml:"temperature"`
		TopP        float64 `yaml:"top_p"`
		MaxTokens   int     `yaml:"max_tokens"`
		MaxRetries  int     `yaml:"max_retries"`
	} `yaml:"doubao"`
	Pipeline struct {
		Concurrency int    `yaml:"concurrency"`
		InputPath   string `yaml:"input_path"`
		OutputPath  string `yaml:"output_path"`
	} `yaml:"pipeline"`
}

// loadConfig 加载配置文件
func loadConfig(configPath string) (*Config, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %w", err)
	}

	return &config, nil
}

func main() {
	// 命令行参数
	mode := flag.String("mode", "mock", "运行模式: mock 或 llm")
	input := flag.String("input", "", "输入文件路径（默认使用配置文件中的值）")
	output := flag.String("output", "", "输出文件路径（默认使用配置文件中的值）")
	concurrency := flag.Int("concurrency", 0, "并发数（默认使用配置文件中的值）")
	configPath := flag.String("config", "config.yaml", "配置文件路径")
	flag.Parse()

	// 加载配置文件
	config, err := loadConfig(*configPath)
	if err != nil {
		fmt.Printf("❌ 加载配置文件失败: %s\n", err.Error())
		os.Exit(1)
	}

	// 命令行参数优先于配置文件
	if *input == "" {
		*input = config.Pipeline.InputPath
	}
	if *output == "" {
		*output = config.Pipeline.OutputPath
	}
	if *concurrency == 0 {
		*concurrency = config.Pipeline.Concurrency
	}

	// 创建客户端
	var client extractor.LLMClient
	var modelUsed string

	if *mode == "llm" {
		// 从配置文件获取 API Key 和 Model
		apiKey := config.Doubao.APIKey
		modelEndpoint := config.Doubao.Model

		if apiKey == "" || modelEndpoint == "" {
			fmt.Println("❌ 错误: 请在 config.yaml 中配置 doubao.api_key 和 doubao.model")
			os.Exit(1)
		}

		clientConfig := extractor.DoubaoConfig{
			APIKey:      apiKey,
			BaseURL:     config.Doubao.BaseURL,
			Model:       modelEndpoint,
			Temperature: config.Doubao.Temperature,
			TopP:        config.Doubao.TopP,
			MaxTokens:   config.Doubao.MaxTokens,
			MaxRetries:  config.Doubao.MaxRetries,
		}
		client = extractor.NewDoubaoClient(clientConfig)
		modelUsed = modelEndpoint
		fmt.Printf("🚀 使用豆包 API 模式 (模型: %s)\n", modelEndpoint)
	} else {
		client = extractor.NewMockClient()
		modelUsed = "mock-client"
		fmt.Println("🧪 使用 Mock 模式")
	}

	// 创建流水线
	pipeline := extractor.NewPipeline(client, *concurrency, modelUsed)

	// 执行处理
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	if err := pipeline.Run(ctx, *input, *output); err != nil {
		fmt.Printf("❌ 处理失败: %s\n", err.Error())
		os.Exit(1)
	}

	fmt.Printf("📁 结果已保存到: %s\n", *output)
}

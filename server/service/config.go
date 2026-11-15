package service

import (
	"errors"
	"fmt"
	"strings"

	"server/core/db"
	"server/model"

	"github.com/1szt/go-pkg/env"
)

// ------------------------ 系统配置（sys_config 表，key-value） ------------------------
// 读取顺序：数据库（管理页保存的值）→ data/.env（旧配置兼容）→ 内置默认值。

// 配置键
const (
	CfgAiBaseURL         = "ai_base_url"         // 大模型接口地址（OpenAI 兼容，不带 /chat/completions）
	CfgAiAPIKey          = "ai_api_key"          // 大模型 API Key
	CfgAiModel           = "ai_model"            // 模型名，如 deepseek-chat
	CfgAiSystemPrompt    = "ai_system_prompt"    // AI 商品匹配的系统提示词
	CfgAiThinking        = "ai_thinking"         // 思考模式开关："1" 开 / "0" 关
	CfgAiReasoningEffort = "ai_reasoning_effort" // 思考强度：low / medium / high
)

// DefaultAiSystemPrompt AI 商品匹配的内置系统提示词（数据库未配置时使用）
const DefaultAiSystemPrompt = "你是购物匹配助手。根据用户描述从商品列表中模糊匹配挑选商品并给出数量。" +
	"只输出一个 JSON 数组，元素格式 {\"id\":\"商品id\",\"qty\":购买数量}；" +
	"id 必须来自商品列表，qty 为 1-200 的整数；没有匹配的商品时输出 []。" +
	"不要输出任何其他文字、解释或代码块标记。"

// DefaultAiBaseURL 内置接口地址（兼容旧 data/.env 配置）
const DefaultAiBaseURL = "https://api.deepseek.com/v1"

// getCfg 读取配置：数据库优先 → env 兜底 → 内置默认
func getCfg(key, envKey, def string) string {
	if db.DB != nil {
		var c model.SysConfig
		if err := db.DB.First(&c, "key = ?", key).Error; err == nil && strings.TrimSpace(c.Value) != "" {
			return strings.TrimSpace(c.Value)
		}
	}
	if envKey != "" {
		if v := strings.TrimSpace(env.Get(envKey)); v != "" {
			return v
		}
	}
	return def
}

// setCfg 写入配置（upsert）
func setCfg(key, value, remark string) error {
	return db.DB.Where("key = ?", key).Assign(map[string]any{
		"value": value, "remark": remark,
	}).FirstOrCreate(&model.SysConfig{Key: key}).Error
}

// ------------------------ AI 设置（管理页） ------------------------

type GetAiSettingsInput struct {
}
type GetAiSettingsOutput struct {
	BaseURL      string `json:"base_url"`
	APIKey       string `json:"api_key"`
	Model        string `json:"model"`
	SystemPrompt string `json:"system_prompt"`
	// Thinking 思考模式开关：请求以 thinking.type = enabled/disabled 控制深度思考
	Thinking bool `json:"thinking"`
	// ReasoningEffort 思考强度：low / medium / high
	ReasoningEffort string `json:"reasoning_effort"`
	// UserPromptTemplate 真实拼接的 user 消息模板（商品列表取当前上架商品，
	// 与 MatchGoods 实际发送一致；点单文字用占位符），供管理页「拼接后的提示词」展示
	UserPromptTemplate string `json:"user_prompt_template"`
}

// GetAiSettings 读取 AI 配置（DB 优先，env/内置默认兜底），供管理页展示编辑
func GetAiSettings(i GetAiSettingsInput) (o GetAiSettingsOutput, err error) {
	o.BaseURL = getCfg(CfgAiBaseURL, "LFDD_AI_BASE_URL", DefaultAiBaseURL)
	o.APIKey = getCfg(CfgAiAPIKey, "LFDD_AI_API_KEY", "")
	o.Model = getCfg(CfgAiModel, "LFDD_AI_MODEL", "deepseek-chat")
	o.SystemPrompt = getCfg(CfgAiSystemPrompt, "", DefaultAiSystemPrompt)
	o.Thinking = getCfg(CfgAiThinking, "", "0") == "1"
	o.ReasoningEffort = getCfg(CfgAiReasoningEffort, "", "medium")

	// 拼真实 user 消息模板：与 MatchGoods 的 fmt.Sprintf 结构完全一致
	list, _, err := buildMatchGoodsList()
	if err != nil {
		return o, err
	}
	if list == "" {
		o.UserPromptTemplate = "（暂无上架商品）"
	} else {
		o.UserPromptTemplate = fmt.Sprintf("商品列表：\n%s\n用户点单：<用户输入的文字>", list)
	}
	return o, nil
}

type SaveAiSettingsInput struct {
	BaseURL      string `json:"base_url"`
	APIKey       string `json:"api_key"`
	Model        string `json:"model"`
	SystemPrompt string `json:"system_prompt"`
	// Thinking 思考模式开关
	Thinking bool `json:"thinking"`
	// ReasoningEffort 思考强度：low / medium / high
	ReasoningEffort string `json:"reasoning_effort"`
}
type SaveAiSettingsOutput struct {
}

// SaveAiSettings 保存 AI 配置（写入 sys_config，立即生效）
func SaveAiSettings(i SaveAiSettingsInput) (o SaveAiSettingsOutput, err error) {
	if db.DB == nil {
		return o, errors.New("数据库未连接")
	}
	i.BaseURL = strings.TrimRight(strings.TrimSpace(i.BaseURL), "/")
	i.Model = strings.TrimSpace(i.Model)
	if i.BaseURL == "" || i.Model == "" {
		return o, errors.New("接口地址和模型名不能为空")
	}
	if i.APIKey = strings.TrimSpace(i.APIKey); i.APIKey == "" {
		return o, errors.New("API Key 不能为空")
	}
	if strings.TrimSpace(i.SystemPrompt) == "" {
		return o, errors.New("系统提示词不能为空")
	}
	effort := strings.ToLower(strings.TrimSpace(i.ReasoningEffort))
	if effort == "" {
		effort = "medium"
	}
	if effort != "low" && effort != "medium" && effort != "high" {
		return o, errors.New("思考强度只支持 low / medium / high")
	}

	if err = setCfg(CfgAiBaseURL, i.BaseURL, "大模型接口地址"); err != nil {
		return o, err
	}
	if err = setCfg(CfgAiAPIKey, i.APIKey, "大模型 API Key"); err != nil {
		return o, err
	}
	if err = setCfg(CfgAiModel, i.Model, "模型名"); err != nil {
		return o, err
	}
	if err = setCfg(CfgAiSystemPrompt, i.SystemPrompt, "AI 商品匹配系统提示词"); err != nil {
		return o, err
	}
	thinking := "0"
	if i.Thinking {
		thinking = "1"
	}
	if err = setCfg(CfgAiThinking, thinking, "思考模式开关"); err != nil {
		return o, err
	}
	if err = setCfg(CfgAiReasoningEffort, effort, "思考强度"); err != nil {
		return o, err
	}
	return o, nil
}

type TestAiInput struct {
	BaseURL string `json:"base_url"` // 空 = 用已保存配置
	APIKey  string `json:"api_key"`
	Model   string `json:"model"`
}
type TestAiOutput struct {
	Reply string `json:"reply"` // 模型原样回复
}

// TestAi 测试大模型连通性：发一条最简消息，能拿到回复即成功。
// 入参为空的字段回退到已保存配置（方便「先填再测」）。
func TestAi(i TestAiInput) (o TestAiOutput, err error) {
	baseURL := strings.TrimSpace(i.BaseURL)
	if baseURL == "" {
		baseURL = getCfg(CfgAiBaseURL, "LFDD_AI_BASE_URL", DefaultAiBaseURL)
	}
	modelName := strings.TrimSpace(i.Model)
	if modelName == "" {
		modelName = getCfg(CfgAiModel, "LFDD_AI_MODEL", "deepseek-chat")
	}
	apiKey := strings.TrimSpace(i.APIKey)
	if apiKey == "" {
		apiKey = getCfg(CfgAiAPIKey, "LFDD_AI_API_KEY", "")
	}
	if apiKey == "" {
		return o, errors.New("未配置 API Key")
	}

	out, err := chatCompletionExt(ChatCompletionExtInput{
		BaseURL: baseURL,
		Model:   modelName,
		APIKey:  apiKey,
		System:  "你是连通性测试助手。",
		User:    "收到请只回复：ok",
	})
	if err != nil {
		return o, err
	}
	o.Reply = strings.TrimSpace(out.Content)
	if o.Reply == "" {
		o.Reply = "已连通（模型返回空内容）"
	}
	return o, nil
}

package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"server/core/db"
	"server/model"
	"strings"
	"time"
	"unicode/utf8"
)

// ------------------------ MatchGoods AI 商品匹配 ------------------------
// 前端传自然语言点单文字 → 后端取库中商品（id+名称）拼提示词 → 调大模型 →
// 解析返回的 JSON 数组 [{"id":"1","qty":2}] → 校验后回传命中的商品完整信息，
// 前端据此把商品加入购物车。

const matchTextMaxRunes = 1000 // 点单文字长度上限（字符）
const matchMaxQty = 500        // 单商品匹配数量上限

type MatchGoodsInput struct {
	Text string `json:"text"` // 用户自然语言点单描述，如「两杯柠檬水」
}
type MatchedGoods struct {
	GoodsView
	Qty int `json:"qty"` // 匹配到的购买数量
}
type MatchGoodsOutput struct {
	List []MatchedGoods `json:"list"` // 命中商品（空数组 = 没匹配到）
}

// matchPair 大模型返回的单个匹配项（id 可能是字符串或数字，用 RawMessage 兼容两种）
type matchPair struct {
	ID  json.RawMessage `json:"id"`
	Qty int             `json:"qty"`
}

// fenceRe 剥掉模型可能包裹的 ```json 代码块
var fenceRe = regexp.MustCompile("(?s)```(?:json)?\\s*(.*?)```")

func MatchGoods(i MatchGoodsInput) (o MatchGoodsOutput, err error) {
	o.List = make([]MatchedGoods, 0)

	text := strings.TrimSpace(i.Text)
	if text == "" {
		return o, errors.New("请输入点单内容")
	}
	if utf8.RuneCountInString(text) > matchTextMaxRunes {
		return o, fmt.Errorf("点单内容不能超过 %d 个字", matchTextMaxRunes)
	}
	apiKey := getCfg(CfgAiAPIKey, "")
	if apiKey == "" {
		return o, errors.New("未配置大模型：请到管理页「AI 设置」填写 API Key")
	}

	// 1. 取上架商品列表作为提示词素材
	list, byID, err := buildMatchGoodsList()
	if err != nil {
		return o, err
	}
	if len(byID) == 0 {
		return o, errors.New("暂无上架商品")
	}

	// 2. 调大模型（OpenAI 兼容 /chat/completions）：系统提示词 DB 优先，未配置用内置默认
	system := getCfg(CfgAiSystemPrompt, DefaultAiSystemPrompt)
	user := fmt.Sprintf("商品列表：\n%s\n用户点单：%s", list, text)

	completion, err := chatCompletion(ChatCompletionInput{
		APIKey: apiKey, System: system, User: user,
	})
	if err != nil {
		return o, err
	}

	// 3. 解析并校验匹配结果（只回传库里真实存在的上架商品）
	parsed, err := parseMatchJSON(ParseMatchJSONInput{Content: completion.Content})
	if err != nil {
		return o, errors.New("大模型返回格式异常，请重试")
	}
	pairs := parsed.Pairs

	merged := map[int64]int{} // id → 数量（合并重复项）
	order := []int64{}
	for _, p := range pairs {
		idOut, e := parseMatchID(ParseMatchIDInput{Raw: p.ID})
		if e != nil || p.Qty <= 0 {
			continue
		}
		id := idOut.ID
		g, ok := byID[id]
		if !ok {
			continue // 模型编造的 id 直接丢弃
		}
		if _, seen := merged[id]; !seen {
			order = append(order, id)
		}
		merged[id] += p.Qty
		if merged[id] > matchMaxQty {
			merged[id] = matchMaxQty // 数量封顶 200
		}
		if g.Stock > 0 && merged[id] > g.Stock {
			merged[id] = g.Stock // 再封顶到库存
		}
	}

	for _, id := range order {
		g := byID[id]
		o.List = append(o.List, MatchedGoods{
			GoodsView: GoodsView{
				ID: g.ID, MenuID: g.MenuID, Name: g.Name, Alias: g.Alias, Price: g.Price,
				Spec: g.Spec, Stock: g.Stock, Image: g.Image, Sort: g.Sort, Status: g.Status,
			},
			Qty: merged[id],
		})
	}
	return o, nil
}

// buildMatchGoodsList 拼接上架商品列表文本（提示词 user 消息的固定素材）。
// 返回逐行 "- id=xxx 名称=xxx 规格=xxx\n" 的文本 + id→商品映射；无上架商品时文本为空、映射为空。
// MatchGoods 实际发送与管理页「拼接后的提示词展示」共用此函数，保证所见即所得。
func buildMatchGoodsList() (list string, byID map[int64]model.Goods, err error) {
	var goods []model.Goods
	if err = db.DB.Where("status = ?", 1).Order("sort ASC, id DESC").Find(&goods).Error; err != nil {
		return "", nil, err
	}
	byID = make(map[int64]model.Goods, len(goods))
	var sb strings.Builder
	for _, g := range goods {
		line := fmt.Sprintf("- id=%d 名称=%s", g.ID, g.Name)
		if g.Alias != "" {
			line += " 别名=" + g.Alias
		}
		if g.Spec != "" {
			line += " 规格=" + g.Spec
		}
		sb.WriteString(line)
		sb.WriteByte('\n')
		byID[g.ID] = g
	}
	return sb.String(), byID, nil
}

type ChatCompletionInput struct {
	APIKey string `json:"api_key"` // 大模型 API Key
	System string `json:"system"`  // system 提示词
	User   string `json:"user"`    // user 消息内容
}
type ChatCompletionOutput struct {
	Content string `json:"content"` // 首个 choice 的 content
}

// chatCompletion 调 OpenAI 兼容的 chat/completions（地址/模型走配置），返回首个 choice 的 content
func chatCompletion(i ChatCompletionInput) (o ChatCompletionOutput, err error) {
	ext, err := chatCompletionExt(ChatCompletionExtInput{
		BaseURL: getCfg(CfgAiBaseURL, DefaultAiBaseURL),
		Model:   getCfg(CfgAiModel, "deepseek-chat"),
		APIKey:  i.APIKey,
		System:  i.System,
		User:    i.User,
	})
	o.Content = ext.Content
	return o, err
}

type ChatCompletionExtInput struct {
	BaseURL string `json:"base_url"` // 接口地址（不带 /chat/completions）
	Model   string `json:"model"`
	APIKey  string `json:"api_key"`
	System  string `json:"system"`
	User    string `json:"user"`
}
type ChatCompletionExtOutput struct {
	Content string `json:"content"`
}

// chatCompletionExt 按显式地址/模型调 chat/completions（测试连接用已填未存的值）
func chatCompletionExt(i ChatCompletionExtInput) (o ChatCompletionExtOutput, err error) {
	baseURL := i.BaseURL
	if baseURL == "" {
		baseURL = DefaultAiBaseURL
	}
	model := i.Model
	if model == "" {
		model = "deepseek-chat"
	}

	reqBody := map[string]any{
		"model":       model,
		"temperature": 0,
		"messages": []map[string]string{
			{"role": "system", "content": i.System},
			{"role": "user", "content": i.User},
		},
	}
	// 思考模式开关：thinking.type = enabled/disabled（OpenAI 兼容端点通用写法，
	// 如 DeepSeek/豆包/通义等。不支持思考的模型/端点会忽略该字段）
	thinkingType := "disabled"
	if getCfg(CfgAiThinking, "0") == "1" {
		thinkingType = "enabled"
		reqBody["reasoning_effort"] = getCfg(CfgAiReasoningEffort, "medium")
	}
	reqBody["thinking"] = map[string]string{"type": thinkingType}
	body, _ := json.Marshal(reqBody)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		strings.TrimRight(baseURL, "/")+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return o, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+i.APIKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return o, errors.New("大模型请求失败，请检查网络")
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20)) // 上限 1MB
	if err != nil {
		return o, err
	}

	var out struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
		Error *struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	if err = json.Unmarshal(data, &out); err != nil {
		return o, errors.New("大模型响应解析失败")
	}
	if out.Error != nil {
		return o, fmt.Errorf("大模型错误: %s", out.Error.Message)
	}
	if resp.StatusCode != http.StatusOK {
		return o, fmt.Errorf("大模型请求失败（%d）", resp.StatusCode)
	}
	if len(out.Choices) == 0 {
		return o, errors.New("大模型未返回内容")
	}
	o.Content = out.Choices[0].Message.Content
	return o, nil
}

type ParseMatchJSONInput struct {
	Content string `json:"content"` // 大模型返回的原始文本
}
type ParseMatchJSONOutput struct {
	Pairs []matchPair `json:"pairs"` // 解析出的匹配项数组
}

// parseMatchJSON 解析模型输出为匹配项数组（容忍代码块包裹/前后杂散文字）
func parseMatchJSON(i ParseMatchJSONInput) (o ParseMatchJSONOutput, err error) {
	s := strings.TrimSpace(i.Content)
	if m := fenceRe.FindStringSubmatch(s); m != nil {
		s = strings.TrimSpace(m[1])
	}
	// 截取首个 [ ... ] 区间，规避模型在数组前后加文字
	if start := strings.IndexByte(s, '['); start >= 0 {
		if end := strings.LastIndexByte(s, ']'); end > start {
			s = s[start : end+1]
		}
	}
	var pairs []matchPair
	if err = json.Unmarshal([]byte(s), &pairs); err != nil {
		return o, err
	}
	o.Pairs = pairs
	return o, nil
}

type ParseMatchIDInput struct {
	Raw json.RawMessage `json:"raw"` // 模型返回的 id（可能是字符串或数字）
}
type ParseMatchIDOutput struct {
	ID int64 `json:"id,string"` // 解析出的商品 id
}

// parseMatchID 兼容模型把 id 输出成字符串 "1" 或数字 1
func parseMatchID(i ParseMatchIDInput) (o ParseMatchIDOutput, err error) {
	var s string
	if e := json.Unmarshal(i.Raw, &s); e == nil {
		_, e = fmt.Sscanf(s, "%d", &o.ID)
		return o, e
	}
	err = json.Unmarshal(i.Raw, &o.ID)
	return o, err
}

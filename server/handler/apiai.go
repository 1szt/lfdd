package handler

import (
	"server/service"

	"github.com/gin-gonic/gin"
)

type ApiAiHandler struct {
}

// Settings 读取 AI 配置（接口地址 / API Key / 模型 / 系统提示词）
func (h *ApiAiHandler) Settings(c *gin.Context) {
	out, err := service.GetAiSettings(service.GetAiSettingsInput{})
	if err != nil {
		c.JSON(500, gin.H{
			"message": "服务器内部错误",
		})
		return
	}

	c.JSON(200, out)
}

// Save 保存 AI 配置（写入 sys_config，立即生效）
func (h *ApiAiHandler) Save(c *gin.Context) {
	var req service.SaveAiSettingsInput
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{
			"message": "请求参数错误",
		})
		return
	}

	if _, err := service.SaveAiSettings(req); err != nil {
		c.JSON(400, gin.H{
			"message": err.Error(),
		})
		return
	}

	c.JSON(200, gin.H{
		"message": "ok",
	})
}

// Test 测试大模型连通性（可传已填未存的值；空字段回退已保存配置）
// body: {"base_url":"...","api_key":"...","model":"..."} → {"reply":"ok"}
func (h *ApiAiHandler) Test(c *gin.Context) {
	var req service.TestAiInput
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{
			"message": "请求参数错误",
		})
		return
	}

	out, err := service.TestAi(req)
	if err != nil {
		c.JSON(400, gin.H{
			"message": err.Error(),
		})
		return
	}

	c.JSON(200, out)
}

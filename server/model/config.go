package model

import "time"

// SysConfig 系统配置表：key-value 存储（AI 模型 API、系统提示词等可在线修改的配置）
type SysConfig struct {
	Key   string `json:"key" gorm:"primaryKey;size:64"`
	Value string `json:"value" gorm:"size:4096"`
	// 备注说明（前端展示用，如「DeepSeek API Key」）
	Remark string `json:"remark" gorm:"size:128"`
	// 是否敏感值（如 api_key，列表接口脱敏返回）
	Sensitive bool      `json:"sensitive"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TableName SysConfig 表结构
func (SysConfig) TableName() string {
	return "sys_config"
}

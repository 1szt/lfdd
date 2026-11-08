// Package service 业务层：业务规则与数据编排，只依赖 model/db，不依赖 gin/HTTP。
// 被 handler 与 middleware 共用的逻辑放这里；仅单模块私有的逻辑留在模块内。
package service

import "errors"

type Auth struct {
	Username string
	Email    string
	Password string
	Token    string
	IP       string
	Device   string
}

// Register 注册新用户：查重 → 密码哈希 → 写库 → 创建会话
// 返回登录 token（自动登录）
func (a *Auth) Register() (string, error) {
	// 检查用户名是否存在
	if a.Username == "" {
		return "", errors.New("用户名不能为空")
	}

	// 检查邮箱是否存在
	if a.Email == "" {
		return "", errors.New("邮箱不能为空")
	}

	// 检查密码是否为空
	if a.Password == "" {
		return "", errors.New("密码不能为空")
	}

	// 密码哈希
	// 写库
	// 创建会话
	// 返回登录 token（自动登录）
	return "", nil
}


// passwd 密码哈希
// 返回密码哈希串
func (a *Auth) passwd() string {
	return ""
}

// CheckPassword 哈希对比
// 返回是否匹配
func (a *Auth) CheckPassword() bool {
	return false
}

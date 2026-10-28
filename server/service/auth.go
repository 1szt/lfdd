// Package service 业务层：业务规则与数据编排，只依赖 model/db，不依赖 gin/HTTP。
// 被 handler 与 middleware 共用的逻辑放这里；仅单模块私有的逻辑留在模块内。
package service

import (
	"errors"
	"hxzl/core/db"
	"hxzl/model"

	"golang.org/x/crypto/bcrypt"
)

// ErrDBNotReady 数据库未连接（未初始化安装时）
var ErrDBNotReady = errors.New("数据库未连接")

// ErrUserExists 用户名或邮箱已存在
var ErrUserExists = errors.New("用户名或邮箱已存在")

// Register 注册新用户：查重 → 密码哈希 → 写库
func Register(username, email, password string) (*model.User, error) {
	if db.DB == nil {
		return nil, ErrDBNotReady
	}

	// 查重（username / email 均有唯一索引）
	var count int64
	db.DB.Model(&model.User{}).
		Where("username = ? OR email = ?", username, email).
		Count(&count)
	if count > 0 {
		return nil, ErrUserExists
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &model.User{
		ID:       model.NewID(),
		Username: username,
		Email:    email,
		Password: string(hash),
	}
	if err := db.DB.Create(user).Error; err != nil {
		return nil, err
	}
	return user, nil
}

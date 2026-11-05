// Package service 业务层：业务规则与数据编排，只依赖 model/db，不依赖 gin/HTTP。
// 被 handler 与 middleware 共用的逻辑放这里；仅单模块私有的逻辑留在模块内。
package service

import (
	crand "crypto/rand"
	"encoding/base64"
	"errors"
	"hxzl/core/db"
	"hxzl/model"
	"math/rand/v2"
	"time"

	"gorm.io/gorm"
)

type Auth struct {
	Username string
	Email    string
	Password string
}

// Register 注册新用户：查重 → 密码哈希 → 写库
func (a *Auth) Register() (*model.User, error) {
	_, err := a.CheckUser()
	switch {
	case err == nil:
		return nil, errors.New("用户名或邮箱已存在")
	case !errors.Is(err, gorm.ErrRecordNotFound):
		return nil, err // 查询本身出错，不是"用户不存在"
	}

	hash, err := HashPassword(a.Password)
	if err != nil {
		return nil, err
	}

	user := &model.User{
		ID:       a.NewID(),
		Username: a.Username,
		Email:    a.Email,
		Password: hash,
	}
	if err := db.DB.Create(user).Error; err != nil {
		return nil, err
	}
	return user, nil
}

// 查询用户是否存在
func (a *Auth) CheckUser() (*model.User, error) {
	var user model.User
	if err := db.DB.Where("username = ? OR email = ?", a.Username, a.Email).
		First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// NewID 生成 18 位数字 ID：毫秒时间戳(13位) + 随机数(5位)，不依赖数据库自增。
// 多实例部署时如出现冲突，再升级为雪花算法。
func (Auth) NewID() int64 {
	return time.Now().UnixMilli()*100_000 + int64(rand.IntN(100_000))
}

// NewToken 生成 URL 安全 token：32 字节密码学随机数，base64url 编码为 43 字符。
func (Auth) NewToken() (string, error) {
	b := make([]byte, 32) // 256 bit
	if _, err := crand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

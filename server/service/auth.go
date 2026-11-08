// Package service 业务层：业务规则与数据编排，只依赖 model/db，不依赖 gin/HTTP。
// 被 handler 与 middleware 共用的逻辑放这里；仅单模块私有的逻辑留在模块内。
package service

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	"hxzl/core/db"
	"hxzl/model"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// sessionTTL 会话（refresh token）有效期
const sessionTTL = 30 * 24 * time.Hour

// RegisterInput 注册入参（每个函数自定义自己的入参结构体，互不耦合）
type RegisterInput struct {
	Username string
	Email    string
	Password string
	IP       string
	Device   string
}

// Register 注册新用户：校验 → 查重 → 密码哈希 → 写库 → 创建会话
// 返回登录 token（自动登录）
func Register(in RegisterInput) (string, error) {
	if err := validateRegisterInput(&in); err != nil {
		return "", err
	}
	if db.DB == nil {
		return "", errors.New("数据库未连接")
	}

	// 查重：用户名、邮箱任一已存在即拒绝
	var count int64
	if err := db.DB.Model(&model.User{}).
		Where("username = ? OR email = ?", in.Username, in.Email).
		Count(&count).Error; err != nil {
		return "", fmt.Errorf("查询用户失败: %w", err)
	}
	if count > 0 {
		return "", errors.New("用户名或邮箱已存在")
	}

	// 密码哈希
	hash, err := HashPassword(in.Password)
	if err != nil {
		return "", fmt.Errorf("密码哈希失败: %w", err)
	}

	// 写库 + 创建会话（同一事务，任一步失败整体回滚）
	token, err := newToken()
	if err != nil {
		return "", fmt.Errorf("生成会话 token 失败: %w", err)
	}
	id, err := newID()
	if err != nil {
		return "", fmt.Errorf("生成用户 ID 失败: %w", err)
	}
	user := &model.User{
		ID:       id,
		Username: in.Username,
		Email:    in.Email,
		Password: hash,
	}
	session := &model.Session{
		Token:     token,
		UserID:    user.ID,
		IP:        in.IP,
		Device:    in.Device,
		ExpiresAt: time.Now().Add(sessionTTL),
	}
	if err := db.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(user).Error; err != nil {
			return err
		}
		return tx.Create(session).Error
	}); err != nil {
		return "", fmt.Errorf("注册失败: %w", err)
	}
	return token, nil
}

// validateRegisterInput 校验并规范化入参：去除首尾空白，邮箱统一转小写（与 model.User 约定一致）
func validateRegisterInput(in *RegisterInput) error {
	in.Username = strings.TrimSpace(in.Username)
	in.Email = strings.ToLower(strings.TrimSpace(in.Email))
	if in.Username == "" {
		return errors.New("用户名不能为空")
	}
	if in.Email == "" {
		return errors.New("邮箱不能为空")
	}
	if in.Password == "" {
		return errors.New("密码不能为空")
	}
	return nil
}

// HashPassword 密码哈希（bcrypt）
// 返回密码哈希串
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// VerifyPassword 哈希对比：明文 password 与哈希串 hash 比对
// 返回是否匹配；不匹配返回 (false, nil)，比对出错返回错误
func VerifyPassword(password, hash string) (bool, error) {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err == nil {
		return true, nil
	}
	if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
		return false, nil
	}
	return false, err
}

// newToken 生成 32 字节随机串的十六进制编码（64 字符），用作 session token
func newToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// newID 生成内部主键：毫秒时间戳 × 1000 + 3 位随机数（见 model.User.ID 注释）
func newID() (int64, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(1000))
	if err != nil {
		return 0, err
	}
	return time.Now().UnixMilli()*1000 + n.Int64(), nil
}

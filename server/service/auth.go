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

// APIError 业务错误：携带 HTTP 状态码，handler 通过 errors.As 提取后映射响应
type APIError struct {
	Code    int
	Message string
}

func (e *APIError) Error() string { return e.Message }

func badRequest(msg string) error   { return &APIError{Code: 400, Message: msg} }
func unauthorized(msg string) error { return &APIError{Code: 401, Message: msg} }
func conflict(msg string) error     { return &APIError{Code: 409, Message: msg} }

// AuthInput 登录/注册共用入参（每个字段都可能为空，由各函数自行校验）
type AuthInput struct {
	Username string
	Email    string
	Password string
	IP       string
	Device   string
}

// Login 登录：账号（用户名或邮箱）+ 密码 → 校验 → 创建会话
// 返回会话 token；账号或密码错误一律返回"用户名或密码错误"，不泄露账号是否存在
func Login(in AuthInput) (string, error) {
	in.Username = strings.TrimSpace(in.Username)
	in.Email = strings.ToLower(strings.TrimSpace(in.Email))
	if in.Username == "" && in.Email == "" {
		return "", badRequest("请输入用户名或邮箱")
	}
	if in.Password == "" {
		return "", badRequest("密码不能为空")
	}
	if db.DB == nil {
		return "", errors.New("数据库未连接")
	}

	// 用户名、邮箱均可作为登录账号
	account := in.Username
	if account == "" {
		account = in.Email
	}
	var user model.User
	if err := db.DB.Where("username = ? OR email = ?", account, account).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", unauthorized("用户名或密码错误")
		}
		return "", fmt.Errorf("查询用户失败: %w", err)
	}

	// 先验密码，再报状态，避免通过状态码探测账号是否存在
	ok, err := VerifyPassword(in.Password, user.Password)
	if err != nil {
		return "", fmt.Errorf("密码校验失败: %w", err)
	}
	if !ok {
		return "", unauthorized("用户名或密码错误")
	}
	if user.Status != 1 {
		return "", unauthorized("账号已被禁用")
	}

	// 更新最近登录时间（失败不阻断登录）
	now := time.Now()
	if err := db.DB.Model(&user).Update("last_login_at", now).Error; err != nil {
		return "", fmt.Errorf("更新登录时间失败: %w", err)
	}

	return createSession(db.DB, user.ID, in.IP, in.Device)
}

// Register 注册新用户：校验 → 查重 → 密码哈希 → 写库 → 创建会话
// 返回登录 token（自动登录）
func Register(in AuthInput) (string, error) {
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
		return "", conflict("用户名或邮箱已存在")
	}

	// 密码哈希
	hash, err := HashPassword(in.Password)
	if err != nil {
		return "", fmt.Errorf("密码哈希失败: %w", err)
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

	// 写库 + 创建会话（同一事务，任一步失败整体回滚）
	var token string
	if err := db.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(user).Error; err != nil {
			return err
		}
		t, err := createSession(tx, user.ID, in.IP, in.Device)
		if err != nil {
			return err
		}
		token = t
		return nil
	}); err != nil {
		return "", fmt.Errorf("注册失败: %w", err)
	}
	return token, nil
}

// Logout 退出：按 token 删除会话
// 会话不存在也视为已退出（幂等），保证重复调用不报错
func Logout(token string) error {
	if token == "" {
		return badRequest("缺少 token")
	}
	if db.DB == nil {
		return errors.New("数据库未连接")
	}
	if err := db.DB.Where("token = ?", token).Delete(&model.Session{}).Error; err != nil {
		return fmt.Errorf("退出失败: %w", err)
	}
	return nil
}

// Refresh 刷新 token：校验旧会话有效 → 轮换为新 token（旧 token 作废）
// 返回新 token
func Refresh(token string) (string, error) {
	if token == "" {
		return "", badRequest("缺少 token")
	}
	if db.DB == nil {
		return "", errors.New("数据库未连接")
	}

	// 只认未过期的会话
	var session model.Session
	if err := db.DB.Where("token = ? AND expires_at > ?", token, time.Now()).
		First(&session).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", unauthorized("会话不存在或已过期")
		}
		return "", fmt.Errorf("查询会话失败: %w", err)
	}

	// 轮换 token 并续期
	newTok, err := newToken()
	if err != nil {
		return "", fmt.Errorf("生成会话 token 失败: %w", err)
	}
	if err := db.DB.Model(&model.Session{}).
		Where("token = ?", token).
		Updates(map[string]any{
			"token":      newTok,
			"expires_at": time.Now().Add(sessionTTL),
		}).Error; err != nil {
		return "", fmt.Errorf("刷新会话失败: %w", err)
	}
	return newTok, nil
}

// validateRegisterInput 校验并规范化入参：去除首尾空白，邮箱统一转小写（与 model.User 约定一致）
func validateRegisterInput(in *AuthInput) error {
	in.Username = strings.TrimSpace(in.Username)
	in.Email = strings.ToLower(strings.TrimSpace(in.Email))
	if in.Username == "" {
		return badRequest("用户名不能为空")
	}
	if in.Email == "" {
		return badRequest("邮箱不能为空")
	}
	if in.Password == "" {
		return badRequest("密码不能为空")
	}
	return nil
}

// createSession 创建会话并返回 token；tx 为事务句柄（注册时与写用户同一事务，登录时为 db.DB）
func createSession(tx *gorm.DB, userID int64, ip, device string) (string, error) {
	token, err := newToken()
	if err != nil {
		return "", fmt.Errorf("生成会话 token 失败: %w", err)
	}
	session := &model.Session{
		Token:     token,
		UserID:    userID,
		IP:        ip,
		Device:    device,
		ExpiresAt: time.Now().Add(sessionTTL),
	}
	if err := tx.Create(session).Error; err != nil {
		return "", fmt.Errorf("创建会话失败: %w", err)
	}
	return token, nil
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

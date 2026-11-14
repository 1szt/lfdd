package service

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	"strings"
	"time"

	"server/core/db"
	"server/model"

	"golang.org/x/crypto/argon2"
	"gorm.io/gorm"
)

// ================= 登录 =================
type LoginInput struct {
	Account  string `json:"account"`
	Password string `json:"password"`
	IP       string `json:"ip"`
	Device   string `json:"device"`
}
type LoginOutput struct {
	Token string `json:"token"`
}

// Login 登录：按用户名或邮箱查找 → 校验密码 → 更新登录时间 → 创建会话
func Login(i LoginInput) (o LoginOutput, err error) {
	account := strings.TrimSpace(i.Account)
	if account == "" || i.Password == "" {
		return o, errors.New("账号或密码不能为空")
	}

	// 1. 按用户名或邮箱查找（邮箱不区分大小写）
	var user model.User
	if err = db.DB.Where("username = ? OR email = ?", account, strings.ToLower(account)).
		First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return o, errors.New("账号或密码错误")
		}
		return o, err
	}

	// 2. 校验密码（与账号不存在返回相同提示，不泄露账号是否存在）
	v, err := VerifyPassword(VerifyPasswordInput{Password: i.Password, Hash: user.Password})
	if err != nil {
		return o, err
	}
	if !v.Valid {
		return o, errors.New("账号或密码错误")
	}

	// 3. 账号状态检查
	if user.Status == 0 {
		return o, errors.New("账号已禁用")
	}

	// 4. 更新最近登录时间
	now := time.Now()
	if err = db.DB.Model(&user).Update("last_login_at", now).Error; err != nil {
		return o, err
	}

	// 5. 创建会话，签发 token
	token, err := CreateSession(CreateSessionInput{
		UserID: user.ID,
		IP:     i.IP,
		Device: i.Device,
	})
	if err != nil {
		return o, err
	}

	o.Token = token.Token
	return o, nil
}

// ================= 注册 =================
type RegisterInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	IP       string `json:"ip"`
	Device   string `json:"device"`
}

type RegisterOutput struct {
	Token string `json:"token"`
}

// Register 注册：校验邮箱未占用 → 写入用户（argon2id 哈希）→ 创建会话
func Register(i RegisterInput) (o RegisterOutput, err error) {
	email := strings.ToLower(strings.TrimSpace(i.Email))
	if email == "" {
		return o, errors.New("邮箱不能为空")
	}
	if len(i.Password) < 8 {
		return o, errors.New("密码长度至少 8 位")
	}

	// 1. 检测是否存在相同邮箱的用户
	var count int64
	if err = db.DB.Model(&model.User{}).Where("email = ?", email).Count(&count).Error; err != nil {
		return o, err
	}
	if count > 0 {
		return o, errors.New("邮箱已存在")
	}

	// 2. 密码哈希（argon2id，存 PHC 编码串）
	h, err := HashPassword(HashPasswordInput{Password: i.Password})
	if err != nil {
		return o, err
	}

	// 2.5 首个注册用户自动成为管理员（role=2），后续用户默认普通（role=1）
	var userCount int64
	if err = db.DB.Model(&model.User{}).Count(&userCount).Error; err != nil {
		return o, err
	}
	role := int8(1)
	if userCount == 0 {
		role = 2
	}

	// 3. 生成业务主键与登录名
	id, err := GenUserID(GenUserIDInput{})
	if err != nil {
		return o, err
	}
	un, err := GenUsername(GenUsernameInput{Email: email})
	if err != nil {
		return o, err
	}

	// 4. 注册用户（ID、username 由业务层生成）
	user := &model.User{
		ID:       id.ID,
		Username: un.Username,
		Email:    email,
		Password: h.Hash,
		Role:     role,
	}
	if err = db.DB.Create(user).Error; err != nil {
		return o, err
	}

	// 5. 生成 token
	token, err := CreateSession(CreateSessionInput{
		UserID: user.ID,
		IP:     i.IP,
		Device: i.Device,
	})
	if err != nil {
		return o, err
	}

	o.Token = token.Token
	return o, nil
}

// ================= 退出登录 =================
type LogoutInput struct {
	Token string `json:"token"`
}
type LogoutOutput struct {
}

// Logout 退出登录（吊销当前 token）
func Logout(in LogoutInput) (o LogoutOutput, err error) {
	if _, err = DeleteSession(DeleteSessionInput{Token: in.Token}); err != nil {
		return o, err
	}
	return o, nil
}

// ================= 刷新token =================
type RefreshInput struct {
	Token string `json:"token"`
}
type RefreshOutput struct {
	Token string `json:"token"`
}

// Refresh 刷新token（校验旧 token 并换发新 token）
func Refresh(in RefreshInput) (o RefreshOutput, err error) {
	t, err := RefreshSession(RefreshSessionInput{Token: in.Token})
	if err != nil {
		return o, err
	}
	o.Token = t.Token
	return o, nil
}

// ================= 密码哈希（argon2id，PHC 编码） =================

const (
	argonTime    uint32 = 3
	argonMemory  uint32 = 64 * 1024 // 64 MiB，单位 KiB
	argonThreads uint8  = 4
	argonSaltLen        = 16
	argonKeyLen  uint32 = 32
)

type HashPasswordInput struct {
	Password string `json:"password"`
}
type HashPasswordOutput struct {
	Hash string `json:"hash"`
}

// HashPassword 生成 argon2id 密码哈希
// 格式：$argon2id$v=19$m=65536,t=3,p=4$<salt>$<hash>（均 RawStdEncoding）
func HashPassword(i HashPasswordInput) (o HashPasswordOutput, err error) {
	salt := make([]byte, argonSaltLen)
	if _, err = rand.Read(salt); err != nil {
		return o, err
	}
	hash := argon2.IDKey([]byte(i.Password), salt, argonTime, argonMemory, argonThreads, argonKeyLen)
	o.Hash = fmt.Sprintf("$argon2id$v=19$m=%d,t=%d,p=%d$%s$%s",
		argonMemory, argonTime, argonThreads,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(hash),
	)
	return o, nil
}

type VerifyPasswordInput struct {
	Password string `json:"password"`
	Hash     string `json:"hash"`
}
type VerifyPasswordOutput struct {
	Valid bool `json:"valid"`
}

// VerifyPassword 校验密码与 PHC 哈希是否匹配（常量时间比较）
func VerifyPassword(i VerifyPasswordInput) (o VerifyPasswordOutput, err error) {
	// 解析 PHC 串：["", "argon2id", "v=19", "m=...,t=...,p=...", salt, hash]
	parts := strings.Split(i.Hash, "$")
	if len(parts) != 6 || parts[1] != "argon2id" || parts[2] != "v=19" {
		return o, nil
	}
	var mem, cost uint32
	var par uint8
	if _, err = fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &mem, &cost, &par); err != nil {
		return o, nil
	}
	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return o, nil
	}
	expected, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return o, nil
	}
	actual := argon2.IDKey([]byte(i.Password), salt, cost, mem, par, uint32(len(expected)))
	o.Valid = subtle.ConstantTimeCompare(actual, expected) == 1
	return o, nil
}

// ================= 业务 ID / username 生成 =================

type GenUserIDInput struct {
}
type GenUserIDOutput struct {
	ID int64 `json:"id,string"`
}

// GenUserID 生成用户主键：毫秒时间戳（高位）+ 10 位随机数（低位）
func GenUserID(i GenUserIDInput) (o GenUserIDOutput, err error) {
	var b [2]byte
	if _, err = rand.Read(b[:]); err != nil {
		return o, err
	}
	o.ID = time.Now().UnixMilli()<<10 | int64(binary.BigEndian.Uint16(b[:])&0x3FF)
	return o, nil
}

type GenUsernameInput struct {
	Email string `json:"email"`
}
type GenUsernameOutput struct {
	Username string `json:"username"`
}

// GenUsername 生成登录名：邮箱前缀（限 20 字符）+ 4 位随机后缀，保证唯一
func GenUsername(i GenUsernameInput) (o GenUsernameOutput, err error) {
	prefix := strings.SplitN(i.Email, "@", 2)[0]
	var sb strings.Builder
	for _, r := range prefix {
		if sb.Len() >= 20 {
			break
		}
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '_' {
			sb.WriteRune(r)
		}
	}
	if sb.Len() == 0 {
		sb.WriteString("user")
	}
	var b [2]byte
	if _, err = rand.Read(b[:]); err != nil {
		return o, err
	}
	o.Username = fmt.Sprintf("%s_%04x", sb.String(), binary.BigEndian.Uint16(b[:]))
	return o, nil
}

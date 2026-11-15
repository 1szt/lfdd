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

// ================= 首次启动管理员 =================
type EnsureAdminInput struct{}
type EnsureAdminOutput struct {
	Created  bool   `json:"created"`  // 本次是否新建
	Email    string `json:"email"`    // 管理员邮箱（仅新建时回传，用于启动日志打印）
	Password string `json:"password"` // 管理员初始密码（仅新建时回传，库里只存哈希）
}

// EnsureAdmin 首次启动（用户表为空）时生成管理员账号：随机邮箱 + 随机密码。
// 注册功能已关闭，这是系统唯一的初始账号来源；凭据由调用方打印到启动日志。
func EnsureAdmin(_ EnsureAdminInput) (o EnsureAdminOutput, err error) {
	var count int64
	if err = db.DB.Model(&model.User{}).Count(&count).Error; err != nil {
		return o, err
	}
	if count > 0 {
		return o, nil
	}

	// 随机邮箱：admin_ + 8 位十六进制 @lfdd.local
	var b [4]byte
	if _, err = rand.Read(b[:]); err != nil {
		return o, err
	}
	email := fmt.Sprintf("admin_%x@lfdd.local", b)

	// 随机密码：12 位（字母表去掉易混淆字符）
	pwd, err := randString(12)
	if err != nil {
		return o, err
	}
	h, err := HashPassword(HashPasswordInput{Password: pwd})
	if err != nil {
		return o, err
	}
	id, err := GenUserID(GenUserIDInput{})
	if err != nil {
		return o, err
	}

	user := &model.User{
		ID:       id.ID,
		Username: strings.SplitN(email, "@", 2)[0], // 用户名与邮箱前缀相同，登录时两者都可
		Email:    email,
		Password: h.Hash,
		Role:     2, // 管理员
	}
	if err = db.DB.Create(user).Error; err != nil {
		return o, err
	}

	o.Created = true
	o.Email = email
	o.Password = pwd
	return o, nil
}

// randString 生成 n 位随机串（字母表去掉 0O1lI 等易混淆字符）
func randString(n int) (s string, err error) {
	const alphabet = "abcdefghjkmnpqrstuvwxyzABCDEFGHJKMNPQRSTUVWXYZ23456789"
	b := make([]byte, n)
	if _, err = rand.Read(b); err != nil {
		return "", err
	}
	for i := range b {
		b[i] = alphabet[int(b[i])%len(alphabet)]
	}
	return string(b), nil
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

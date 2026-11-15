package service

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"server/core/db"
	"server/model"
	"time"

	"gorm.io/gorm"
)

// ------------------------ CreateSession 创建会话（生成 Token） ------------------------
type CreateSessionInput struct {
	UserID int64  `json:"user_id,string"`
	IP     string `json:"ip"`
	Device string `json:"device"`
}
type CreateSessionOutput struct {
	Token string `json:"token"`
}

// CreateSession 生成一个新的 Token 并持久化保存到数据库
func CreateSession(i CreateSessionInput) (o CreateSessionOutput, err error) {
	// 1. 生成 32 字节（256位）的随机二进制数据
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return o, err
	}

	// 2. 转换为 16 进制字符串作为 Token
	tokenStr := hex.EncodeToString(bytes)

	// 3. 设置过期时间为 7 天
	expireAt := time.Now().Add(7 * 24 * time.Hour)

	// 4. 保存会话到数据库
	if err = db.DB.Create(&model.Session{
		Token:     tokenStr,
		UserID:    i.UserID,
		IP:        i.IP,
		Device:    i.Device,
		ExpiresAt: expireAt,
	}).Error; err != nil {
		return o, err
	}

	o.Token = tokenStr
	return o, nil
}

// ------------------------ RefreshSession 刷新会话（换发新 Token） ------------------------
type RefreshSessionInput struct {
	Token string `json:"token"`
}
type RefreshSessionOutput struct {
	Token string `json:"token"`
}

func RefreshSession(i RefreshSessionInput) (o RefreshSessionOutput, err error) {
	// 1. 查旧会话，须存在且未过期
	var s model.Session
	if err = db.DB.Where("token = ?", i.Token).First(&s).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return o, errors.New("token 无效或已过期")
		}
		return o, err
	}
	if time.Now().After(s.ExpiresAt) {
		return o, errors.New("token 无效或已过期")
	}

	// 2. 生成新 Token（32 字节随机数 → 16 进制）
	bytes := make([]byte, 32)
	if _, err = rand.Read(bytes); err != nil {
		return o, err
	}
	newToken := hex.EncodeToString(bytes)

	// 3. 替换旧 Token 并滑动续期 7 天（Token 是主键，直接 UPDATE 主键完成换发）
	if err = db.DB.Model(&model.Session{}).Where("token = ?", i.Token).Updates(map[string]interface{}{
		"token":      newToken,
		"expires_at": time.Now().Add(7 * 24 * time.Hour),
	}).Error; err != nil {
		return o, err
	}

	o.Token = newToken
	return o, nil
}

// ------------------------ DeleteSession 删除会话（登出/吊销） ------------------------
type DeleteSessionInput struct {
	Token string `json:"token"`
}
type DeleteSessionOutput struct {
	// 保持结构体配对规范，留空以备未来扩展
}

func DeleteSession(i DeleteSessionInput) (o DeleteSessionOutput, err error) {
	// 对应登出/踢出：从数据库中软删除或物理删除该 Session
	if err = db.DB.Where("token = ?", i.Token).Delete(&model.Session{}).Error; err != nil {
		return o, err
	}
	return o, nil
}

// ------------------------ GetSession 按 Token 查询会话 ------------------------

// SessionView 会话视图（查询结果输出；须查 model.Session 再映射，匿名结构体 GORM 推不出表名）
type SessionView struct {
	Token  string `json:"token"`
	UserID int64  `json:"user_id,string"`
	IP     string `json:"ip"`
	Device string `json:"device"`
}

type GetSessionInput struct {
	Token string `json:"token"`
}
type GetSessionOutput struct {
	Session SessionView `json:"session"`
}

func GetSession(i GetSessionInput) (o GetSessionOutput, err error) {
	// 只返回未过期的会话（过期即失效）
	var s model.Session
	if err = db.DB.Where("token = ? AND expires_at > ?", i.Token, time.Now()).First(&s).Error; err != nil {
		return o, err
	}
	o.Session = SessionView{Token: s.Token, UserID: s.UserID, IP: s.IP, Device: s.Device}
	return o, nil
}

// ------------------------ ListSessions 查询用户的会话列表 ------------------------
type ListSessionsInput struct {
	UserID int64 `json:"user_id,string"`
}
type ListSessionsOutput struct {
	Sessions []SessionView `json:"sessions"`
}

func ListSessions(i ListSessionsInput) (o ListSessionsOutput, err error) {
	// 只返回未过期的会话
	var sessions []model.Session
	if err = db.DB.Where("user_id = ? AND expires_at > ?", i.UserID, time.Now()).Find(&sessions).Error; err != nil {
		return o, err
	}
	o.Sessions = make([]SessionView, 0, len(sessions))
	for _, s := range sessions {
		o.Sessions = append(o.Sessions, SessionView{Token: s.Token, UserID: s.UserID, IP: s.IP, Device: s.Device})
	}
	return o, nil
}

// ------------------------ CleanExpiredSessions 扫描清理过期会话 ------------------------
type CleanExpiredSessionsInput struct {
}
type CleanExpiredSessionsOutput struct {
}

func CleanExpiredSessions(i CleanExpiredSessionsInput) (err error) {
	// 1. 查找所有过期的会话
	if err = db.DB.Where("expires_at < ?", time.Now()).Delete(&model.Session{}).Error; err != nil {
		return err
	}
	return nil
}
